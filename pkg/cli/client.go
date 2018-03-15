package cli

import (
	pb "github.com/eirsyl/apollo/pkg/api"
	"google.golang.org/grpc"
)

// NewCLIClient returns a new cli client
func NewCLIClient(addr string) (*pb.CLIClient, error) {
	var opts = []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	client := pb.NewCLIClient(conn)
	return &client, nil
}
