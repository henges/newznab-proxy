package newznab

import (
	"encoding/xml"
	"fmt"
)

type ServerImplementation interface {
	Search(params SearchParams) (*RssFeed, error)
}

type ServerError struct {
	XMLName     xml.Name `xml:"error"`
	Code        int      `xml:"code"`
	Description string   `xml:"description"`
}

func (s ServerError) Error() string {
	return fmt.Sprintf("%d: %s", s.Code, s.Description)
}

type SearchParams struct {
	//APIKey   string `schema:"apikey"`
	Query    string `schema:"q"`
	Group    string `schema:"group,omitempty"`
	Limit    int    `schema:"limit,omitempty"`
	Category string `schema:"cat,omitempty"`
	Output   string `schema:"o,omitempty"`
	Attrs    string `schema:"attrs,omitempty"`
	Extended bool   `schema:"extended,omitempty"`
	Del      bool   `schema:"del,omitempty"`
	MaxAge   int    `schema:"maxage,omitempty"`
	Offset   int    `schema:"offset,omitempty"`
}
