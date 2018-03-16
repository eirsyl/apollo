package orchestrator

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
		return "NOT_SET"
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
		return "NOT_SET"
	}
}
