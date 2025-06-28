package newznab

import (
	"encoding/xml"
	"fmt"
	"strings"
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
	// Query is the search input (URL/UTF-8 encoded). Case insensitive.
	Query string `schema:"q"`

	// Group is a list of usenet groups to search, delimited by “,”
	Group string `schema:"group,omitempty"`

	// Limit is the upper limit for the number of items to be returned.
	Limit int `schema:"limit,omitempty"`

	// Category is a list of categories to search, delimited by “,”
	Category string `schema:"cat,omitempty"`

	// Output is the output format, either “JSON” or “XML”. Default is “XML”.
	Output string `schema:"o,omitempty"`

	// Attrs is the list of requested extended attributes, delimited by “,”
	Attrs string `schema:"attrs,omitempty"`

	// Extended lists all extended attributes (attrs is ignored if true)
	Extended bool `schema:"extended,omitempty"`

	// Del deletes the item from a user's cart on download if set to true
	Del bool `schema:"del,omitempty"`

	// MaxAge only returns results posted to usenet in the last X days.
	MaxAge int `schema:"maxage,omitempty"`

	// Offset is the 0-based query offset defining which part of the response we want.
	Offset int `schema:"offset,omitempty"`
}

func (s SearchParams) Categories() []string {

	return strings.Split(s.Category, ",")
}

func (s SearchParams) Attributes() []string {

	return strings.Split(s.Attrs, ",")
}
