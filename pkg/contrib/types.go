package contrib

import "errors"

// Command describes commands that the manager can issue to agents
type Command int

const (
	// CommandAddSlots is used to assign slots to a node
	CommandAddSlots Command = 1
	// CommandSetReplicate represents the command used to set replication on a slave
	CommandSetReplicate Command = 2
	// CommandSetEpoch is used to set the cluster epoch
	CommandSetEpoch Command = 3
	// CommandJoinCluster is used to call the CLUSTER MEET command.
	CommandJoinCluster Command = 4
	// CommandCountKeysInSlots is used to count keys given a list of slots.
	CommandCountKeysInSlots Command = 5
	// CommandSetSlotState updates slot state
	CommandSetSlotState Command = 6
	// CommandBumpEpoch increments cluster epoch
	CommandBumpEpoch Command = 7
	// CommandDelSlots deletes a slot from the node
	CommandDelSlots Command = 8
	// CommandMigrateSlots migrates a slot from one node to another
	CommandMigrateSlots Command = 9
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
	switch command {
	case 1:
		commandType = CommandAddSlots
	case 2:
		commandType = CommandSetReplicate
	case 3:
		commandType = CommandSetEpoch
	case 4:
		commandType = CommandJoinCluster
	case 5:
		commandType = CommandCountKeysInSlots
	case 6:
		commandType = CommandSetSlotState
	case 7:
		commandType = CommandBumpEpoch
	case 8:
		commandType = CommandDelSlots
	case 9:
		commandType = CommandMigrateSlots
	default:
		return nil, errors.New("Unsupported command type")
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
