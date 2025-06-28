package proxy

import (
	"context"
	"fmt"
	"sync"

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

func (p *Proxy) Search(ctx context.Context, params newznab.SearchParams) (*newznab.RssFeed, error) {
	matches, err := p.s.SearchForFeedItem(ctx, params.Query)
	if err != nil {
		return nil, err
	}
	if len(matches) > 0 {
		return feedItemsToRssFeed(matches, p.c.Web.ExternalHost, p.c.Web.TLS), nil
	}

	var wg sync.WaitGroup
	type result struct {
		err  error
		vals []model.FeedItem
	}
	results := make([]result, len(p.backends))
	for i, b := range p.backends {
		wg.Add(1)
		go func() {
			defer wg.Done()
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
		if res.err != nil {
			fmt.Printf("Failed to get results from %s because: %s", p.backends[i].name, err)
			continue
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
