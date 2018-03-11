package orchestrator

import (
	"time"

	"github.com/coreos/bbolt"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/eirsyl/apollo/pkg/manager/orchestrator/planner"
	log "github.com/sirupsen/logrus"
)

type clusterState int
type clusterHealth int

const (
	clusterUnknown      clusterState = 0
	clusterUnconfigured clusterState = 1
	clusterConfigured   clusterState = 2

	clusterOK    clusterHealth = 0
	clusterWarn  clusterHealth = 1
	clusterError clusterHealth = 2
)

// Cluster stores the manager observations of the cluster
type Cluster struct {
	state              clusterState
	health             clusterHealth
	desiredReplication int
	minNodesCreate     int
	db                 *bolt.DB
	startTime          time.Time
	planner            *planner.Planner
}

// NewCluster creates a new cluster planner
func NewCluster(desiredReplication, minNodesCreate int, db *bolt.DB) (*Cluster, error) {
	// Create buckets
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("nodes"))
		return err
	})
	if err != nil {
		return nil, err
	}

	p, err := planner.NewPlanner()
	if err != nil {
		return nil, err
	}

	return &Cluster{
		state:              clusterUnknown,
		health:             clusterError,
		desiredReplication: desiredReplication,
		minNodesCreate:     minNodesCreate,
		db:                 db,
		startTime:          time.Now().UTC(),
		planner:            p,
	}, nil
}

// Run starts the manager reconciliation loop
// The loop is responsible for calculating optimal cluster state
func (c *Cluster) Run() error {
	for {
		switch c.state {
		case clusterUnknown:
			c.findClusterConfiguration()
		case clusterUnconfigured:
			c.configureCluster()
		case clusterConfigured:
			c.iteration()
		default:
			log.WithField("state", c.state).Warnf("Unhandled cluster state")
		}
		time.Sleep(10 * time.Second)
	}
}

// findClusterConfiguration is responsible for trying to figure out if the cluster is configured or not
func (c *Cluster) findClusterConfiguration() {
	log.Info("Starting configuration detection")
	nodes, err := nodeList(c.db)
	if err != nil {
		log.Warn("Could not retrieve nodes, skipping configuration detection: %v", err)
		return
	}

	if len(nodes) < 3 {
		log.Infof("Skipping configuration detection 3 or more nodes is required, current count: %d", len(nodes))
		return
	}

	var onlineNodes = []Node{}
	var latestAllowedObservation = time.Now().UTC().Add(-2 * time.Minute)
	for _, node := range nodes {
		if node.LastObservation.After(c.startTime) && node.LastObservation.After(latestAllowedObservation) {
			onlineNodes = append(onlineNodes, node)
		}
	}
	log.Infof("Total nodes: %d, online nodes: %d", len(nodes), len(onlineNodes))
	if len(onlineNodes) < 3 {
		log.Infof("Skipping configuration detection, minimum 3 nodes has to be online and reporting state.")
		return
	}

	var onlineNodesEmpty = true
	for _, node := range onlineNodes {
		if !node.IsEmpty {
			onlineNodesEmpty = false
			break
		}
	}

	if onlineNodesEmpty {
		log.Infof("Online nodes empty, setting cluster state to unconfigured and preparing for cluster initialization")
		c.state = clusterUnconfigured
	} else {
		log.Infof("Nodes not empty, setting cluster to configured")
		c.state = clusterConfigured
	}
}

// configureCluster configures the redis nodes as a cluster if all the requirements is meet
func (c *Cluster) configureCluster() {
	nodes, err := nodeList(c.db)
	if err != nil {
		log.Warn("Could not retrieve nodes, skipping configuration detection: %v", err)
		return
	}

	var onlineNodes = []Node{}
	var latestAllowedObservation = time.Now().UTC().Add(-2 * time.Minute)
	for _, node := range nodes {
		if node.LastObservation.After(c.startTime) && node.LastObservation.After(latestAllowedObservation) {
			onlineNodes = append(onlineNodes, node)
		}
	}

	if len(onlineNodes) < c.minNodesCreate {
		log.Infof("Skipping cluster creation %d nodes is required", c.minNodesCreate)
		return
	}

	log.Infof("Generating cluster creation plan")
	var nodeIds []string
	for _, node := range onlineNodes {
		nodeIds = append(nodeIds, node.ID)
	}

	validConfiguration := validateClusterSize(len(nodeIds), c.desiredReplication)
	if !validConfiguration {
		log.Warnf("invalid cluster, Redis Cluster requires at least 3 master nodes. %d nodes is required with the replication %d", 3*(c.desiredReplication+1), c.desiredReplication)
		return
	}

	// Create the cluster creation task and update cluster state
	err = setClusterNodes(c.db, nodeIds)
	if err != nil {
		log.Warnf("Could not update cluster members: %v", err)
		return
	}

	slots, err := allocSlots(&onlineNodes, c.desiredReplication)

	err = c.planner.NewCreateClusterTask(slots)
	if err != nil {
		log.Infof("Could not generate cluster creation plan: %v", err)
		return
	}

	log.Info("Updating cluster state")
	c.health = clusterWarn
	c.state = clusterConfigured
}

// iteration watches the cluster after everything is configured and up an running
func (c *Cluster) iteration() {
	l := log.WithFields(log.Fields{"clusterHealth": c.health, "clusterState": c.state})
	l.Info("Running iteration")

	t, err := c.planner.CurrentTask()
	if err != nil {
		l.Infof("Planner error: %v", err)
	}

	for id, command := range t.Commands {
		log.Infof("commandID: %v, Command: %v", id, *command)
	}
}

// ReportState collects the state from the reporting node
func (c *Cluster) ReportState(node *pb.StateRequest) error {
	return nodeStore(c.db, node)
}

// NextExecution sends the next command to a node. The planner returns a command if
// a step is planned by the manager.
func (c *Cluster) NextExecution(req *pb.NextExecutionRequest) (*pb.NextExecutionResponse, error) {
	// TODO: Lookup commands and send to node
	return &pb.NextExecutionResponse{
		Commands: []*pb.ExecutionCommand{},
	}, nil
}

// ReportExecution reports the status of a command executed on a node
func (c *Cluster) ReportExecution(req *pb.ReportExecutionRequest) error {
	// TODO: Parse and store execution result
	return c.planner.ReportResult(req.NodeID, &planner.CommandResult{})
}

/**
 * Functions bellow this line is used to generate and manage cluster state
 */
