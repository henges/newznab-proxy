package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	srv := newznab.NewServer(prox, newznab.WithAPIKeyValidation(func() ([]string, error) {
		return []string{"0"}, nil
	}), newznab.WithMiddleware(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lmw := &loggingMiddleware{rw, 0}
			handler.ServeHTTP(lmw, r)
			dur := time.Since(start)
			if r.URL.Path == "/healthz" || r.URL.Path == "/favicon.ico" {
				return
			}
			log.Printf("%s %s %d %s", r.Method, r.URL, lmw.statusCode, dur)
		})
	}))
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

type loggingMiddleware struct {
	d          http.ResponseWriter
	statusCode int
}

func (l *loggingMiddleware) Header() http.Header {
	return l.d.Header()
}

func (l *loggingMiddleware) Write(bytes []byte) (int, error) {
	return l.d.Write(bytes)
}

func (l *loggingMiddleware) WriteHeader(statusCode int) {
	l.statusCode = statusCode
	l.d.WriteHeader(statusCode)
}
