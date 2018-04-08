package orchestrator

import (
	"math"

	"errors"
	"math/rand"

	"github.com/eirsyl/apollo/pkg"
	"github.com/eirsyl/apollo/pkg/manager/orchestrator/planner"
	log "github.com/sirupsen/logrus"
)

/**
 * allocSlots uses the information provided by each node to assign slots and replica configuration
 * TODO: This is a simple implementation, a more advanced implementation based on annotations is required.
 * TODO: Node state, memory size etc should also be included in slot allocation.
 */
func allocSlots(nodes *[]Node, replication int) (map[string]*planner.CreateClusterNodeOpts, error) {
	config := map[string]*planner.CreateClusterNodeOpts{}
	nodesCount := len(*nodes)

	masters := nodesCount / (replication + 1)
	masterNodes := (*nodes)[0:masters]

	slotsPerNode := pkg.ClusterHashSlots / masters
	first := 0
	cursor := 0.0

	for i := 0; i < nodesCount; i++ {
		if i < masters {
			// Slot assignment
			node := (*nodes)[i]

			last := int(math.Round(cursor + float64(slotsPerNode) - 1))

			if last > pkg.ClusterHashSlots || i == masters-1 {
				last = pkg.ClusterHashSlots - 1
			}
			if last < first {
				last = first
			}

			var slots []int
			for s := first; s <= last; s++ {
				slots = append(slots, s)
			}

			first = last + 1
			cursor += float64(slotsPerNode)

			config[node.ID] = &planner.CreateClusterNodeOpts{Slots: slots, Addr: node.Addr}
		} else {
			// Replication assignment
			replicationTarget := masterNodes[i%masters]
			node := (*nodes)[i]
			config[node.ID] = &planner.CreateClusterNodeOpts{ReplicationTarget: replicationTarget.ID, Addr: node.Addr}
		}
	}

	return config, nil
}

// allocEmptySlot is responsible for allocation an empty slot on a master inside the cluster
// TODO: Base this decision on resource limits
func allocEmptySlot(nodes *[]Node) (*Node, error) {
	var masters []Node
	for _, node := range *nodes {
		if node.MySelf.Role == "master" {
			masters = append(masters, node)
		}
	}
	if len(masters) == 0 {
		return nil, errors.New("no masters found")
	}
	n := rand.Int() % len(masters)
	return &masters[n], nil
}

type slotCoveragePlanner struct {
	nm *nodeManager
}

func newSlotCoveragePlanner(nm *nodeManager) (*slotCoveragePlanner, error) {
	return &slotCoveragePlanner{nm: nm}, nil
}

func (scp *slotCoveragePlanner) convertCounts(counts map[string]planner.SlotKeyCounts) map[int][]string {
	res := map[int][]string{}
	for node, c := range counts {
		for slot, keys := range c.Counts {
			nodes, ok := res[slot]
			if ok {
				if keys > 0 {
					res[slot] = append(nodes, node)
				}
			} else {
				if keys > 0 {
					res[slot] = []string{node}
				} else {
					res[slot] = []string{}
				}
			}
		}
	}
	return res
}

func (scp *slotCoveragePlanner) AllocateSlotsWithoutKeys(counts map[string]planner.SlotKeyCounts) (map[string]planner.AllocationResult, error) {
	nodeSlots := scp.convertCounts(counts)
	var emptySlots []int
	for slot, nodes := range nodeSlots {
		if len(nodes) == 0 {
			emptySlots = append(emptySlots, slot)
		}
	}

	// Assign slots to random masters
	res := map[string]planner.AllocationResult{}
	clusterNodes, err := scp.nm.getClusterNodes()
	if err != nil {
		return nil, err
	}
	allNodes, err := scp.nm.allNodes()
	if err != nil {
		return nil, err
	}
	var nodes []Node
	for _, node := range clusterNodes {
		n, ok := allNodes[node]
		if ok {
			nodes = append(nodes, n)
		}
	}

	for _, slot := range emptySlots {
		node, err := allocEmptySlot(&nodes)
		if err != nil {
			return nil, err
		}

		ar, ok := res[node.ID]
		if ok {
			ar.Slots = append(ar.Slots, slot)
		} else {
			res[node.ID] = planner.AllocationResult{Slots: []int{slot}}
		}
	}

	return res, nil
}

func (scp *slotCoveragePlanner) AllocateSlotsWithOneNode(counts map[string]planner.SlotKeyCounts) (map[string]planner.AllocationResult, error) {
	nodeSlots := scp.convertCounts(counts)
	slotsWithOneNode := map[string][]int{}
	for slot, nodes := range nodeSlots {
		if len(nodes) == 1 {
			node := nodes[0]
			slots, ok := slotsWithOneNode[node]
			if ok {
				slotsWithOneNode[node] = append(slots, slot)
			} else {
				slotsWithOneNode[node] = []int{slot}
			}
		}
	}

	// Assign slots to the node with keys
	res := map[string]planner.AllocationResult{}
	for node, slots := range slotsWithOneNode {
		res[node] = planner.AllocationResult{
			Slots: slots,
		}
	}

	return res, nil
}

func (scp *slotCoveragePlanner) AllocateSlotsWithMultipleNodes(counts map[string]planner.SlotKeyCounts) (map[string]planner.AdvancedAllocationResult, error) {
	// TODO: Implement this
	nodeSlots := scp.convertCounts(counts)
	var slotsWithMultipleNodes []int
	for slot, nodes := range nodeSlots {
		if len(nodes) > 1 {
			slotsWithMultipleNodes = append(slotsWithMultipleNodes, slot)
		}
	}
	log.Info("[WIP] Slots with multiple nodes: %v", slotsWithMultipleNodes)
	return map[string]planner.AdvancedAllocationResult{}, nil
}

func (scp *slotCoveragePlanner) IsMasterNode(nodeID string) (bool, error) {
	node, err := scp.nm.getNode(nodeID)
	if err != nil {
		return false, err
	}
	return node.MySelf.Role == "master", nil
}
