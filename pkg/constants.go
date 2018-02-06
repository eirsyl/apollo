package pkg

const (
	// PortWindow describes the range between the redis server port and the agent port.
	// The redis cluster protocol uses a window of 10000
	// Redis Port = 6379, Redis Cluster = 16379, Apollo Agent = 21379
	PortWindow = 15000
)
