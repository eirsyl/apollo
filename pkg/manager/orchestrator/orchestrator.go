package orchestrator

import (
	"fmt"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
	"io"
	"net"
	"time"
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
func (s *Server) AgentHealth(stream pb.Manager_AgentHealthServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		fmt.Printf("Received: %d\n", req.InstanceId)
		time.Sleep(2 * time.Second)
		resp := &pb.HealthResponse{
			Ack: true,
		}
		err = stream.Send(resp)
		if err != nil {
			return err
		}
	}
	return nil
}
