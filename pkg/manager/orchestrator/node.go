package orchestrator

import (
	"time"

	"regexp"

	"strconv"

	"errors"

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

type slotAction int

var (
	importing slotAction = 1
	migrating slotAction = 2
)

func stringToSlotAction(action string) slotAction {
	if action == "<" {
		return importing
	}
	return migrating
}

type openSlot struct {
	slot        int
	action      slotAction
	destination string
}

// openSlots returns a map of open slots, the status and destination node
func (cn *ClusterNeighbour) openSlots() (*[]openSlot, error) {
	var slots []openSlot
	reg := regexp.MustCompile(`^\[(?P<slot>\d+)\-(?P<operation>[<|>])\-(?P<destination>\w+)\]$`)
	for _, slot := range cn.Slots {
		match := reg.FindStringSubmatch(slot)
		if len(match) == 4 {
			slot, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, err
			}
			slots = append(slots, openSlot{
				slot:        slot,
				action:      stringToSlotAction(match[2]),
				destination: match[3],
			})
		}
	}
	return &slots, nil
}

// allSlots returns a list of slots managed by the node
func (cn *ClusterNeighbour) allSlots() ([]int, error) {
	slotRange := regexp.MustCompile(`^(?P<from>\d+)\-(?P<to>\d+)$`)
	slotSingle := regexp.MustCompile(`^(?P<slot>\d+)$`)
	var slots []int

	for _, s := range cn.Slots {
		match := slotRange.FindStringSubmatch(s)
		if len(match) == 3 {
			from, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, err
			}
			to, err := strconv.Atoi(match[2])
			if err != nil {
				return nil, err
			}
			if to < from {
				return nil, errors.New("from has to be lower than to")
			}
			for i := from; i <= to; i++ {
				slots = append(slots, i)
			}
			continue
		}
		match = slotSingle.FindStringSubmatch(s)
		if len(match) == 2 {
			i, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, err
			}
			slots = append(slots, i)
			continue
		}
	}
	return slots, nil
}

// Node represents an member of the cluster.
type Node struct {
	ID                string                      `json:"id"`
	Addr              string                      `json:"addr"`
	IsEmpty           bool                        `json:"isEmpty"`
	MySelf            ClusterNeighbour            `json:"myself"`
	Nodes             map[string]ClusterNeighbour `json:"nodes"`
	HostAnnotations   map[string]string           `json:"annotations"`
	Metrics           map[string]float64          `json:"metrics"`
	LastObservation   time.Time                   `json:"lastObservation"`
	MarkedForDeletion bool                        `json:"markedForDeletion"`
}

func newClusterNeighbourFromPb(nodes []*pb.StateRequest_ClusterNode) (map[string]ClusterNeighbour, ClusterNeighbour, error) {
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
		Addr:            node.Addr,
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
