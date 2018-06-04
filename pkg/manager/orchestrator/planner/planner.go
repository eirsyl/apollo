package planner

import (
	"sync"

	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

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

// NextCommands sends the next command to be executed by a node
func (p *Planner) NextCommands(nodeID string) (*Task, []*Command, error) {
	task, err := p.CurrentTask()
	if err != nil {
		return nil, nil, err
	}

	if task == nil {
		return nil, nil, nil
	}

	commands, err := task.NextCommands(nodeID)
	if err != nil {
		return nil, nil, err
	}

	return task, commands, nil
}

// ReportResult saves a execution result from a node
func (p *Planner) ReportResult(nodeID string, results []*CommandResult) error {
	cr := map[uuid.UUID]*CommandResult{}
	for _, result := range results {
		cr[result.ID] = result
	}

	p.lock.Lock()
	for _, task := range p.tasks {
		if task.Status == StatusWaiting || task.Status == StatusExecuted {
			log.Infof("Skipping task, task not active: %v", task.Type)
			continue
		}

		for _, command := range task.Commands {
			result, ok := cr[command.ID]
			if !ok {
				continue
			}

			command.ReportResult(result)
			delete(cr, command.ID)
		}

		task.UpdateStatus()
	}
	p.lock.Unlock()

	for commandID := range cr {
		log.Warnf("Unable to store command result for task: %v", commandID)
	}

	return nil
}

// StatusExport produces planner metrics to the protobuf interface (cli usage)
func (p *Planner) StatusExport() ([]*pb.StatusResponse_Task, error) {
	var tasks []*pb.StatusResponse_Task

	p.lock.Lock()
	for id, task := range p.tasks {
		var commands []*pb.StatusResponse_Task_Command
		for _, command := range task.Commands {
			commands = append(commands, &pb.StatusResponse_Task_Command{
				Id:           command.ID.String(),
				Status:       command.Status.Int64(),
				Type:         command.Type.Int64(),
				NodeID:       command.NodeID,
				Retries:      command.Retries,
				Dependencies: int64(len(command.Dependencies)),
			})
		}
		tasks = append(tasks, &pb.StatusResponse_Task{
			Id:       int64(id),
			Status:   task.Status.Int64(),
			Type:     task.Type.Int64(),
			Commands: commands,
		})
	}
	p.lock.Unlock()

	return tasks, nil
}
