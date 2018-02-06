package manager

import (
	"context"

	"net/http"

	log "github.com/sirupsen/logrus"
)

// Manager exports the manager struct used to operate an manager.
type Manager struct {
	httpServer *HTTPServer
}

// NewManager initializes a new manager instance and returns a pinter to it.
func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

// Run starts the manager
func (m *Manager) Run() error {
	/*
	* Start the manager service
	*
	* Export metrics on http
	* GRPC server for agent to manager communication
	*
	 */
	httpServer, err := NewHTTPServer(
		":8000",
	)
	if err != nil {
		return err
	}
	m.httpServer = httpServer

	log.Infof("Starting http server on %s", m.httpServer.SRV.Addr)
	err = m.httpServer.SRV.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Exit gracefully shuts down the manager
func (m *Manager) Exit() error {
	err := m.httpServer.SRV.Shutdown(context.Background())
	return err
}
