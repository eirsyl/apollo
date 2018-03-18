package orchestrator

import (
	"sort"
	"strings"
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
