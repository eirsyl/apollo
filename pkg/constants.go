package pkg

import (
	"time"
)

const (
	// HTTPPortWindow is the offset between the redis server port and the
	// http debug server.
	HTTPPortWindow = 2000

	// ClusterHashSlots is the number of hash slots in a redis cluster
	ClusterHashSlots = 16384

	// PrecheckWaitDelay defines the amount of time to wait before the agent exits
	// if the precheck failes
	PrecheckWaitDelay = 10 * time.Second
)

var (
	//AgentMetrics contains a list with the metrics required by the manager to calculate instance state
	AgentMetrics = map[string]bool{
		"uptime_in_seconds": true,
		"connected_clients": true,
	}
)
