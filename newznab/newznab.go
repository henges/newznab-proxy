package newznab

type Server interface {
	Search(params SearchParams) (*RssFeed, error)
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
