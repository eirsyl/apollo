package orchestrator

import (
	"time"

	pb "github.com/eirsyl/apollo/pkg/api"
)

// ClusterNeighbour stores the neighbour each node observes as a member of the cluster
type ClusterNeighbour struct {
	ID          string   `json:"id"`
	Addr        string   `json:"addr"`
	Role        string   `json:"role"`
	MasterID    string   `json:"masterId"`
	PingSent    int64    `json:"pingSent"`
	PingRecv    int64    `json:"pingRecv"`
	ConfigEpoch int64    `json:"configEpoch"`
	LinkStatus  string   `json:"linkStatus"`
	Slots       []string `json:"slots"`
}

// Node represents an member of the cluster.
type Node struct {
	ID              string                      `json:"id"`
	IsEmpty         bool                        `json:"isEmpty"`
	MySelf          ClusterNeighbour            `json:"myself"`
	Nodes           map[string]ClusterNeighbour `json:"nodes"`
	HostAnnotations map[string]string           `json:"annotations"`
	Metrics         map[string]float64          `json:"metrics"`
	LastObservation time.Time                   `json:"lastObservation"`
}

func newClusterNeighbourFromPb(nodes []*pb.ClusterNode) (map[string]ClusterNeighbour, ClusterNeighbour, error) {
	res := map[string]ClusterNeighbour{}
	var myself ClusterNeighbour

	for _, node := range nodes {
		cn := ClusterNeighbour{
			ID:          node.NodeID,
			Addr:        node.Addr,
			Role:        node.Role,
			MasterID:    node.MasterID,
			PingSent:    node.PingSent,
			PingRecv:    node.PingRecv,
			ConfigEpoch: node.ConfigEpoch,
			LinkStatus:  node.LinkStatus,
			Slots:       node.Slots,
		}
		res[cn.ID] = cn

		if node.Myself {
			myself = cn
		}
	}

	return res, myself, nil
}

func newHostAnnotationsFromPb(node *pb.StateRequest) (map[string]string, error) {
	var res = map[string]string{}
	for _, annotation := range node.HostAnnotations {
		res[annotation.Name] = annotation.Value
	}
	return res, nil
}

func newMetricsFromPb(node *pb.StateRequest) (map[string]float64, error) {
	var res = map[string]float64{}
	for _, metric := range node.Metrics {
		res[metric.Name] = metric.Value
	}
	return res, nil
}

// NewNodeFromPb is a shortcut for creating a new node struct.
func NewNodeFromPb(node *pb.StateRequest) (*Node, error) {
	nodes, myself, err := newClusterNeighbourFromPb(node.Nodes)
	if err != nil {
		return nil, err
	}

	annotations, err := newHostAnnotationsFromPb(node)
	if err != nil {
		return nil, err
	}

	metrics, err := newMetricsFromPb(node)
	if err != nil {
		return nil, err
	}

	return &Node{
		ID:              myself.ID,
		IsEmpty:         node.IsEmpty,
		MySelf:          myself,
		Nodes:           nodes,
		HostAnnotations: annotations,
		Metrics:         metrics,
	}, nil
}

func (n *Node) setObservationTime() {
	n.LastObservation = time.Now().UTC()
}
