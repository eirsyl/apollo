package agent

import (
	"github.com/eirsyl/apollo/pkg/agent/redis"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	"net"
)

// Executor implements an GRPC server and a reconciliation loop
// that it used for detecting problems with the redis instance.
type Executor struct {
	redis       *redis.Client
	agentAddr   string
	managerAddr string
	listener    *net.Listener
	server      *grpc.Server
	loop        *ReconciliationLoop
}

// NewExecutor creates a new Executor instance
func NewExecutor(agentAddr, managerAddr string, redis *redis.Client) (*Executor, error) {
	lis, err := net.Listen("tcp", agentAddr)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	loop, err := NewReconciliationLoop(redis, managerAddr)
	if err != nil {
		return nil, err
	}

	executor := &Executor{
		agentAddr:   agentAddr,
		managerAddr: managerAddr,
		redis:       redis,
		listener:    &lis,
		server:      server,
		loop:        loop,
	}

	log.Info("Initializing executor server")
	pb.RegisterAgentServer(server, executor)

	return executor, nil
}

// Run starts the executor instance
func (e *Executor) Run() error {
	var errChan = make(chan error, 1)

	go func() {
		errChan <- e.server.Serve(*e.listener)
	}()

	go func() {
		errChan <- e.loop.Run()
	}()

	return <-errChan
}

// GetListenAddr returns the address the executor server is listening on
func (e *Executor) GetListenAddr() string {
	return e.agentAddr
}

// Shutdown stops the executor gracefully
func (e *Executor) Shutdown() error {
	return nil
}
