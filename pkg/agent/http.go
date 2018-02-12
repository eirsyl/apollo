package agent

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPServer exposes an http server with prometheus monitoring enabled
type HTTPServer struct {
	server *http.Server
}

// NewHTTPServer creates a new HTTPServer
func NewHTTPServer(listenAddr string) (*HTTPServer, error) {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Handler:      r,
		Addr:         listenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &HTTPServer{server: server}, nil
}

// Run starts the server and listens for incoming connections
func (s *HTTPServer) Run() error {
	err := s.server.ListenAndServe()
	if err == http.ErrServerClosed {
		// Don't fail if the server is stopped gracefully
		return nil
	}
	return err
}

// GetListenAddr returns the address the server is listening on
func (s *HTTPServer) GetListenAddr() string {
	return s.server.Addr
}

// Shutdown closes the server gracefully
func (s *HTTPServer) Shutdown() error {
	return s.server.Shutdown(context.Background())
}