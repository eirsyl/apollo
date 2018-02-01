package manager

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/eirsyl/apollo/pkg/api"
)

// Manager exports the manager struct used to operate an manager.
type Manager struct {
}

// NewManager initializes a new manager instance and returns a pinter to it.
func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

// Run starts the manager
func (m *Manager) Run() error {
	lis, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	pb.RegisterManagerServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
}

// Exit gracefully shuts down the manager
func (m *Manager) Exit() error {
	return nil
}

type server struct {
}

func (s *server) SendHealth(ctx context.Context, in *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Healthy: !in.Healthy}, nil
}
