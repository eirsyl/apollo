package manager

import (
	"context"
	"net/http"
	"time"

	"strconv"

	"github.com/coreos/bbolt"
	"github.com/eirsyl/apollo/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPServer exposes an http server with prometheus monitoring enabled
type HTTPServer struct {
	server *http.Server
}

// NewHTTPServer creates a new HTTPServer
func NewHTTPServer(listenAddr string, buildInfo map[string]string, db *bolt.DB) (*HTTPServer, error) {
	r := mux.NewRouter()

	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/", utils.BuildInformationHandler(buildInfo))
	r.HandleFunc("/backup", backupHandleFunc(db))

	srv := &http.Server{
		Handler:      r,
		Addr:         listenAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &HTTPServer{server: srv}, nil
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

func backupHandleFunc(db *bolt.DB) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := db.View(func(tx *bolt.Tx) error {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="apollo.db"`)
			w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
			_, err := tx.WriteTo(w)
			return err
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
