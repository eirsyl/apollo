package pkg

const (
	// GRPCPortWindow describes the range between the redis server port
	// and the agent port.
	// The redis cluster protocol uses a window of 10000
	// Redis Port = 6379, Redis Cluster = 16379, Apollo Agent = 7379
	GRPCPortWindow = 1000

	// HTTPPortWindow is the offset between the redis server port and the
	// http debug server.
	HTTPPortWindow = 2000

	// ClusterHashSlots is the number of hash slots in a redis cluster
	ClusterHashSlots = 16384
)
