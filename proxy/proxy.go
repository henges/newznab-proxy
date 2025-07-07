package proxy

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/henges/newznab-proxy/newznab"
	"github.com/samber/lo"
)

type Proxy struct {
	c        *Config
	s        *Store
	backends []backend

	pollerWg     *sync.WaitGroup
	pollerCancel func()
	done         chan struct{}
}

var _ newznab.ServerImplementation = (*Proxy)(nil)

type backend struct {
	name   string
	client *newznab.Client
	rssCfg *RSSConfig
}

func NewProxy(ctx context.Context, c *Config) (*Proxy, error) {
	db, err := NewStore(ctx, c.Storage.DBPath)
	if err != nil {
		return nil, err
	}
	backends := make([]backend, 0, len(c.Backends))
	for _, bcfg := range c.Backends {
		cl := newznab.NewClient(bcfg.BaseURL, bcfg.APIKey)
		backends = append(backends, backend{
			name:   bcfg.Name,
			client: cl,
			rssCfg: bcfg.RSS,
		})
	}
	return &Proxy{
		c:        c,
		s:        db,
		backends: backends,
		pollerWg: &sync.WaitGroup{},
	}, nil
}

func (p *Proxy) StartRSSPolls(ctx context.Context) {

	ctx, p.pollerCancel = context.WithCancel(ctx)
	p.done = make(chan struct{})
	for _, b := range p.backends {
		if b.rssCfg == nil {
			continue
		}
		for _, feed := range b.rssCfg.Feeds {
			params := maps.Clone(b.rssCfg.RSSQueryParams)
			maps.Copy(params, feed.QueryParams)
			p.pollerWg.Add(1)
			go func() {
				defer p.pollerWg.Done()
				poll := func() error {
					items, err := b.client.PollRSS(ctx, b.rssCfg.RSSPath, params)
					if err != nil {
						return err
					}
					feedItems := lo.Map(items.Channel.Items, func(item newznab.Item, index int) FeedItem {
						return FeedItemFromNewznab(item, b.name, FeedItemSourceRSS)
					})
					ids := lo.Map(feedItems, func(item FeedItem, index int) string {
						return item.UUID
					})
					existingIDs, err := p.s.GetFeedItemUUIDs(ctx, ids)
					if err != nil {
						return err
					}
					for _, fi := range feedItems {
						if _, ok := existingIDs[fi.UUID]; ok {
							continue
						}
						err = p.s.InsertFeedItem(ctx, fi)
						if err != nil {
							return err
						}
					}
					return nil
				}
				var err error
				err = poll()
				if err != nil {
					fmt.Printf("%s feed %s: error polling: %s", b.name, feed.Name, err)
				}

				for {
					select {
					case <-time.After(feed.PollInterval):
						{
							err = poll()
							if err != nil {
								fmt.Printf("%s feed %s: error polling: %s", b.name, feed.Name, err)
							}
						}
					case <-p.done:
						{
							return
						}
					case <-ctx.Done():
						{
							return
						}
					}
				}
			}()
		}
	}
}

func (p *Proxy) StopRSSPolls() error {
	if p.pollerCancel == nil {
		return nil
	}
	close(p.done)
	done := make(chan struct{}, 1)
	go func() {
		p.pollerWg.Wait()
		done <- struct{}{}
	}()
	for {
		select {
		case <-time.After(10 * time.Second):
			{
				p.pollerCancel()
				return errors.New("forced cancel")
			}
		case <-done:
			{
				p.pollerCancel()
				return nil
			}
		}
	}
}

func feedItemsToRssFeed(fis []FeedItem, host string, port uint16, tls bool) *newznab.RssFeed {
	newzItems := lo.Map(fis, func(item FeedItem, index int) newznab.Item {
		return item.ToRewrittenNewznabItem(host, port, tls)
	})
	ret := newznab.NewRssFeedFromItems(0, len(newzItems), newzItems)
	return &ret
}

