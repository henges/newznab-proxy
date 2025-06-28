package proxy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/henges/newznab-proxy/newznab"
	"github.com/henges/newznab-proxy/proxy/model"
	"github.com/samber/lo"
)

type Proxy struct {
	c        *Config
	s        *Store
	backends []backend
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
	}, nil
}

func feedItemsToRssFeed(fis []model.FeedItem, host string, tls bool) *newznab.RssFeed {
	newzItems := lo.Map(fis, func(item model.FeedItem, index int) newznab.Item {
		return item.ToRewrittenNewznabItem(host, tls)
	})
	ret := newznab.NewRssFeedFromItems(0, 0, newzItems)
	return &ret
}

const requeryThreshold = time.Hour * 24

func (p *Proxy) Search(ctx context.Context, params newznab.SearchParams) (*newznab.RssFeed, error) {
	matches, err := p.s.SearchForFeedItem(ctx, params.Query)
	if err != nil {
		return nil, err
	}
	if len(matches) > 0 {
		return feedItemsToRssFeed(matches, p.c.Web.ExternalHost, p.c.Web.TLS), nil
	}

	params = params.WithSanitisedQuery()
	searchCache, err := p.s.LoadCurrentSearchCacheEntriesForQuery(ctx, params.Query, time.Now().Add(-requeryThreshold))
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	type result struct {
		skipped bool
		err     error
		vals    []model.FeedItem
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
			results[i] = result{vals: lo.Map(searchRes.Channel.Items, func(item newznab.Item, index int) model.FeedItem {
				return model.FeedItemFromNewznab(item, b.name, model.FeedItemSourceSearch)
			})}
		}()
	}
	wg.Wait()
	remoteMatches := make([]model.FeedItem, 0, 10)
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
			err = p.s.UpsertSearchCacheEntry(ctx, model.SearchCacheEntry{
				IndexerName:        b.name,
				Query:              params.Query,
				FirstTried:         time.Now(),
				LastTried:          time.Now(),
				SearchResultStatus: model.SearchResultStatusError,
				ErrorMessage:       res.err.Error(),
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		// If we got here then we either got a hit or a miss for this indexer
		status := model.SearchResultStatusHit
		if len(res.vals) == 0 {
			status = model.SearchResultStatusMiss
		}
		err = p.s.UpsertSearchCacheEntry(ctx, model.SearchCacheEntry{
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
	return feedItemsToRssFeed(remoteMatches, p.c.Web.ExternalHost, p.c.Web.TLS), nil
}
