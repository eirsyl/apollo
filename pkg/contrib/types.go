package contrib

// Command describes commands that the manager can issue to agents
type Command int

const (
	// AddSlotsCommand is used to assign slots to a node
	AddSlotsCommand Command = 1
)

// NodeCommand stores a command that the manager issues to an agent
type NodeCommand struct {
	id        string
	command   Command
	arguments []string
}

// NodeCommandResult stores the result of a command execution, used to report result to the manager
type NodeCommandResult struct {
	id     string
	result string
}
