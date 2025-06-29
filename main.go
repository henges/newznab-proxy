package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/henges/newznab-proxy/newznab"
	"github.com/henges/newznab-proxy/proxy"
)

type dummyServerImpl struct{}

func (d dummyServerImpl) Search(ctx context.Context, params newznab.SearchParams) (*newznab.RssFeed, error) {

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

	ctx := context.Background()
	cfg := proxy.MustGetConfig()
	prox, err := proxy.NewProxy(ctx, cfg)
	if err != nil {
		log.Fatal(err)
		return
	}
	prox.StartRSSPolls(ctx)

	srv := newznab.NewServer(prox)
	hsrv := http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Web.ListenAddr, cfg.Web.Port),
		Handler: srv.Handler(),
	}
	go func() {
		hsrv.ListenAndServe()
	}()
	fmt.Println("Server up")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	err = hsrv.Shutdown(ctx)
	if err != nil {
		log.Printf("Error shutting down http server: %s", err)
	}
	err = prox.StopRSSPolls()
	if err != nil {
		log.Printf("Error shutting down RSS polls: %s", err)
	}
}

func mainClient() {

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
