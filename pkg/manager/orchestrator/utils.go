package orchestrator

import (
	"sort"
	"strings"

	"github.com/eirsyl/apollo/pkg"
	"github.com/eirsyl/apollo/pkg/utils"
)

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

	return utils.MapKeysToString(members), len(memberSignatures) == 1, nil
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

	return utils.MapKeysToInt(missingSlots)
}