const requeryThreshold = time.Hour * 24

func (p *Proxy) Search(ctx context.Context, params newznab.SearchParams) (*newznab.RssFeed, error) {
	matches, err := p.s.SearchForFeedItem(ctx, params.Query)
	if err != nil {
		return nil, err
	}
	if len(matches) > 0 {
		return feedItemsToRssFeed(matches, p.c.Web.ExternalHost, p.c.Web.Port, p.c.Web.TLS), nil
	}

	params = params.WithSanitisedQuery()
	searchCache, err := p.s.LoadSearchCacheEntriesForQuery(ctx, params.Query)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	type result struct {
		skipped bool
		err     error
		vals    []FeedItem
	}
	results := make([]result, len(p.backends))
	for i, b := range p.backends {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cacheEntry, ok := searchCache[b.name]
			if ok && cacheEntry.SearchResultStatus.ShouldNotRequery() {
				results[i] = result{skipped: true}
				return
			}
			searchRes, err := b.client.Search(ctx, params)
			if err != nil {
				results[i] = result{err: err}
				return
			}
			results[i] = result{vals: lo.Map(searchRes.Channel.Items, func(item newznab.Item, index int) FeedItem {
				return FeedItemFromNewznab(item, b.name, FeedItemSourceSearch)
			})}
		}()
	}
	wg.Wait()
	remoteMatches := make([]FeedItem, 0, 10)
	for i, res := range results {
		b := p.backends[i]
		if res.skipped {
			cacheEntry := searchCache[b.name]
			fmt.Printf("%s: skipped because search result status was %s, err message %s",
				p.backends[i].name, cacheEntry.SearchResultStatus, cacheEntry.ErrorMessage)
			continue
		}

		if res.err != nil {
			fmt.Printf("%s: Failed to get results because: %s", b.name, res.err)
			err = p.s.UpsertSearchCacheEntry(ctx, SearchCacheEntry{
				IndexerName:        b.name,
				Query:              params.Query,
				FirstTried:         time.Now(),
				LastTried:          time.Now(),
				SearchResultStatus: SearchResultStatusError,
				ErrorMessage:       res.err.Error(),
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		// If we got here then we either got a hit or a miss for this indexer
		status := SearchResultStatusHit
		if len(res.vals) == 0 {
			status = SearchResultStatusMiss
		}
		err = p.s.UpsertSearchCacheEntry(ctx, SearchCacheEntry{
			IndexerName:        b.name,
			Query:              params.Query,
			FirstTried:         time.Now(),
			LastTried:          time.Now(),
			SearchResultStatus: status,
			ErrorMessage:       "",
		})
		if err != nil {
			return nil, err
		}
		for _, fi := range res.vals {

			err = p.s.InsertFeedItem(ctx, fi)
			if err != nil {
				return nil, err
			}
			remoteMatches = append(remoteMatches, fi)
		}
	}
	return feedItemsToRssFeed(remoteMatches, p.c.Web.ExternalHost, p.c.Web.Port, p.c.Web.TLS), nil
}

func (p *Proxy) GetNZB(ctx context.Context, id string) (newznab.NZB, error) {

	var ret newznab.NZB
	nzbData, err := p.s.GetNZBDataByUUID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ret, newznab.ServerError{
				Code:        400,
				Description: "no NZB found with id " + id,
			}
		}
		return ret, err
	}
	var source *backend
	for _, b := range p.backends {
		if b.name == nzbData.IndexerName {
			source = &b
			break
		}
	}
	if source == nil {
		return ret, newznab.ServerError{
			Code:        400,
			Description: "the indexer that provided this NZB is no longer configured: " + nzbData.IndexerName,
		}
	}
	data, err := source.client.GetNZB(ctx, nzbData.URL)
	if err != nil {
		return ret, err
	}
	ret = newznab.NZB{
		Filename: fmt.Sprintf("%s.nzb", nzbData.Title),
		Data:     data,
	}
	return ret, nil
}
