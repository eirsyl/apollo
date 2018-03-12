package planner

import "sync"

// Planner is responsible for storing the plans that the cluster manager plans to
// execute.
type Planner struct {
	lock  sync.Mutex
	tasks []*Task
}

// NewPlanner creates a new planner instance
func NewPlanner() (*Planner, error) {
	return &Planner{}, nil
}

// CurrentTask returns the current task performed by the nodes
func (p *Planner) CurrentTask() (*Task, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, task := range p.tasks {
		if task.Status == StatusWaiting || task.Status == StatusExecuting {
			return task, nil
		}
	}

	return nil, nil
}

// NextCommand sends the next command to be executed by a node
func (p *Planner) NextCommand(nodeID string) (*Command, error) {
	task, err := p.CurrentTask()
	if err != nil {
		return nil, err
	}

	if task == nil {
		return nil, nil
	}

	command, err := task.NextCommand(nodeID)
	if err != nil {
		return nil, err
	}

	return command, nil
}

// ReportResult saves a execution result from a node
func (p *Planner) ReportResult(nodeID string, result *CommandResult) error {
	return nil
}
