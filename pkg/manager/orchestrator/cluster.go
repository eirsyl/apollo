package orchestrator

import (
	"time"

	"github.com/coreos/bbolt"
	pb "github.com/eirsyl/apollo/pkg/api"
	log "github.com/sirupsen/logrus"
)

type clusterState int
type clusterHealth int

const (
	ClusterUnknown      clusterState = 0
	ClusterUnconfigured clusterState = 1
	ClusterConfigured   clusterState = 2

	ClusterOK    clusterHealth = 0
	ClusterWarn  clusterHealth = 1
	ClusterError clusterHealth = 2
)

// Cluster stores the manager observations of the cluster
type Cluster struct {
	state              clusterState
	health             clusterHealth
	desiredReplication int
	minNodesCreate     int
	db                 *bolt.DB
	startTime          time.Time
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

	return &Cluster{
		state:              ClusterUnknown,
		health:             ClusterOK,
		desiredReplication: desiredReplication,
		minNodesCreate:     minNodesCreate,
		db:                 db,
		startTime:          time.Now().UTC(),
	}, nil
}

// Run starts the manager reconciliation loop
// The loop is responsible for calculating optimal cluster state
func (c *Cluster) Run() error {
	for {
		switch c.state {
		case ClusterUnknown:
			c.findClusterConfiguration()
		case ClusterUnconfigured:
			c.configureCluster()
		case ClusterConfigured:
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

	var onlineNodes = 0
	var latestAllowedObservation = time.Now().UTC().Add(-2 * time.Minute)
	for _, node := range nodes {
		if node.LastObservation.After(c.startTime) && node.LastObservation.After(latestAllowedObservation) {
			onlineNodes++
		}
	}
	log.Infof("Total nodes: %d, online nodes: %d", len(nodes), onlineNodes)
	if onlineNodes < 3 {
		log.Infof("Skipping configuration detection, minimum 3 nodes has to be online and reporting state.")
		return
	}
}

// configureCluster configures the redis nodes as a cluster if all the requirements is meet
func (c *Cluster) configureCluster() {

}

// iteration watches the cluster after everything is configured and up an running
func (c *Cluster) iteration() {

}

// ReportState collects the state from the reporting node
func (c *Cluster) ReportState(node *pb.StateRequest) error {
	return nodeStore(c.db, node)
}
