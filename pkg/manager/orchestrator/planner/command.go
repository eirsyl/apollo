package planner

import (
	"time"

	"github.com/satori/go.uuid"
)

type commandStatus int
type commandType int

var (
	CommandWaiting commandStatus = 0

	CommandAddSlots     commandType = 1
	CommandSetReplicate commandType = 2
	CommandSetEpoch     commandType = 3
	CommandJoinCluster  commandType = 4
)

type CommandOpts struct {
	KeyString  map[string]string
	KeyIntList map[string][]int
}

func NewCommandOpts() *CommandOpts {
	return &CommandOpts{
		KeyString:  map[string]string{},
		KeyIntList: map[string][]int{},
	}
}

func (co *CommandOpts) AddKS(key, value string) {
	co.KeyString[key] = value
}

func (co *CommandOpts) AddKIL(key string, value []int) {
	co.KeyIntList[key] = value
}

// Command represents a command that should be executed on a node
type Command struct {
	ID           uuid.UUID
	NodeId       string
	Type         commandType
	Status       commandStatus
	Opts         CommandOpts
	Creation     time.Time
	Execution    time.Time
	Retries      int64
	Dependencies []*Command
}

// NewCommand creates a new command
func NewCommand(nodeId string, ct commandType, opts CommandOpts, dependencies []*Command) (*Command, error) {
	if dependencies == nil {
		dependencies = []*Command{}
	}
	return &Command{
		ID:           uuid.NewV4(),
		NodeId:       nodeId,
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
