package orchestrator

import (
	"encoding/json"

	"time"

	"github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
)

// node manager is responsible for storing node related data in boltDB
type nodeManager struct {
	db        *bolt.DB
	startTime time.Time
}

// newNodeManager creates a new node manager
func newNodeManager(db *bolt.DB) (*nodeManager, error) {
	return &nodeManager{db: db, startTime: time.Now().UTC()}, nil
}

// updateNode updates a node based on the status received by the manager.
func (nm *nodeManager) updateNode(node *Node) error {
	return nm.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))

		node.setObservationTime()

		buf, err := json.Marshal(node)
		if err != nil {
			return err
		}

		return b.Put([]byte(node.ID), buf)
	})
}

// onlineNodes returns a list of nodes that have checked in a status within the last 2 min
func (nm *nodeManager) onlineNodes() ([]Node, error) {
	var nodes []Node

	if err := nm.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var node Node
			err := json.Unmarshal(v, &node)
			if err != nil {
				return err
			}

			if nm.isOnline(&node) {
				nodes = append(nodes, node)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return nodes, nil
}

// allNodes returns a map of all nodes in the DB
func (nm *nodeManager) allNodes() (map[string]Node, error) {
	var nodes = map[string]Node{}

	err := nm.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var node Node
			err := json.Unmarshal(v, &node)
			if err != nil {
				return err
			}
			nodes[string(k)] = node
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// garbageCollectNodes removes old and offline that is'nt considered part of the cluster
func (nm *nodeManager) garbageCollectNodes() error {
	clusterNodes, err := nm.getClusterNodes()
	if err != nil {
		return err
	}

	n := stringListToMap(clusterNodes)
	collected := 0

	return nm.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var node Node
			err := json.Unmarshal(v, &node)
			if err != nil {
				return err
			}

			if nm.isOnline(&node) {
				continue
			}

			if n[node.ID] {
				continue
			}

			err = b.Delete(k)
			if err != nil {
				return err
			}

			collected++
		}

		if collected > 0 {
			log.Infof("Node manager gc removed %d nodes", collected)
		}

		return nil
	})
}

/**
 * Cluster Status
 */

// setClusterNodes resets all the cluster nodes
func (nm *nodeManager) setClusterNodes(nodeIds []string) error {
	return nm.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("cluster"))

		buf, err := json.Marshal(nodeIds)
		if err != nil {
			return err
		}

		return b.Put([]byte("clusterNodes"), buf)
	})
}

// getClusterNodes returns the list of configured cluster members
func (nm *nodeManager) getClusterNodes() ([]string, error) {
	var nodeIds []string

	err := nm.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("cluster"))
		return json.Unmarshal(b.Get([]byte("clusterNodes")), &nodeIds)
	})

	if err != nil {
		return nil, err
	}

	return nodeIds, nil
}

// isOnline checks id a node is considered as online
func (nm *nodeManager) isOnline(node *Node) bool {
	lastAllowedTime := time.Now().Add(-30 * time.Second)
	return node.LastObservation.After(nm.startTime) && node.LastObservation.After(lastAllowedTime)
}
