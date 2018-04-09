package planner

import (
	"strconv"
	"strings"
)

// SlotKeyCounts contains information about slot key counts from nodes
type SlotKeyCounts struct {
	Counts map[int]int
}

// NewSlotKeyCounts creates a new SlotKeyCounts
func NewSlotKeyCounts(result []string) SlotKeyCounts {
	counts := map[int]int{}

	for _, count := range result {
		split := strings.Split(count, "=")
		if len(split) != 2 {
			continue
		}
		slot, err := strconv.Atoi(split[0])
		if err != nil {
			continue
		}
		c, err := strconv.Atoi(split[1])
		if err != nil {
			continue
		}
		counts[slot] = c
	}
	return SlotKeyCounts{Counts: counts}
}

// AllocationResult stores slot allocation decisions
type AllocationResult struct {
	Slots []int
}

// AdvancedAllocationResult stores slot allocation decisions
type AdvancedAllocationResult struct {
}

type IsMasterPlanner interface {
	IsMasterNode(nodeID string) (bool, error)
}

// SlotCoveragePlanner defined an interface used to decide where to store cluster slots
type SlotCoveragePlanner interface {
	IsMasterPlanner
	AllocateSlotsWithoutKeys(counts map[string]SlotKeyCounts) (map[string]AllocationResult, error)
	AllocateSlotsWithOneNode(counts map[string]SlotKeyCounts) (map[string]AllocationResult, error)
	AllocateSlotsWithMultipleNodes(counts map[string]SlotKeyCounts) (map[string]AdvancedAllocationResult, error)
}

type SlotCloserPlanner interface {
	IsMasterPlanner
	SlotOwners(slot int) ([]string, error)
	MigratingNodes(slot int) ([]string, error)
	ImportingNodes(slot int) ([]string, error)
	GetAddr(nodeID string) (string, error)
}
