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

	executor := &Executor{
		agentAddr:   agentAddr,
		managerAddr: managerAddr,
		redis:       redis,
		listener:    &lis,
		server:      server,
	}

	log.Info("Initializing executor server")
	pb.RegisterAgentServer(server, executor)

	return executor, nil
}

// Run starts the executor instance
func (e *Executor) Run() error {
	err := e.server.Serve(*e.listener)
	return err
}

// GetListenAddr returns the address the executor server is listening on
func (e *Executor) GetListenAddr() string {
	return e.agentAddr
}

// Shutdown stops the executor gracefully
func (e *Executor) Shutdown() error {
	return nil
}
