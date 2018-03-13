package orchestrator

/**
 * Validate the number of cluster members based on the replica setting
 */
func validateClusterSize(nodes, replicas int) bool {
	masters := nodes / (replicas + 1)
	return masters >= 3
}
