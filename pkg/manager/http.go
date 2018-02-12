package manager

import (
	"net/http"
	"time"

	"github.com/eirsyl/apollo/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPServer exposes an http server with prometheus monitoring enabled
type HTTPServer struct {
	SRV *http.Server
}

// NewHTTPServer creates a new HTTPServer
func NewHTTPServer(listenAddr string, buildInfo map[string]string) (*HTTPServer, error) {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/", utils.BuildInformationHandler(buildInfo))

	srv := &http.Server{
		Handler:      r,
		Addr:         listenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &HTTPServer{SRV: srv}, nil
}
