package planner

import (
	"strconv"
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

	// Assign slots
	slotAssignments := []*Command{}
	replicationAssignments := []*Command{}

	for nodeID, config := range opts {
		co := *NewCommandOpts()
		var command *Command
		var err error

		if config.ReplicationTarget != "" {
			co.AddKS("target", config.ReplicationTarget)
			command, err = NewCommand(nodeID, CommandSetReplicate, co, nil)
			if err != nil {
				return err
			}
			replicationAssignments = append(replicationAssignments, command)
		} else {
			co.AddKIL("slots", config.Slots)
			command, err = NewCommand(nodeID, CommandAddSlots, co, nil)
			if err != nil {
				return err
			}
			slotAssignments = append(slotAssignments, command)
		}
	}

	// Set epochs
	epochAssignements := []*Command{}
	epoch := 1
	for nodeID := range opts {
		co := *NewCommandOpts()
		co.AddKS("epoch", strconv.Itoa(epoch))
		epoch++
		command, err := NewCommand(nodeID, CommandSetEpoch, co, slotAssignments)
		if err != nil {
			return err
		}
		epochAssignements = append(epochAssignements, command)
	}

	// Join cluster command
	clusterJoin := []*Command{}
	lastAddr := ""
	for nodeID, config := range opts {
		if lastAddr == "" {
			lastAddr = config.Addr
			continue
		}

		co := *NewCommandOpts()
		co.AddKS("node", lastAddr)
		command, err := NewCommand(nodeID, CommandJoinCluster, co, epochAssignements)
		if err != nil {
			return err
		}
		clusterJoin = append(clusterJoin, command)

		lastAddr = config.Addr
	}

	for _, command := range replicationAssignments {
		command.Dependencies = clusterJoin
	}

	commands := append(slotAssignments, epochAssignements...)
	commands = append(commands, clusterJoin...)
	commands = append(commands, replicationAssignments...)

	createClusterTask, err := NewTask(TaskCreateCluster, commands)
	if err != nil {
		return err
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, createClusterTask)
	p.lock.Unlock()

	return nil
}

// NewClusterMemberFixupTask tries to notify agents about the cluster topology in order to fix
// cluster partitions
func (p *Planner) NewClusterMemberFixupTask() error {
	// TODO: create task
	return nil
}

// NewOpenSlotsFixupTask tries to close open slots
func (p *Planner) NewOpenSlotsFixupTask() error {
	// TODO: create task
	return nil
}

// NewSlotCoverageFixupTask tries to fix slot coverage
func (p *Planner) NewSlotCoverageFixupTask() error {
	// TODO: create task
	return nil
}

// NewAddNodeTask add additional nodes to the cluster
func (p *Planner) NewAddNodeTask() error {
	return nil
}

// NewRemoveNodeTask add additional nodes to the cluster
func (p *Planner) NewRemoveNodeTask() error {
	return nil
}
