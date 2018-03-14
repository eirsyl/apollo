package agent

import (
	"context"
	"time"

	"errors"

	"github.com/eirsyl/apollo/pkg"
	"github.com/eirsyl/apollo/pkg/agent/redis"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/eirsyl/apollo/pkg/contrib"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

/**
 * This file contains the main reconciliation loop. The loop is responsible
 * for running startup checks, fetching commands from the manager, execute the
 * commands and report the status back to the manager.
 */

// ReconciliationLoop is responsible for ensuring redis node state
type ReconciliationLoop struct {
	redis           *redis.Client
	client          pb.ManagerClient
	listener        *grpc.ClientConn
	done            chan bool
	scrapeResults   chan redis.ScrapeResult
	Metrics         *Metrics
	skipPrechecks   bool
	hostAnnotations map[string]string
	nodeID          string
}

// NewReconciliationLoop creates a new loop and returns the instance
func NewReconciliationLoop(r *redis.Client, managerAddr string, skipPrechecks bool, hostAnnotations map[string]string) (*ReconciliationLoop, error) {
	var opts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
	}
	useTLS := viper.GetBool("managerTLS")

	if !useTLS {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(managerAddr, opts...)
	if err != nil {
		return nil, err
	}

	client := pb.NewManagerClient(conn)

	m, err := NewMetrics()
	if err != nil {
		return nil, err
	}

	return &ReconciliationLoop{
		redis:           r,
		client:          client,
		listener:        conn,
		done:            make(chan bool),
		scrapeResults:   make(chan redis.ScrapeResult),
		Metrics:         m,
		skipPrechecks:   skipPrechecks,
		hostAnnotations: hostAnnotations,
	}, nil
}

// Run starts the loop
func (r *ReconciliationLoop) Run() error {
	/*
	* Ensure node state
	*
	* - Run prechecks to make sure the redis node is configured in cluster mode.
	* - Record cluster state
	* - Report state to manager
	* - Ask for optimal state
	* - Try to execute
	* - Go to record cluster state
	 */

	var precheckDone = make(chan bool)

	go func() {
		if !r.skipPrechecks {
			r.prechecks()
		}
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
			case <-time.After(10 * time.Second):
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
			if err == err.(*redis.ErrNodeIncompatible) {
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
	// Scrape node information
	err := r.redis.ScrapeInformation(&r.scrapeResults)
	if err != nil {
		log.Warnf("Could not scrape node information: %v", err)
	} else {
		log.Debug("Node scrape success")
	}

	err = r.reportInstanceState()
	if err != nil {
		return err
	}

	commands, err := r.fetchCommands()
	if err != nil {
		return err
	}

	if commands == nil {
		return nil
	}

	results, err := r.performActions(commands)
	if err != nil {
		return err
	}

	return r.reportResults(results)
}

// reportInstanceState collects the instance state and tries to send the state to the manager
// Failed requests is retried for a fixed amount of time
func (r *ReconciliationLoop) reportInstanceState() error {
	var metricsChan = make(chan Metric, 1)
	isEmpty, _ := r.redis.IsEmpty()
	nodes, _ := r.redis.ClusterNodes()
	go r.Metrics.ExportMetrics(&metricsChan)

	// Set node id
	r.setNodeID(nodes)

	state := pb.StateRequest{
		IsEmpty:         isEmpty,
		Nodes:           *transformNodes(&nodes),
		HostAnnotations: *transformHostAnnotations(&r.hostAnnotations),
		Metrics:         *transformMetrics(&metricsChan),
	}

	var resultChan = make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, err := r.client.ReportState(context.Background(), &state)
				if err != nil {
					log.Warnf("Could not report instance state: %v", err)
				} else {
					log.Debug("State sent to manager")
					resultChan <- err
					return
				}
				time.Sleep(2 * time.Second)
			}
		}
	}()

	select {
	case <-time.After(10 * time.Second):
		return errors.New("could not push instance state, loop continues")
	case err := <-resultChan:
		return err
	}
}

// fetchCommands is responsible for fetching the´
func (r *ReconciliationLoop) fetchCommands() ([]*contrib.NodeCommand, error) {
	request := pb.NextExecutionRequest{
		NodeID: r.nodeID,
	}
	response, err := r.client.NextExecution(context.Background(), &request)
	if err != nil {
		return nil, err
	}

	var commands []*contrib.NodeCommand

	for _, command := range response.Commands {
		c, err := contrib.NewNodeCommand(command.Id, command.Command, command.Arguments)
		if err != nil {
			return nil, err
		}
		commands = append(commands, c)
	}

	return commands, nil
}

func (r *ReconciliationLoop) performActions(commands []*contrib.NodeCommand) ([]*contrib.NodeCommandResult, error) {
	if len(commands) == 0 {
		return nil, nil
	}

	log.Infof("Agent received %d commands from the manager", len(commands))

	var results []*contrib.NodeCommandResult

	for _, command := range commands {
		log.Infof("Running command: %v %v", command.ID, command.Command)
		// TODO: Run the actual command and construct a command result
		result, err := contrib.NewNodeCommandResult(command.ID, []string{}, true)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *ReconciliationLoop) reportResults(results []*contrib.NodeCommandResult) error {
	var cm []*pb.ExecutionResult

	for _, result := range results {
		log.Infof("Sending command result for task: %v", result.ID)
		cm = append(cm, &pb.ExecutionResult{
			Id:      result.ID,
			Result:  result.Result,
			Success: result.Success,
		})
	}

	request := pb.ReportExecutionRequest{
		NodeID:         r.nodeID,
		CommandResults: cm,
	}
	_, err := r.client.ReportExecutionResult(context.Background(), &request)

	return err
}

func (r *ReconciliationLoop) collectScrape() {
	log.Info("Collecting scrape metrics")
	for {
		scrapeResult := <-r.scrapeResults
		r.Metrics.RegisterMetric(scrapeResult)
	}
}

// setNodeID updates the nodeID used by the agent based on the nodes list retrieved from redis.
func (r *ReconciliationLoop) setNodeID(nodes []redis.ClusterNode) {
	var nodeID string
	for _, node := range nodes {
		if node.Myself {
			nodeID = node.NodeID
			break
		}
	}

	if nodeID != "" {
		if nodeID != r.nodeID {
			log.Infof("Updating nodeID, new ID: %s", nodeID)
			r.nodeID = nodeID
		}
	} else {
		log.Warn("Could not find nodeID, invalid nodes list")
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
