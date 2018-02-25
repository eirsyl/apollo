package contrib

// NodeCommand describes commands that the manager can issue to agents
type NodeCommand int

const (
	// AddSlotsCommand is used to assign slots to a node
	AddSlotsCommand NodeCommand = 1
)
