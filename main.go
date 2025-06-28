package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/henges/newznab-proxy/newznab"
)

type dummyServerImpl struct{}

func (d dummyServerImpl) Search(params newznab.SearchParams) (*newznab.RssFeed, error) {

	ch := newznab.RssChannel{
		AtomLink: newznab.AtomLink{
			Href: "http://abcd.xyz",
			Rel:  "self",
			Type: "",
		},
		Title:       "hey!",
		Description: "",
		SiteLink:    "",
		Language:    "",
		WebMaster:   "",
		Category:    "",
		Image:       nil,
		Response:    newznab.NewznabResponse{},
		Items:       nil,
	}
	resp := newznab.NewRssFeed(ch)

	return &resp, nil
}

func main() {

	cl := newznab.NewClient("", "")
	search, err := cl.Search(context.Background(), newznab.SearchParams{
		Query: "tom james scott",
	})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(*search)
	}
}

func mainServer() {

	dum := dummyServerImpl{}
	srv := newznab.NewServer(dum)

	hsrv := http.Server{Addr: "0.0.0.0:80", Handler: srv.Handler()}

	hsrv.ListenAndServe()
}
