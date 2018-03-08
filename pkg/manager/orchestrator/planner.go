package orchestrator

import (
	"time"

	"sync"

	"github.com/satori/go.uuid"
)

type plannerStatus int
type taskType int

var (
	statusWaiting   plannerStatus = 1
	statusExecuting plannerStatus = 2
	statusExecuted  plannerStatus = 3

	taskCreateCluster     taskType = 1
	taskCheckCluster      taskType = 2
	taskFixSlotAllocation taskType = 3
	taskRebalanceCluster  taskType = 4
	taskAddNodeCluster    taskType = 5
	taskRemoveNodeCluster taskType = 6
)

type command struct {
	ID           uuid.UUID
	OrderKey     int64
	NodeId       string
	Status       plannerStatus
	Creation     time.Time
	Execution    time.Time
	Dependencies []uuid.UUID
}

type commandResult struct {
}

// task represents a task that the cluster manager should execute
type task struct {
	Type     taskType
	Status   plannerStatus
	Commands []*command
}

func (t *task) getNextCommand(nodeId string) (*command, error) {
	return nil, nil
}

// planner is responsible for storing the plans that the cluster manager plans to
// execute.
type planner struct {
	lock  sync.Mutex
	tasks []task
}

func newPlanner() (*planner, error) {
	return &planner{}, nil
}

func (p *planner) currentTask() (*task, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, task := range p.tasks {
		if task.Status == statusWaiting || task.Status == statusExecuting {
			return &task, nil
		}
	}

	return nil, nil
}

func (p *planner) nextCommand(nodeId string) (*command, error) {
	task, err := p.currentTask()
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, nil
	}

	command, err := task.getNextCommand(nodeId)
	if err != nil {
		return nil, err
	}

	return command, nil
}

func (p *planner) reportResult(nodeId string, result *commandResult) error {
	return nil
}

/**
 * Task creators
 */
func (p *planner) newCreateClusterTask() {
	p.lock.Lock()
	defer p.lock.Unlock()

	createClusterTask := task{
		Type: taskCreateCluster,
	}

	p.tasks = append(p.tasks, createClusterTask)
}
