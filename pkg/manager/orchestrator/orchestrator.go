package orchestrator

import (
	"context"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	"net"
)

// Server implements the GRPC orchestrator server used by agents to coordinate
// cluster changes.
type Server struct {
	listener   *net.Listener
	grpcServer *grpc.Server
}

// NewServer creates a new GRPC orchestrator server
func NewServer(managerAddr string) (*Server, error) {
	lis, err := net.Listen("tcp", managerAddr)
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	manager := &Server{
		listener:   &lis,
		grpcServer: server,
	}

	log.Info("Initializing orchestrator server")
	pb.RegisterManagerServer(server, manager)

	return manager, nil
}

// Run starts the server
func (s *Server) Run() error {
	return s.grpcServer.Serve(*s.listener)
}

// Shutdown stops the server gracefully
func (s *Server) Shutdown() error {
	return nil
}

// AgentHealth grpc endpoint
func (s *Server) AgentHealth(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	log.Infof("Request: %v", req)
	return &pb.HealthResponse{Ack: true}, nil
}
