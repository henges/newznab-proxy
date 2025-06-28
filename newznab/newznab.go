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
	APIKey   string `schema:"apikey"`
	Query    string `schema:"q"`
	Group    string `schema:"group"`
	Limit    int    `schema:"limit"`
	Category string `schema:"cat"`
	Output   string `schema:"o"`
	Attrs    string `schema:"attrs"`
	Extended bool   `schema:"extended"`
	Del      bool   `schema:"del"`
	MaxAge   int    `schema:"maxage"`
	Offset   int    `schema:"offset"`
}
