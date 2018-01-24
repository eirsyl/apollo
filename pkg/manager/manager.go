package manager

// Manager exports the manager struct used to operate an manager.
type Manager struct {
}

// NewManager initializes a new manager instance and returns a pinter to it.
func NewManager() (*Manager, error) {
	return &Manager{}, nil
}

// Run starts the manager
func (m *Manager) Run() error {
	return nil
}

// Exit gracefully shuts down the manager
func (m *Manager) Exit() error {
	return nil
}
