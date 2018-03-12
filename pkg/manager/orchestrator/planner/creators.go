package planner

import (
	log "github.com/sirupsen/logrus"
)

// CreateClusterNodeOpts is used to provide extra data to the NewCreateClusterTask func.
// A struct is ude
type CreateClusterNodeOpts struct {
	Slots             []int
	ReplicationTarget string
	Addr              string
}

// NewCreateClusterTask creates a new cluster creation task. This function is responsible for creating
// the commands required to form a cluster.
func (p *Planner) NewCreateClusterTask(opts map[string]*CreateClusterNodeOpts) error {

	// Assign slots and replicate nodes
	roleAssignments := []*Command{}
	for nodeID, config := range opts {
		co := *NewCommandOpts()
		var command *Command
		var err error

		if config.ReplicationTarget != "" {
			co.AddKS("target", config.ReplicationTarget)
			command, err = NewCommand(nodeID, CommandSetReplicate, co, nil)
		} else {
			co.AddKIL("slots", config.Slots)
			command, err = NewCommand(nodeID, CommandAddSlots, co, nil)
		}

		if err != nil {
			return err
		}
		roleAssignments = append(roleAssignments, command)
	}

	// Set epochs
	epochAssignements := []*Command{}
	epoch := 1
	for nodeID := range opts {
		co := *NewCommandOpts()
		co.AddKS("epoch", string(epoch))
		epoch++
		command, err := NewCommand(nodeID, CommandSetEpoch, co, roleAssignments)
		if err != nil {
			return err
		}
		epochAssignements = append(epochAssignements, command)
	}

	// Join cluster command
	clusterJoin := []*Command{}
	lastNode := ""
	for nodeID, config := range opts {
		if lastNode == "" {
			lastNode = nodeID
			continue
		}

		co := *NewCommandOpts()
		co.AddKS("node", config.Addr)
		command, err := NewCommand(nodeID, CommandJoinCluster, co, epochAssignements)
		if err != nil {
			return err
		}
		clusterJoin = append(clusterJoin, command)

		lastNode = nodeID
	}

	commands := append(roleAssignments, epochAssignements...)
	commands = append(commands, clusterJoin...)

	createClusterTask, err := NewTask(TaskCreateCluster, commands)
	if err != nil {
		return err
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, createClusterTask)
	p.lock.Unlock()

	// Issue a check cluster task
	err = p.NewCheckClusterTask()
	if err != nil {
		log.Warnf("Could not create cluster check task: %v", err)
	}

	return nil
}

// NewCheckClusterTask creates a task for cluster checks
func (p *Planner) NewCheckClusterTask() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	return nil
}
