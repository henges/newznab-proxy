package newznab

import "encoding/xml"

type ServerImplementation interface {
	Search(params SearchParams) (*RssFeed, error)
}

type ServerError struct {
	XMLName     xml.Name `xml:"error"`
	Code        int      `xml:"code"`
	Description string   `xml:"description"`
}

type SearchParams struct {
	APIKey   string
	Query    string
	Group    string
	Limit    int
	Category string
	Output   string
	Attrs    string
	Extended bool
	Del      bool
	MaxAge   int
	Offset   int
}
