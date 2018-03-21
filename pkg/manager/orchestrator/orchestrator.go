package orchestrator

import (
	"context"
	"net"

	"fmt"

	"github.com/coreos/bbolt"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	grpc "google.golang.org/grpc"
)

// Server implements the GRPC orchestrator server used by agents to coordinate
// cluster changes.
type Server struct {
	listener   *net.Listener
	grpcServer *grpc.Server
	cluster    *Cluster
}

// NewServer creates a new GRPC orchestrator server
func NewServer(managerAddr string, db *bolt.DB, replication, minNodesCreate int) (*Server, error) {
	lis, err := net.Listen("tcp", managerAddr)
	if err != nil {
		return nil, err
	}
	server := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	cluster, err := NewCluster(replication, minNodesCreate, db)
	if err != nil {
		return nil, err
	}

	manager := &Server{
		listener:   &lis,
		grpcServer: server,
		cluster:    cluster,
	}

	log.Info("Initializing orchestrator server")
	pb.RegisterManagerServer(server, manager)
	pb.RegisterCLIServer(server, manager)

	return manager, nil
}

// Run starts the server and cluster reconciliation loop
func (s *Server) Run() error {
	var errChan = make(chan error, 1)

	go func() {
		errChan <- s.grpcServer.Serve(*s.listener)
	}()

	go func() {
		errChan <- s.cluster.Run()
	}()

	return <-errChan
}

// Shutdown stops the server gracefully
func (s *Server) Shutdown() error {
	return nil
}

/**
 * Manager GRPC
 */

// ReportState grpc endpoint
func (s *Server) ReportState(ctx context.Context, req *pb.StateRequest) (*pb.StateResponse, error) {
	err := s.cluster.ReportState(req)
	if err != nil {
		return nil, err
	}
	return &pb.StateResponse{}, nil
}

// NextExecution grpc endpoint
func (s *Server) NextExecution(ctx context.Context, req *pb.NextExecutionRequest) (*pb.NextExecutionResponse, error) {
	return s.cluster.NextExecution(req)
}

// ReportExecutionResult grpc endpoint
func (s *Server) ReportExecutionResult(ctx context.Context, req *pb.ReportExecutionRequest) (*pb.ReportExecutionResponse, error) {
	err := s.cluster.ReportExecution(req)
	if err != nil {
		return nil, err
	}

	return &pb.ReportExecutionResponse{}, nil
}

/**
 * CLI GRPC
 */

// Status sends the cluster status to the cli client
func (s *Server) Status(context.Context, *pb.EmptyMessage) (*pb.StatusResponse, error) {
	tasks, err := s.cluster.planner.StatusExport()
	if err != nil {
		return nil, err
	}

	status := pb.StatusResponse{
		Health: s.cluster.health.Int64(),
		State:  s.cluster.state.Int64(),
		Tasks:  tasks,
	}
	return &status, nil
}

// Nodes sends the cluster nodes to the cli client
func (s *Server) Nodes(context.Context, *pb.EmptyMessage) (*pb.NodesResponse, error) {
	nodes, err := s.cluster.nodeManager.allNodes()
	if err != nil {
		return nil, err
	}

	normalizeNodes := func(members *map[string]ClusterNeighbour) []string {
		var m []string
		for node := range *members {
			m = append(m, node)
		}
		return m
	}

	normalizeAnnotations := func(annotations *map[string]string) []string {
		var a []string
		for key, value := range *annotations {
			a = append(a, fmt.Sprintf("%s=%s", key, value))
		}
		return a
	}

	var resultNodes []*pb.Node
	for _, node := range nodes {
		resultNodes = append(resultNodes, &pb.Node{
			Id:              node.ID,
			Addr:            node.Addr,
			IsEmpty:         node.IsEmpty,
			Nodes:           normalizeNodes(&node.Nodes),
			HostAnnotations: normalizeAnnotations(&node.HostAnnotations),
			Online:          s.cluster.nodeManager.isOnline(&node),
		})
	}

	return &pb.NodesResponse{Nodes: resultNodes}, nil
}
