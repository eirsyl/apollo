package agent

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Agent exports the proxy struct
type Agent struct {
	srv *http.Server
}

// Run func
func (a *Agent) Run() error {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	a.srv = &http.Server{
		Handler:      r,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Infof("Starting http server on %s", ":8080")
	return a.srv.ListenAndServe()
}

// Exit func
func (a *Agent) Exit() error {
	return a.srv.Shutdown(context.Background())
}
