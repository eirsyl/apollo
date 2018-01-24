package chaos

// Chaos exports the chaos struct used to operate an chaos.
type Chaos struct {
}

// NewChaos initializes a new chaos instance and returns a pinter to it.
func NewChaos() (*Chaos, error) {
	return &Chaos{}, nil
}

// Run starts the Chaos
func (m *Chaos) Run() error {
	return nil
}

// Exit gracefully shuts down the chaos
func (m *Chaos) Exit() error {
	return nil
}
