package planner

import (
	log "github.com/sirupsen/logrus"
)

type taskType int

// Int64 returns the value as a int64
func (cs *taskType) Int64() int64 {
	return int64(*cs)
}

type taskStatus int

// Int64 returns the value as a int64
func (cs *taskStatus) Int64() int64 {
	return int64(*cs)
}

var (
	// TaskCreateCluster task is responsible for creating a cluster
	TaskCreateCluster taskType = 1
	// TaskCheckCluster is responsible for checking the configuration of a cluster
	TaskCheckCluster taskType = 2
	// TaskFixSlotAllocation is responsible for fixing slot allocation issues
	TaskFixSlotAllocation taskType = 3
	// TaskRebalanceCluster is responsible for moving slot assignments in order to prevent cluster issues
	TaskRebalanceCluster taskType = 4
	// TaskAddNodeCluster is used to add new nodes to the cluster
	TaskAddNodeCluster taskType = 5
	// TaskRemoveNodeCluster is used to remove a node from the cluster
	TaskRemoveNodeCluster taskType = 6

	// StatusWaiting is used for a task that is in the execution queue
	StatusWaiting taskStatus
	// StatusExecuting is used by tasks that is currently being executed by the nodes.
	StatusExecuting taskStatus = 1
	// StatusExecuted is used on tasks that is done
	StatusExecuted taskStatus = 2
)

// Task represents a task that the cluster manager should execute
type Task struct {
	Type     taskType
	Status   taskStatus
	Commands []*Command
}

// NewTask creates a new task
func NewTask(t taskType, commands []*Command) (*Task, error) {
	task := &Task{
		Type:     t,
		Status:   StatusWaiting,
		Commands: commands,
	}
	return task, nil
}

// NextCommands returns the next command to be executed by a node with the provided NodeID
func (t *Task) NextCommands(nodeID string) ([]*Command, error) {
	commands := []*Command{}

	for _, command := range t.Commands {
		if command.NodeID != nodeID {
			continue
		}

		if command.Status == CommandRunning {
			commands = append(commands, command)
			continue
		}

		if command.Status == CommandWaiting {
			if len(command.Dependencies) == 0 {
				commands = append(commands, command)
			} else {
				available := true
				for _, c := range command.Dependencies {
					if c.Status != CommandFinished {
						available = false
						break
					}
				}
				if available {
					commands = append(commands, command)
				}
			}
			continue
		}
	}
	return commands, nil
}

// UpdateStatus checks the task state based on the command states
func (t *Task) UpdateStatus() {
	status := StatusWaiting
	allFinished := true

L:
	for _, command := range t.Commands {
		switch command.Status {
		case CommandWaiting:
			allFinished = false
			continue
		case CommandRunning:
			allFinished = false
			status = StatusExecuting
			break L
		}
	}

	if allFinished {
		status = StatusExecuted
	}

	if status == StatusWaiting && t.Status != StatusWaiting {
		// Don't change the status back to waiting if marked as started.
		return
	}

	if status != t.Status {
		log.Infof("Updating task status: %v %v", t.Type, status)
		t.Status = status
	}
}
