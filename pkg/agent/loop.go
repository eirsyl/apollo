package agent

import (
	"context"
	"time"

	"errors"

	"strconv"

	"fmt"

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
			case <-time.After(pkg.ReconciliationLoopInterval):
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

	// Make sure the node changes is reflected by the manager
	err = r.reportInstanceState()
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
		Addr:            r.redis.GetAddr(),
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
		var result *contrib.NodeCommandResult
		var err error

		switch command.Command {
		case contrib.CommandAddSlots:
			result, err = r.addSlots(command)
		case contrib.CommandSetReplicate:
			result, err = r.setReplication(command)
		case contrib.CommandSetEpoch:
			result, err = r.setEpoch(command)
		case contrib.CommandJoinCluster:
			result, err = r.joinCluster(command)
		case contrib.CommandCountKeysInSlots:
			result, err = r.countKeysInSlots(command)
		case contrib.CommandSetSlotState:
			result, err = r.setSlotState(command)
		case contrib.CommandBumpEpoch:
			result, err = r.bumpEpoch(command)
		case contrib.CommandDelSlots:
			result, err = r.delSlot(command)
		case contrib.CommandMigrateSlots:
			result, err = r.migrateSlot(command)
		default:
			log.Warnf("Agent received unsupported command: %v %v", command.ID, command.Command)
			result, err = contrib.NewNodeCommandResult(command.ID, []string{"UNSUPPORTED COMMAND"}, false)
		}
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (r *ReconciliationLoop) reportResults(results []*contrib.NodeCommandResult) error {
	var cm []*pb.ReportExecutionRequest_ExecutionResult

	for _, result := range results {
		log.Infof("Sending command result for task: %v", result.ID)
		cm = append(cm, &pb.ReportExecutionRequest_ExecutionResult{
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

/**
 * Private cluster commands and helper utils
 */
func (r *ReconciliationLoop) addSlots(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	c := *command
	var slots []int
	for _, s := range c.Arguments {
		slot, err := strconv.Atoi(s)
		if err != nil {
			log.Warnf("Invalid slot value: %v", s)
		}
		slots = append(slots, slot)
	}

	res, err := r.redis.AddSlots(slots)
	if err != nil {
		log.Warnf("Redis error: %v", err)
	}

	return contrib.NewNodeCommandResult(c.ID, []string{res}, err == nil)
}

func (r *ReconciliationLoop) setReplication(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	c := *command
	if len(c.Arguments) != 1 {
		log.Warnf("Cannot set replication without a replication target")
		return contrib.NewNodeCommandResult(c.ID, []string{"INVALID REPLICATION TARGET"}, false)
	}
	res, err := r.redis.Replicate(c.Arguments[0])
	if err != nil {
		log.Warnf("Redis error: %v", err)
	}
	return contrib.NewNodeCommandResult(c.ID, []string{res}, err == nil)
}

func (r *ReconciliationLoop) setEpoch(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	c := *command
	if len(c.Arguments) != 1 {
		log.Warn("Cannot set epoch without a epoch value")
		return contrib.NewNodeCommandResult(c.ID, []string{"INVALID EPOCH"}, false)
	}
	e := c.Arguments[0]
	epoch, err := strconv.Atoi(e)
	if err != nil {
		log.Warnf("Received invalid epoch value: %v %v", epoch, err)
		return contrib.NewNodeCommandResult(c.ID, []string{"INVALID EPOCH"}, false)
	}
	res, err := r.redis.SetEpoch(epoch)
	if err != nil {
		log.Warnf("Redis error: %v", err)
	}
	return contrib.NewNodeCommandResult(c.ID, []string{res}, err == nil)
}

func (r *ReconciliationLoop) joinCluster(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	c := *command
	if len(c.Arguments) != 1 {
		log.Warn("Cannot join cluster without a node address")
		return contrib.NewNodeCommandResult(c.ID, []string{"INVALID NODE ID"}, false)
	}
	res, err := r.redis.JoinCluster(c.Arguments[0])
	if err != nil {
		log.Warnf("Redis error: %v", err)
	}
	time.Sleep(5 * time.Second)
	return contrib.NewNodeCommandResult(c.ID, []string{res}, err == nil)
}

func (r *ReconciliationLoop) countKeysInSlots(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	c := *command
	if len(c.Arguments) == 0 {
		log.Warn("Cannot count keys without slots")
		return contrib.NewNodeCommandResult(c.ID, []string{"NO SLOTS"}, false)
	}

	var slots []int
	for _, s := range c.Arguments {
		slot, err := strconv.Atoi(s)
		if err != nil {
			log.Warnf("Invalid slot value: %v", s)
		}
		slots = append(slots, slot)
	}

	var results []string
	success := true
	for _, slot := range slots {
		res, err := r.redis.CountKeysInSlot(slot)
		if err != nil {
			log.Warnf("Redis error: %v", err)
			success = false
		}
		results = append(results, fmt.Sprintf("%d=%d", slot, res))
	}

	return contrib.NewNodeCommandResult(c.ID, results, success)
}

func (r *ReconciliationLoop) setSlotState(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	c := *command
	if len(c.Arguments) < 3 {
		log.Warn("state, nodeID and slots is required to set slot state")
		return contrib.NewNodeCommandResult(c.ID, []string{"STATE,NODE_ID,SLOTS REQUIRED"}, false)
	}

	state := c.Arguments[0]
	nodeID := c.Arguments[1]

	var slots []string
	for _, s := range c.Arguments[2:] {
		slot, err := strconv.Atoi(s)
		if err != nil {
			log.Warnf("Invalid slot value: %v", s)
		}
		slots = append(slots, strconv.Itoa(slot))
	}

	var results []string
	success := true
	for _, slot := range slots {
		res, err := r.redis.SetSlotState(slot, state, nodeID)
		if err != nil {
			log.Warnf("Redis error: %v", err)
			success = false
		}
		results = append(results, res)
	}

	return contrib.NewNodeCommandResult(c.ID, results, success)
}

func (r *ReconciliationLoop) bumpEpoch(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	success := true
	res, err := r.redis.BumpEpoch()
	if err != nil {
		log.Warnf("Redis error: %v", err)
		success = false
	}

	return contrib.NewNodeCommandResult(command.ID, []string{res}, success)
}

func (r *ReconciliationLoop) delSlot(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	c := *command
	if len(c.Arguments) == 0 {
		log.Warn("Cannot sel slots without slots")
		return contrib.NewNodeCommandResult(c.ID, []string{"NO SLOTS"}, false)
	}

	var slots []int
	for _, s := range c.Arguments {
		slot, err := strconv.Atoi(s)
		if err != nil {
			log.Warnf("Invalid slot value: %v", s)
		}
		slots = append(slots, slot)
	}

	res, err := r.redis.DelSlots(slots)
	if err != nil {
		log.Warnf("Redis error: %v", err)
	}

	return contrib.NewNodeCommandResult(c.ID, []string{res}, err == nil)
}

func (r *ReconciliationLoop) migrateSlot(command *contrib.NodeCommand) (*contrib.NodeCommandResult, error) {
	c := *command
	if len(c.Arguments) > 3 {
		log.Warn("addr, fix and slots is required to migrate slot")
		return contrib.NewNodeCommandResult(c.ID, []string{"ADDR,FIX,SLOTS REQUIRED"}, false)
	}

	addr := c.Arguments[0]
	fixValue := c.Arguments[1]
	var fix bool

	if fixValue == "true" {
		fix = true
	} else {
		fix = false
	}

	var slots []int
	for _, s := range c.Arguments[2:] {
		slot, err := strconv.Atoi(s)
		if err != nil {
			log.Warnf("Invalid slot value: %v", s)
		}
		slots = append(slots, slot)
	}

	var results []string
	res, err := r.redis.MigrateSlots(slots, addr, fix)
	if err != nil {
		log.Warnf("Redis error: %v", err)
	}
	results = append(results, res)

	return contrib.NewNodeCommandResult(c.ID, results, err == nil)
}
