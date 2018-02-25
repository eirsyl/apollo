package manager

import (
	"errors"
	"fmt"
	"time"

	"github.com/eirsyl/apollo/pkg/manager/orchestrator"

	bolt "github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
)

// Manager exports the manager struct used to operate an manager.
type Manager struct {
	managerAddr  string
	httpAddr     string
	httpServer   *HTTPServer
	orchestrator *orchestrator.Server
	db           *bolt.DB
	replication  int
}

// NewManager initializes a new manager instance and returns a pinter to it.
func NewManager(managerAddr, httpAddr, databaseFile string, replicationFactor int) (*Manager, error) {
	if managerAddr == "" {
		return nil, errors.New("The manager address cannot be empty")
	}

	if httpAddr == "" {
		return nil, errors.New("The debug address cannot be empty")
	}

	if databaseFile == "" {
		return nil, errors.New("The database path cannot be empty")
	}

	db, err := bolt.Open(databaseFile, 0600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}

	if !(replicationFactor > 0) {
		return nil, errors.New("The replication factor must be larger than 0")
	}

	return &Manager{
		managerAddr: managerAddr,
		httpAddr:    httpAddr,
		db:          db,
		replication: replicationFactor,
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
			"module":            "manager",
			"managerAddr":       m.managerAddr,
			"httpAddr":          m.httpAddr,
			"replicationFactor": fmt.Sprintf("%d", m.replication),
		})
		if err != nil {
			errChan <- err
			return
		}

		m.httpServer = httpServer

		log.Infof("Starting http server on %s", httpServer.GetListenAddr())
		errChan <- httpServer.Run()
	}(errChan)

	// Start orchestrator server
	go func(errChan chan error) {
		orchestratorServer, err := orchestrator.NewServer(m.managerAddr, m.replication)
		if err != nil {
			errChan <- err
			return
		}

		m.orchestrator = orchestratorServer
		log.Infof("Starting orchestrator server on %s", m.managerAddr)
		errChan <- m.orchestrator.Run()
	}(errChan)

	return <-errChan
}

// Exit gracefully shuts down the manager
func (m *Manager) Exit() error {
	log.Info("Closing http server")
	err := m.httpServer.Shutdown()
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
