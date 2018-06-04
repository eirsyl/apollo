package planner

import (
	"sync"

	"time"

	log "github.com/sirupsen/logrus"
)

// TaskType represents the type of a task
//go:generate stringer -type=TaskType
type TaskType int

// Int64 returns the value as a int64
func (cs *TaskType) Int64() int64 {
	return int64(*cs)
}

// TaskStatus represents the status of a task
//go:generate stringer -type=TaskStatus
type TaskStatus int

// Int64 returns the value as a int64
func (cs *TaskStatus) Int64() int64 {
	return int64(*cs)
}

const (
	// TaskCreateCluster task is responsible for creating a cluster
	TaskCreateCluster TaskType = iota + 1
	// TaskMemberFixup is responsible for making sure each cluster member knows about each other
	TaskMemberFixup
	// TaskFixOpenSlots is responsible for fixing open slots
	TaskFixOpenSlots
	// TaskFixSlotAllocation is responsible for fixing slot allocation issues
	TaskFixSlotAllocation
	// TaskAddNodeCluster is used to add new nodes to the cluster
	TaskAddNodeCluster
	// TaskRemoveNodeCluster is used to remove a node from the cluster
	TaskRemoveNodeCluster
	// TaskReshardCluster is used to reshard the cluster
	TaskReshardCluster
)
const (
	// StatusWaiting is used for a task that is in the execution queue
	StatusWaiting TaskStatus = iota
	// StatusExecuting is used by tasks that is currently being executed by the nodes.
	StatusExecuting
	// StatusExecuted is used on tasks that is done
	StatusExecuted
)

// Task represents a task that the cluster manager should execute
type Task struct {
	Type           TaskType
	Status         TaskStatus
	Commands       []*Command
	ProcessResults func(task *Task) error
	lock           sync.Mutex
	startTime      time.Time
}

// NewTask creates a new task
func NewTask(t TaskType, commands []*Command, startTime time.Time) (*Task, error) {
	task := &Task{
		Type:      t,
		Status:    StatusWaiting,
		Commands:  commands,
		lock:      sync.Mutex{},
		startTime: startTime,
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
				if !shouldExecute(command) {
					continue
				}
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
					if !shouldExecute(command) {
						continue
					}
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

	// Internal task processing used to update task commands
	if t.ProcessResults != nil {
		t.lock.Lock()
		err := t.ProcessResults(t)
		t.lock.Unlock()
		if err != nil {
			log.Warnf("Could not run internal task processing: %v", err)
			t.Status = StatusExecuted
			return
		}
	}

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
		if t.Status == StatusExecuted {
			log.WithFields(
				log.Fields{
					"duration":  time.Since(t.startTime),
					"operation": t.Type,
				},
			).Info("Task completed successfully")
		}
	}
}
