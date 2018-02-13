package agent

import (
	"context"
	"github.com/eirsyl/apollo/pkg/agent/redis"
	pb "github.com/eirsyl/apollo/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"time"
)

// ReconciliationLoop is responsible for ensuring redis instance state
type ReconciliationLoop struct {
	redis  *redis.Client
	client *pb.ManagerClient
}

// NewReconciliationLoop creates a new loop and returns the instance
func NewReconciliationLoop(redis *redis.Client, managerAddr string) (*ReconciliationLoop, error) {
	var opts []grpc.DialOption
	useTLS := viper.GetBool("managerTLS")

	if !useTLS {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(managerAddr, opts...)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Warnf("Could not close client connection: %v", err)
		}
	}()

	client := pb.NewManagerClient(conn)

	return &ReconciliationLoop{
		redis:  redis,
		client: &client,
	}, nil
}

// Run starts the loop
func (r *ReconciliationLoop) Run() error {
	// Post instance health
	func() {
		for {
			status := pb.HealthRequest{
				InstanceId: "localhost",
				Ready:      false,
				Detail:     "prechecks failed",
			}
			res, err := *(r.client).AgentHealth(context.Background(), &status)
			if err != nil {
				log.Warnf("Could not send: %v", err)
			} else {
				log.Infof("Red: %v", res)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	return nil
}

// Shutdown stops the loop
func (r *ReconciliationLoop) Shutdown() error {
	return nil
}
