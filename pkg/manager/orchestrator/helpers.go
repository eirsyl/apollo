package orchestrator

import (
	"math"

	"github.com/eirsyl/apollo/pkg"
	"github.com/eirsyl/apollo/pkg/manager/orchestrator/planner"
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
