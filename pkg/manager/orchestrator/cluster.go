package orchestrator

import (
	"time"

	"github.com/coreos/bbolt"
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
}

// NewCluster creates a new cluster planner
func NewCluster(desiredReplication, minNodesCreate int, db *bolt.DB) (*Cluster, error) {
	return &Cluster{
		state:              ClusterUnknown,
		health:             ClusterOK,
		desiredReplication: desiredReplication,
		minNodesCreate:     minNodesCreate,
		db:                 db,
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

func (c *Cluster) findClusterConfiguration() {
	log.Info("Finding cluster configuration")
}

func (c *Cluster) configureCluster() {

}

func (c *Cluster) iteration() {

}
