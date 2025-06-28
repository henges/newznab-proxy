package model

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
	ID              string
	IndexerName     string
	Title           string
	GUID            string
	GUIDIsPermalink bool
	Link            string
	Comments        string
	PubDate         time.Time
	Category        string
	Description     string
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
		ID:              id,
		IndexerName:     indexer,
		Title:           i.Title,
		GUID:            i.GUID.Value,
		GUIDIsPermalink: i.GUID.IsPermaLink,
		Link:            i.Link,
		Comments:        i.Comments,
		PubDate:         time.Time(i.PubDate),
		Category:        i.Category,
		Description:     i.Description,
		NZBLink:         i.Enclosure.URL,
		Size:            i.Enclosure.Length,
		Source:          source,
		Attrs:           i.AttrsMap(),
	}
}

func (fi FeedItem) RewrittenNZBLink(host string, tls bool) string {

	proto := "http"
	if tls {
		proto = "https"
	}
	return fmt.Sprintf("%s://%s/getnzb/%s", proto, host, fi.ID)
}

func (fi FeedItem) ToRewrittenNewznabItem(host string, tls bool) newznab.Item {

	ret := fi.ToNewznabItem()
	ret.Enclosure.URL = fi.RewrittenNZBLink(host, tls)
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
		Comments:    fi.Comments,
		PubDate:     newznab.RFC1123Time(fi.PubDate),
		Category:    fi.Category,
		Description: fi.Description,
		Enclosure: newznab.RssEnclosure{
			URL:    fi.NZBLink,
			Length: fi.Size,
			Type:   "application/x-nzb",
		},
		Attrs: newznab.AttrsFromMap(fi.Attrs),
	}
}
