package orchestrator

// Instance represents an member of the cluster.
type Instance struct {
	Name string
}

// NewInstance is a shortcut for creating a new instance struct.
func NewInstance(name string) (*Instance, error) {
	return &Instance{Name: name}, nil
}
