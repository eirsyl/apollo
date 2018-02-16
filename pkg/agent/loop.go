package agent

import (
	"context"
	"github.com/eirsyl/apollo/pkg"
	"github.com/eirsyl/apollo/pkg/agent/redis"
	pb "github.com/eirsyl/apollo/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"time"
)

// ReconciliationLoop is responsible for ensuring redis instance state
type ReconciliationLoop struct {
	redis         *redis.Client
	client        pb.ManagerClient
	listener      *grpc.ClientConn
	done          chan bool
	scrapeResults chan redis.ScrapeResult
}

// NewReconciliationLoop creates a new loop and returns the instance
func NewReconciliationLoop(r *redis.Client, managerAddr string) (*ReconciliationLoop, error) {
	var opts []grpc.DialOption
	useTLS := viper.GetBool("managerTLS")

	if !useTLS {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(managerAddr, opts...)
	if err != nil {
		return nil, err
	}

	client := pb.NewManagerClient(conn)

	return &ReconciliationLoop{
		redis:         r,
		client:        client,
		listener:      conn,
		done:          make(chan bool),
		scrapeResults: make(chan redis.ScrapeResult),
	}, nil
}

// Run starts the loop
func (r *ReconciliationLoop) Run() error {
	/*
	* Ensure instance state
	*
	* - Run prechecks to make sure the redis instance is configured in cluster mode.
	* - Record cluster state
	* - Report state to manager
	* - Ask for optimal state
	* - Try to execute
	* - Go to record cluster state
	 */

	var precheckDone = make(chan bool)

	go func() {
		r.prechecks()
		precheckDone <- true
	}()

	select {
	case <-r.done:
		return nil
	case <-time.After(pkg.PrecheckWaitDelay):
		log.Warnf("Prechecks did not complete within %v", pkg.PrecheckWaitDelay)
		return nil
	case <-precheckDone:
		log.Infof("Prechecks passed, starting reconciler")
	}

	go r.collectScrape()

	func() {
		for {
			select {
			case <-r.done:
				return
			case <-time.After(2 * time.Second):
				err := r.iteration()
				if err != nil {
					log.Warnf("ReconciliationLoop iteration error: %v", err)
				}
			}
		}
	}()

	return nil
}

func (r *ReconciliationLoop) prechecks() {
	for {
		err := r.redis.RunPreflightTests()
		if err != nil {
			if err == err.(*redis.ErrInstanceIncompatible) {
				log.Fatal(err)
			}
		} else {
			return
		}
		time.Sleep(time.Second)
	}
}

func (r *ReconciliationLoop) iteration() error {
	log.Debug("Starting new iteration")
	// Scrape instance information
	err := r.redis.ScrapeInformation(&r.scrapeResults)
	if err != nil {
		log.Warnf("Could not scrape instance information: %v", err)
	} else {
		log.Debug("Instance scrape success")
	}

	status := pb.HealthRequest{
		InstanceId: "localhost",
		Ready:      false,
		Detail:     "prechecks failed",
	}
	res, err := r.client.AgentHealth(context.Background(), &status)
	if err != nil {
		log.Warnf("Could not send: %v", err)
	} else {
		log.Infof("Red: %v", res)
	}
	return nil
}

func (r *ReconciliationLoop) collectScrape() {
	log.Info("Collecting scrape metrics")
	for {
		scrapeResult := <-r.scrapeResults
		log.Info(scrapeResult)
	}
}

// Shutdown stops the loop
func (r *ReconciliationLoop) Shutdown() error {
	r.done <- true

	defer func() {
		err := r.listener.Close()
		if err != nil {
			log.Warnf("Cold not close the loop client: %v", err)
		}
	}()
	return nil
}
