package newznab

import (
	"time"

	"encoding/xml"

	"github.com/henges/newznab-proxy/xmlutil"
)

type RssFeed struct {
	XMLName      xml.Name   `xml:"rss"`
	Version      string     `xml:"version,attr"`
	XmlnsAtom    string     `xml:"xmlns:atom,attr"`
	XmlnsNewznab string     `xml:"xmlns:newznab,attr"`
	Channel      RssChannel `xml:"channel"`
}

func NewRssFeed(v RssChannel) RssFeed {
	// <rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom" xmlns:newznab="http://www.newznab.com/DTD/2010/feeds/attributes/">
	return RssFeed{
		Version:      "2.0",
		XmlnsAtom:    "http://www.w3.org/2005/Atom",
		XmlnsNewznab: "http://www.newznab.com/DTD/2010/feeds/attributes/",
		Channel:      v,
	}
}

func NewRssFeedFromItems(offset, total int, v []Item) RssFeed {

	ch := RssChannel{
		Response: NewznabResponse{
			Offset: offset,
			Total:  total,
		},
		Items: v,
	}
	return NewRssFeed(ch)
}

type RssChannel struct {
	AtomLink    AtomLink      `xml:"http://www.w3.org/2005/Atom atom:link"`
	Title       string        `xml:"title"`
	Description string        `xml:"description"`
	SiteLink    string        `xml:"link"`
	Language    string        `xml:"language"`
	WebMaster   string        `xml:"webMaster"`
	Category    string        `xml:"category"`
	Image       *ChannelImage `xml:"image,omitempty"`
	Response    NewznabResponse
	Items       []Item `xml:"item"`
}

type AtomLink struct {
	XMLName xml.Name `xml:"atom:link"`
	Href    string   `xml:"href,attr"`
	Rel     string   `xml:"rel,attr"`
	Type    string   `xml:"type,attr"`
}

type ChannelImage struct {
	URL         string `xml:"url"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

type NewznabResponse struct {
	XMLName xml.Name `xml:"newznab:response"`
	Offset  int      `xml:"offset,attr"`
	Total   int      `xml:"total,attr"`
}

type Item struct {
	Title       string        `xml:"title"`
	GUID        RssGuid       `xml:"guid"`
	Link        string        `xml:"link"`
	Comments    string        `xml:"comments"`
	PubDate     RFC1123Time   `xml:"pubDate"` // RFC1123 with numeric TZ
	Category    string        `xml:"category"`
	Description string        `xml:"description"`
	Enclosure   RssEnclosure  `xml:"enclosure"`
	Attrs       []NewznabAttr `xml:"newznab:attr"`
}

func (r Item) AttrsMap() map[string]string {

	ret := make(map[string]string, len(r.Attrs))
	for _, attr := range r.Attrs {
		ret[attr.Name] = attr.Value
	}
	return ret
}

func AttrsFromMap(m map[string]string) []NewznabAttr {
	ret := make([]NewznabAttr, 0, len(m))
	for k, v := range m {
		ret = append(ret, NewznabAttr{
			Name:  k,
			Value: v,
		})
	}
	return ret
}

type RssGuid struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

type RssEnclosure struct {
	URL    string `xml:"url,attr"`
	Length int64  `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type NewznabAttr struct {
	XMLName xml.Name `xml:"newznab:attr"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type RFC1123Time time.Time

func (r *RFC1123Time) UnmarshalXML(d *xmlutil.UnmarshalDecoder, start xmlutil.UnmarshalStartElement) error {

	var v string
	err := d.DecodeElement(&v, &start)
	if err != nil {
		return err
	}
	t, err := time.Parse(time.RFC1123Z, v)
	if err != nil {
		return err
	}
	*r = RFC1123Time(t)
	return nil
}

func (r *RFC1123Time) MarshalXML(e *xmlutil.MarshalEncoder, start xmlutil.MarshalStartElement) error {

	v := time.Time(*r)
	strval := v.Format(time.RFC1123Z)
	return e.EncodeElement(strval, start)
}

func (r RFC1123Time) String() string {

	return time.Time(r).String()
}
