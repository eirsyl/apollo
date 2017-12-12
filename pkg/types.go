package pkg

// ManagerConfig describes the basic configuration a
// cluster manager needs to start.
type ManagerConfig struct {
	Name string `json:"name"`
}

// Manager interface describes the basic functions
// needed to implement a cluster manager. The functions
// exposed by this interface is mainly used by the CLI
// when starting or stopping the cluster manager.
type Manager interface {
	Start() error
	Reload(*ManagerConfig) error
	Stop() error
}
