package orchestrator

import (
	"encoding/json"

	"github.com/coreos/bbolt"
	pb "github.com/eirsyl/apollo/pkg/api"
)

func nodeStore(db *bolt.DB, node *pb.StateRequest) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		node, err := NewNodeFromPb(node)
		if err != nil {
			return err
		}

		node.setObservationTime()

		buf, err := json.Marshal(node)
		if err != nil {
			return err
		}

		return b.Put([]byte(node.ID), buf)
	})
}

func nodeList(db *bolt.DB) (map[string]Node, error) {
	var nodes = map[string]Node{}

	err := db.View(func(tx *bolt.Tx) error {
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
