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
