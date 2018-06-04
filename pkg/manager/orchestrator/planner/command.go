package planner

import (
	"time"

	"github.com/satori/go.uuid"
)

// CommandStatus represents the status of the command type
//go:generate stringer -type=CommandStatus
type CommandStatus int

// Int64 returns the value as a int64
func (cs *CommandStatus) Int64() int64 {
	return int64(*cs)
}

// CommandType represents the type of the command type
//go:generate stringer -type=CommandType
type CommandType int

// Int64 returns the value as a int64
func (t *CommandType) Int64() int64 {
	return int64(*t)
}

const (
	// CommandWaiting is used before a node has tried to execute the command
	CommandWaiting CommandStatus = iota
	// CommandRunning represents a command that is currently being executed by a
	CommandRunning
	// CommandFinished represents a command that has terminated successfully
	CommandFinished
)
const (
	// CommandAddSlots represents the command that is responsible for adding slots to a master node.
	CommandAddSlots CommandType = iota + 1
	// CommandSetReplicate represents the command used to set replication on a slave
	CommandSetReplicate
	// CommandSetEpoch is used to set the cluster epoch
	CommandSetEpoch
	// CommandJoinCluster is used to call the CLUSTER MEET command.
	CommandJoinCluster
	// CommandCountKeysInSlots is used to count keys given a list of keys
	CommandCountKeysInSlots
	// CommandSetSlotState changes the state of a set of slots
	CommandSetSlotState
	// CommandBumpEpoch bumps the current cluster epoch
	CommandBumpEpoch
	// CommandDelSlots deletes a slot from a node
	CommandDelSlots
	// CommandMigrateSlots is responsible for migrating keys from one node to another
	CommandMigrateSlots
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

// GetKS is a helper function for retrieving values from the key string mapping
func (co *CommandOpts) GetKS(key string) string {
	return co.KeyString[key]
}

// GetKIL is helper function for retrieving a int list from the key int list
func (co *CommandOpts) GetKIL(key string) []int {
	return co.KeyIntList[key]
}

// Command represents a command that should be executed on a node
type Command struct {
	ID            uuid.UUID
	NodeID        string
	Type          CommandType
	Status        CommandStatus
	Opts          CommandOpts
	Creation      time.Time
	Execution     time.Time
	Retries       int64
	Dependencies  []*Command
	Results       [][]string
	ShouldExecute *func(command *Command) (bool, error)
}

// NewCommand creates a new command
func NewCommand(nodeID string, ct CommandType, opts CommandOpts, dependencies []*Command) (*Command, error) {
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
		Results:      [][]string{},
	}, nil
}

// UpdateStatus changes the command status
func (c *Command) UpdateStatus(status CommandStatus) {
	if c.Status == CommandRunning && status == CommandRunning {
		c.IncrementRetry()
	}

	c.Status = status

	if c.Status == CommandRunning {
		c.Execution = time.Now().UTC()
	}
}

// ReportResult is used to set the execution result of the task
func (c *Command) ReportResult(cr *CommandResult) {
	if cr.Success {
		c.Status = CommandFinished
	}
	c.Results = append(c.Results, cr.Result)
}

// IncrementRetry increments the retry counter
func (c *Command) IncrementRetry() {
	c.Retries++
}

// CommandResult stores a execution of a command on a node
type CommandResult struct {
	ID      uuid.UUID
	Result  []string
	Success bool
}

// NewCommandResult creates a new CommandResult
func NewCommandResult(id string, result []string, success bool) (*CommandResult, error) {
	uuidID, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}
	return &CommandResult{
		ID:      uuidID,
		Result:  result,
		Success: success,
	}, nil
}
