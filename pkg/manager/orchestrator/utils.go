package orchestrator

import (
	"sort"
	"strings"

	"github.com/eirsyl/apollo/pkg"
)

// HumanizeClusterState returns a human readable string based on a clusterState
func HumanizeClusterState(state int64) string {
	switch clusterState(state) {
	case clusterUnknown:
		return "Unknown"
	case clusterUnconfigured:
		return "Unconfigured"
	case clusterConfigured:
		return "Configured"
	default:
		return "ClusterState_NOT_SET"
	}

}

// HumanizeClusterHealth returns a human readable string based on a clusterHealth
func HumanizeClusterHealth(health int64) string {
	switch clusterHealth(health) {
	case clusterOK:
		return "OK"
	case clusterWarn:
		return "Warning"
	case clusterError:
		return "Error"
	default:
		return "ClusterHealth_NOT_SET"
	}
}

// mapKeysToString extracts the keys from a map
func mapKeysToString(m map[string]bool) (res []string) {
	for key := range m {
		res = append(res, key)
	}
	return
}

// mapKeysToInt extracts the keys from a map
func mapKeysToInt(m map[int]bool) (res []int) {
	for key := range m {
		res = append(res, key)
	}
	return
}

// stringListToMap creates a map[string]bool based on a []string
func stringListToMap(list []string) (res map[string]bool) {
	res = make(map[string]bool)
	for _, element := range list {
		res[element] = true
	}
	return
}

// findClusterNodes parses a list of nodes and returns all cluster members reported by the given nodes
func findClusterNodes(nodes *[]Node) ([]string, bool, error) {
	members := map[string]bool{}
	memberSignatures := map[string]bool{}

	for _, node := range *nodes {
		var memberList []string
		for _, member := range node.Nodes {
			members[member.ID] = true
			memberList = append(memberList, member.ID)
		}
		sort.Strings(memberList)
		signature := strings.Join(memberList, "")
		memberSignatures[signature] = true
	}

	return mapKeysToString(members), len(memberSignatures) == 1, nil
}

// nodeSignatures groups nodes by signature, not the same ass findClusterNodes
func nodeSignatures(nodes *[]Node) (map[string][]Node, error) {
	signatures := map[string][]Node{}

	for _, node := range *nodes {
		var memberList []string
		for _, member := range node.Nodes {
			memberList = append(memberList, member.ID)
		}
		sort.Strings(memberList)
		signature := strings.Join(memberList, "")

		lookup, ok := signatures[signature]
		if ok {
			lookup = append(lookup, node)
			signatures[signature] = lookup
		} else {
			signatures[signature] = []Node{node}
		}

	}

	return signatures, nil
}

// findOpenSlots is used to find missing slots from a list of assigned slots
func findOpenSlots(slots []int) []int {
	missingSlots := map[int]bool{}
	for i := 0; i < pkg.ClusterHashSlots; i++ {
		missingSlots[i] = true
	}

	for _, i := range slots {
		delete(missingSlots, i)
	}

	return mapKeysToInt(missingSlots)
}
