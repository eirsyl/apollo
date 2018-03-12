package planner

import (
	"time"

	"github.com/satori/go.uuid"
)

type commandStatus int
type commandType int

var (
	// CommandWaiting is used before a node has tried to execute the command
	CommandWaiting commandStatus

	// CommandAddSlots represents the command that is responsible for adding slots to a master node.
	CommandAddSlots commandType = 1
	// CommandSetReplicate represents the command used to set replication on a slave
	CommandSetReplicate commandType = 2
	// CommandSetEpoch is used to set the cluster epoch
	CommandSetEpoch commandType = 3
	// CommandJoinCluster is used to call the CLUSTER MEET command.
	CommandJoinCluster commandType = 4
)

// CommandOpts is used to attach extra data to commands.
type CommandOpts struct {
	KeyString  map[string]string
	KeyIntList map[string][]int
}

// NewCommandOpts creates a new CommandOpts used to pass extra data to commands.
func NewCommandOpts() *CommandOpts {
	return &CommandOpts{
		KeyString:  map[string]string{},
		KeyIntList: map[string][]int{},
	}
}

// AddKS is used to attach a string to a key in the key value mapping.
func (co *CommandOpts) AddKS(key, value string) {
	co.KeyString[key] = value
}

// AddKIL is used to add a list of integers to a key in the key to int list mapping.
func (co *CommandOpts) AddKIL(key string, value []int) {
	co.KeyIntList[key] = value
}

// Command represents a command that should be executed on a node
type Command struct {
	ID           uuid.UUID
	NodeID       string
	Type         commandType
	Status       commandStatus
	Opts         CommandOpts
	Creation     time.Time
	Execution    time.Time
	Retries      int64
	Dependencies []*Command
}

// NewCommand creates a new command
func NewCommand(nodeID string, ct commandType, opts CommandOpts, dependencies []*Command) (*Command, error) {
	if dependencies == nil {
		dependencies = []*Command{}
	}
	return &Command{
		ID:           uuid.NewV4(),
		NodeID:       nodeID,
		Type:         ct,
		Opts:         opts,
		Status:       CommandWaiting,
		Creation:     time.Now().UTC(),
		Dependencies: dependencies,
	}, nil
}

// CommandResult stores a execution of a command on a node
type CommandResult struct {
}
