package proxy

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Proxy exports the proxy struct
type Proxy struct {
	srv *http.Server
}

// Run func
func (p *Proxy) Run() error {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	p.srv = &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Infof("Starting http server on %s", ":8080")
	return p.srv.ListenAndServe()
}

// Exit func
func (p *Proxy) Exit() error {
	return p.srv.Shutdown(context.Background())
}
