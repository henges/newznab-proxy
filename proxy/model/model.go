package model

import (
	"time"

	"github.com/henges/newznab-proxy/newznab"
)

type FeedItemSource string

const (
	FeedItemSourceRSS    FeedItemSource = "rss"
	FeedItemSourceSearch FeedItemSource = "search"
)

type FeedItem struct {
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
	Attrs           map[string]string
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
