package contrib

// Command describes commands that the manager can issue to agents
//go:generate stringer -type=Command
type Command int

const (
	// CommandAddSlots is used to assign slots to a node
	CommandAddSlots Command = iota + 1
	// CommandSetReplicate represents the command used to set replication on a slave
	CommandSetReplicate
	// CommandSetEpoch is used to set the cluster epoch
	CommandSetEpoch
	// CommandJoinCluster is used to call the CLUSTER MEET command.
	CommandJoinCluster
	// CommandCountKeysInSlots is used to count keys given a list of slots.
	CommandCountKeysInSlots
	// CommandSetSlotState updates slot state
	CommandSetSlotState
	// CommandBumpEpoch increments cluster epoch
	CommandBumpEpoch
	// CommandDelSlots deletes a slot from the node
	CommandDelSlots
	// CommandMigrateSlots migrates a slot from one node to another
	CommandMigrateSlots
)

// NodeCommand stores a command that the manager issues to an agent
type NodeCommand struct {
	ID        string
	Command   Command
	Arguments []string
}

// NewNodeCommand is used to create a new node command from the protobuf response
func NewNodeCommand(id string, command int64, arguments []string) (*NodeCommand, error) {
	var commandType Command
	{
		commandType = Command(command)
	}

	return &NodeCommand{
		ID:        id,
		Command:   commandType,
		Arguments: arguments,
	}, nil
}

// NodeCommandResult stores the result of a command execution, used to report result to the manager
type NodeCommandResult struct {
	ID      string
	Result  []string
	Success bool
}

// NewNodeCommandResult creates a new NodeCommandResult
func NewNodeCommandResult(id string, result []string, success bool) (*NodeCommandResult, error) {
	return &NodeCommandResult{
		ID:      id,
		Result:  result,
		Success: success,
	}, nil
}
