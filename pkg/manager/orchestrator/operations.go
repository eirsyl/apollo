package orchestrator

// ClusterOperations is responsible for the manager orchestration of cluster changes.
type ClusterOperations struct {
}

// NewClusterOperations creates a new ClusterOperations instance
func NewClusterOperations() (*ClusterOperations, error) {
	return &ClusterOperations{}, nil
}

func createCluster(hosts []Agent) error {
	return nil
}

func check() error {
	return nil
}

func info() error {
	return nil
}

func fix() error {
	return nil
}

func reshard() error {
	return nil
}

func addNode() error {
	return nil
}

func deleteNode() error {
	return nil
}
