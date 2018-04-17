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

func (scp *slotCoveragePlanner) AllocateSlotsWithMultipleNodes(counts map[string]planner.SlotKeyCounts) (map[int]planner.AdvancedAllocationResult, error) {
	/**
	* Steps:
	* 1. Find node with most keys within the actual slot
	* 2. Target addslot
	* 3. Target set slot stable
	* 4. For all other nodes with keys
	*	 * Set slot to importing
	*    * Move keys from source to target
	*    * Set slot to stable
	 */
	nodeSlots := scp.convertCounts(counts)
	var slotsWithMultipleNodes []int
	res := map[int]planner.AdvancedAllocationResult{}
	for slot, nodes := range nodeSlots {
		if len(nodes) > 1 {
			slotsWithMultipleNodes = append(slotsWithMultipleNodes, slot)
			var maxCount int
			var master string

			for _, node := range nodes {
				count, ok := counts[node]
				if ok {
					if master == "" || count.Counts[slot] > maxCount {
						maxCount = count.Counts[slot]
						master = node
					}
				}
			}

			res[slot] = planner.AdvancedAllocationResult{
				Nodes:  nodes,
				Master: master,
			}
		}
	}
	log.Info("Slots with multiple nodes: %v", slotsWithMultipleNodes)
	return res, nil
}

func (scp *slotCoveragePlanner) IsMasterNode(nodeID string) (bool, error) {
	node, err := scp.nm.getNode(nodeID)
	if err != nil {
		return false, err
	}
	return node.MySelf.Role == "master", nil
}

func (scp *slotCoveragePlanner) GetAddr(nodeID string) (string, error) {
	node, err := scp.nm.getNode(nodeID)
	if err != nil {
		return "", err
	}
	return node.Addr, nil
}

type slotClosePlanner struct {
	nm *nodeManager
}

func newSlotClosePlanner(nm *nodeManager) (*slotClosePlanner, error) {
	return &slotClosePlanner{nm: nm}, nil
}

func (scp *slotClosePlanner) clusterNodes() ([]Node, error) {
	var res []Node
	clusterNodes, err := scp.nm.getClusterNodes()
	if err != nil {
		return nil, err
	}
	nodes, err := scp.nm.allNodes()
	if err != nil {
		return nil, err
	}

	clusterNodeMap := map[string]bool{}
	for _, node := range clusterNodes {
		clusterNodeMap[node] = true
	}

	for _, node := range nodes {
		if clusterNodeMap[node.ID] {
			res = append(res, node)
		}
	}

	return res, nil
}

func (scp *slotClosePlanner) SlotOwners(slot int) ([]string, error) {
	var owners []string

	clusterNodes, err := scp.clusterNodes()
	if err != nil {
		return nil, err
	}

	for _, node := range clusterNodes {
		if node.MySelf.Role == "master" {
			slots, err := node.MySelf.allSlots()
			if err != nil {
				return nil, err
			}
			for _, s := range slots {
				if s == slot {
					owners = append(owners, node.ID)
				}
			}
		}
	}

	return owners, nil
}

func (scp *slotClosePlanner) MigratingNodes(slot int) ([]string, error) {
	var res []string

	clusterNodes, err := scp.clusterNodes()
	if err != nil {
		return nil, err
	}

	for _, node := range clusterNodes {
		if node.MySelf.Role == "master" {
			openSlots, err := node.MySelf.openSlots()
			if err != nil {
				return nil, err
			}

			for _, s := range *openSlots {
				if slot == s.slot && s.action == migrating {
					res = append(res, node.ID)
				}
			}
		}
	}

	return res, nil
}

func (scp *slotClosePlanner) ImportingNodes(slot int) ([]string, error) {
	var res []string

	clusterNodes, err := scp.clusterNodes()
	if err != nil {
		return nil, err
	}

	for _, node := range clusterNodes {
		if node.MySelf.Role == "master" {
			openSlots, err := node.MySelf.openSlots()
			if err != nil {
				return nil, err
			}

			for _, s := range *openSlots {
				if slot == s.slot && s.action == importing {
					res = append(res, node.ID)
				}
			}
		}
	}

	return res, nil
}

func (scp *slotClosePlanner) IsMasterNode(nodeID string) (bool, error) {
	node, err := scp.nm.getNode(nodeID)
	if err != nil {
		return false, err
	}
	return node.MySelf.Role == "master", nil
}

func (scp *slotClosePlanner) GetAddr(nodeID string) (string, error) {
	node, err := scp.nm.getNode(nodeID)
	if err != nil {
		return "", err
	}
	return node.Addr, nil
}

type addNodePlanner struct {
	nm *nodeManager
}

func newAddNodePlanner(nm *nodeManager) (*addNodePlanner, error) {
	return &addNodePlanner{
		nm: nm,
	}, nil
}

func (anp *addNodePlanner) GetNodePlans(nodes []string, replication int) ([]planner.AddNodePlan, error) {
	nodeMap, err := anp.nm.allNodes()
	if err != nil {
		return nil, err
	}
	members, err := anp.nm.getClusterNodes()
	if err != nil {
		return nil, err
	}

	masters := map[string]int{}

	// Construct master -> replica table
	for _, member := range members {
		memberNode, ok := nodeMap[member]
		if !ok {
			continue
		}
		if memberNode.MarkedForDeletion {
			continue
		}

		if memberNode.MySelf.Role == "master" {
			_, ok := masters[memberNode.ID]
			if !ok {
				masters[memberNode.ID] = 0
			}
		} else {
			_, ok := masters[memberNode.MySelf.MasterID]
			if !ok {
				masters[memberNode.MySelf.MasterID] = 1
			} else {
				masters[memberNode.MySelf.MasterID] += 1
			}
		}
	}

	var plans []planner.AddNodePlan

	// Assign node roles
	for _, node := range nodes {
		var mastersWithoutReplicas []string
		for master, replicas := range masters {
			if replicas < replication {
				log.Infof("Master has less replicas than required. Node %s has %d but %d is required.", master, replicas, replication)
				mastersWithoutReplicas = append(mastersWithoutReplicas, master)
			}
		}
		if len(mastersWithoutReplicas) == 0 {
			// Assign master-role to the node
			log.Infof("Add node: %s master", node)
			plans = append(plans, planner.AddNodePlan{
				NodeID:   node,
				IsMaster: true,
			})
			masters[node] = 0
		} else {
			// Pick random node to replicate
			log.Infof("Add node: %s replica", node)
			replicationTargetIndex := rand.Int() % len(mastersWithoutReplicas)
			replicationTarget := mastersWithoutReplicas[replicationTargetIndex]
			plans = append(plans, planner.AddNodePlan{
				NodeID:            node,
				IsMaster:          false,
				ReplicationTarget: replicationTarget,
			})
			masters[replicationTarget] += 1
		}
	}

	return plans, nil
}

func (anp *addNodePlanner) GetNodeToJoin() (string, error) {
	clusterMembers, err := anp.nm.getClusterNodes()
	if err != nil {
		return "", nil
	}

	if len(clusterMembers) > 0 {
		nodeID := clusterMembers[0]
		node, err := anp.nm.getNode(nodeID)
		if err != nil {
			return "", err
		}
		return node.Addr, nil
	}

	return "", errors.New("no existing cluster members")
}
