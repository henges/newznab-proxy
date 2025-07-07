package proxy

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/henges/newznab-proxy/newznab"
)

type FeedItemSource string

const (
	FeedItemSourceRSS    FeedItemSource = "rss"
	FeedItemSourceSearch FeedItemSource = "search"
)

type FeedItem struct {
	UUID            string
	IndexerName     string
	Title           string
	GUID            string
	GUIDIsPermalink bool
	Link            string
	PubDate         time.Time
	NZBLink         string
	Size            int64
	Source          FeedItemSource
	Attrs           map[string]string
}

func FeedItemFromNewznab(i newznab.Item, indexer string, source FeedItemSource) FeedItem {
	concat := fmt.Sprintf("%s:%s", indexer, i.GUID.Value)
	sum := sha256.Sum256([]byte(concat))
	id := hex.EncodeToString(sum[:])

	return FeedItem{
		UUID:            id,
		IndexerName:     indexer,
		Title:           i.Title,
		GUID:            i.GUID.Value,
		GUIDIsPermalink: i.GUID.IsPermaLink,
		Link:            i.Link,
		PubDate:         time.Time(i.PubDate),
		NZBLink:         i.Enclosure.URL,
		Size:            i.Enclosure.Length,
		Source:          source,
		Attrs:           i.AttrsMap(),
	}
}

func (fi FeedItem) RewrittenNZBLink(host string, port uint16, tls bool) string {

	proto := "http"
	if tls {
		proto = "https"
	}
	if port != 0 && port != 80 {
		host = fmt.Sprintf("%s:%d", host, port)
	}
	return fmt.Sprintf("%s://%s/getnzb/%s", proto, host, fi.UUID)
}

func (fi FeedItem) ToRewrittenNewznabItem(host string, port uint16, tls bool) newznab.Item {

	ret := fi.ToNewznabItem()
	rewriteLink := fi.RewrittenNZBLink(host, port, tls)
	ret.Enclosure.URL = rewriteLink
	ret.Link = rewriteLink
	return ret
}

func (fi FeedItem) ToNewznabItem() newznab.Item {

	return newznab.Item{
		Title: fi.Title,
		GUID: newznab.RssGuid{
			IsPermaLink: fi.GUIDIsPermalink,
			Value:       fi.GUID,
		},
		Link:        fi.Link,
		Comments:    "",
		PubDate:     newznab.RFC1123Time(fi.PubDate),
		Category:    "",
		Description: "",
		Enclosure: newznab.RssEnclosure{
			URL:    fi.NZBLink,
			Length: fi.Size,
			Type:   "application/x-nzb",
		},
		Attrs: newznab.AttrsFromMap(fi.Attrs),
	}
}

type SearchResultStatus string

const (
	SearchResultStatusHit   SearchResultStatus = "hit"
	SearchResultStatusMiss  SearchResultStatus = "miss"
	SearchResultStatusError SearchResultStatus = "error"
)

func (srs SearchResultStatus) ShouldNotRequery() bool {

	if srs == SearchResultStatusHit || srs == SearchResultStatusMiss {
		return true
	}
	return false
}

type SearchCacheEntry struct {
	IndexerName        string
	Query              string
	FirstTried         time.Time
	LastTried          time.Time
	SearchResultStatus SearchResultStatus
	ErrorMessage       string
}

type NZBData struct {
	Title       string
	IndexerName string
	URL         string
}
