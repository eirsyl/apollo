package orchestrator

// Node represents an member of the cluster.
type Node struct {
	Name string
}

// NewNode is a shortcut for creating a new node struct.
func NewNode(name string) (*Node, error) {
	return &Node{Name: name}, nil
}
