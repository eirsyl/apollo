package redis

// ClusterOperations is responsible for cluster related tasks that each agent need
// in order to execute cluster changes.
type ClusterOperations struct {
	redis *Client
}

// NewClusterOperations is responsible for creating a new operations instance
func NewClusterOperations(redis *Client) (*ClusterOperations, error) {
	co := ClusterOperations{redis: redis}
	return &co, nil
}

// isEmpty
// The node is empty if no database exists and the instance is'nt aware of other cluster members
func isEmpty() (bool, error) {
	// info keyspace !contains db0 && cluster info contains cluster_known_nodes:1
	return true, nil
}
