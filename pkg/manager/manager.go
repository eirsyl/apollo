package manager

import (
	"context"
	"errors"
	"time"

	"github.com/eirsyl/apollo/pkg/manager/orchestrator"

	bolt "github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Manager exports the manager struct used to operate an manager.
type Manager struct {
	managerAddr  string
	httpAddr     string
	httpServer   *HTTPServer
	orchestrator *orchestrator.Server
	db           *bolt.DB
}

// NewManager initializes a new manager instance and returns a pinter to it.
func NewManager() (*Manager, error) {
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		log.Warnf("Could not open DB: %v", err)
		return nil, err
	}

	managerAddr := viper.GetString("managerAddr")
	if managerAddr == "" {
		return nil, errors.New("The manager address cannot be empty")
	}

	httpAddr := viper.GetString("debugAddr")
	if httpAddr == "" {
		return nil, errors.New("The debug address cannot be empty")
	}

	return &Manager{
		managerAddr: managerAddr,
		httpAddr:    httpAddr,
		db:          db,
	}, nil
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

	var errChan = make(chan error, 1)

	// Start the http debug server
	go func(errChan chan error) {
		httpServer, err := NewHTTPServer(m.httpAddr, map[string]string{
			"module":      "manager",
			"managerAddr": m.managerAddr,
			"httpAddr":    m.httpAddr,
		})
		if err != nil {
			errChan <- err
			return
		}

		m.httpServer = httpServer

		log.Infof("Starting http server on %s", m.httpServer.SRV.Addr)
		errChan <- m.httpServer.SRV.ListenAndServe()
	}(errChan)

	// Start orchestrator server
	go func(errChan chan error) {
		orchestratorServer, err := orchestrator.NewServer(m.managerAddr)
		if err != nil {
			errChan <- err
			return
		}

		m.orchestrator = orchestratorServer
		log.Infof("Starting orchestrator server on %s", m.managerAddr)
		errChan <- m.orchestrator.Run()
	}(errChan)

	err := <-errChan
	if err == nil {
		return errors.New("The http server or orchestrator server stopped unexpectedly")
	}
	return err
}

// Exit gracefully shuts down the manager
func (m *Manager) Exit() error {
	log.Info("Closing http server")
	err := m.httpServer.SRV.Shutdown(context.Background())
	if err != nil {
		log.Warnf("Could not gracefully stop the http server: %v", err)

	}

	log.Info("Closing orchestrator")
	err = m.orchestrator.Shutdown()
	if err != nil {
		log.Warnf("Could not stop orchestrator server: %v", err)
	}

	log.Info("Closing DB")
	err = m.db.Close()
	if err != nil {
		log.Warnf("Could not close DB: %v", err)
	}

	return nil
}
