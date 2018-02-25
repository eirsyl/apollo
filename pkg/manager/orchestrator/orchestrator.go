package orchestrator

import (
	"context"
	"net"

	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

// Server implements the GRPC orchestrator server used by agents to coordinate
// cluster changes.
type Server struct {
	listener    *net.Listener
	grpcServer  *grpc.Server
	replication int
}

// NewServer creates a new GRPC orchestrator server
func NewServer(managerAddr string, replication int) (*Server, error) {
	lis, err := net.Listen("tcp", managerAddr)
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	manager := &Server{
		listener:    &lis,
		grpcServer:  server,
		replication: replication,
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

// ReportState grpc endpoint
func (s *Server) ReportState(ctx context.Context, req *pb.StateRequest) (*pb.StateResponse, error) {
	log.Infof("Request: %v", req)
	return &pb.StateResponse{}, nil
}

// NextExecution grpc endpoint
func (s *Server) NextExecution(ctx context.Context, req *pb.NextExecutionRequest) (*pb.NextExecutionResponse, error) {
	return &pb.NextExecutionResponse{}, nil
}

// ReportExecutionResult grpc endpoint
func (s *Server) ReportExecutionResult(ctx context.Context, req *pb.ReportExecutionRequest) (*pb.ReportExecutionResponse, error) {
	return &pb.ReportExecutionResponse{}, nil
}
