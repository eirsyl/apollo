package planner

import (
	"strconv"

	log "github.com/sirupsen/logrus"
)

// CreateClusterNodeOpts is used to provide extra data to the NewCreateClusterTask func.
type CreateClusterNodeOpts struct {
	Slots             []int
	ReplicationTarget string
	Addr              string
}

// OpenSlotsFixupOpts is used to provide extra data to the NewSlotCoverageFixupTask func.
type OpenSlotsFixupOpts struct {
	Slots []int
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

	task, err := NewTask(TaskCreateCluster, commands)
	if err != nil {
		return err
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, task)
	p.lock.Unlock()

	return nil
}

// NewClusterMemberFixupTask tries to notify agents about the cluster topology in order to fix
// cluster partitions
func (p *Planner) NewClusterMemberFixupTask() error {
	// TODO: Cluster does not agree on slot configuration. This has to be fixed manually by a system administrator.
	log.Warn("Cluster configuration is inconsistent, the manager cannot fix this. Admin interference is required.")
	return nil
}

// NewOpenSlotsFixupTask tries to close open slots
func (p *Planner) NewOpenSlotsFixupTask() error {
	/**
	 * Find the slot owner and the states
	 * If no owner: set owner to the node with most keys
	 * If multiple owners: Set owner to the node with most keys
	 * Process different cases:
	 * 1. The slot is in migrating state in one slot, and in importing state in 1 slot
	 * 2. Multiple nodes with the slot in importing state
	 * 3. One empty node with the slot in migrating state
	 * Other cases mut be reported to system administrators
	 */
	var commands []*Command

	task, err := NewTask(TaskFixOpenSlots, commands)
	if err != nil {
		return err
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, task)
	p.lock.Unlock()

	return nil
}

// NewSlotCoverageFixupTask tries to fix slot coverage
func (p *Planner) NewSlotCoverageFixupTask(clusterNodes []string, openSlots []int, planner SlotCoveragePlanner) error {
	/**
	 * Take action for every open slot depending on the cluster condition:
	 * 1. No nodes has keys belonging to the slot
	 * 2. One node have keys for the open slot
	 * 3. Multiple nodes have keys belonging to the slot
	 */

	var commands []*Command

	keysCount := []*Command{}
	for _, node := range clusterNodes {
		isMaster, err := planner.IsMasterNode(node)
		if err != nil {
			return err
		}
		if !isMaster {
			continue
		}

		co := *NewCommandOpts()
		co.AddKIL("slots", openSlots)
		command, err := NewCommand(node, CommandCountKeysInSlots, co, nil)
		if err != nil {
			return err
		}
		keysCount = append(keysCount, command)
	}

	commands = append(commands, keysCount...)

	task, err := NewTask(TaskFixSlotAllocation, commands)
	if err != nil {
		return err
	}

	// This task depends on information retrieved by previous commands
	task.ProcessResults = func(task *Task) error {
		// The processor have already added tasks
		if len(task.Commands) != len(keysCount) {
			return nil
		}

		keys := map[string]SlotKeyCounts{}
		ready := true
		for _, c := range keysCount {
			if c.Status != CommandFinished {
				ready = false
				continue
			}
			if c.Type == CommandCountKeysInSlots && len(c.Results) > 0 {
				kc := NewSlotKeyCounts(c.Results[len(c.Results)-1])
				keys[c.NodeID] = kc
			}
		}

		if ready {
			log.Info("Calculation slot allocation strategy")

			var newCommands []*Command

			slotAssignments, e := planner.AllocateSlotsWithoutKeys(keys)
			if e != nil {
				return e
			}
			for node, allocation := range slotAssignments {
				co := *NewCommandOpts()
				co.AddKIL("slots", allocation.Slots)
				command, e := NewCommand(node, CommandAddSlots, co, keysCount)
				if e != nil {
					return e
				}
				newCommands = append(newCommands, command)
			}

			slotAssignments, e = planner.AllocateSlotsWithOneNode(keys)
			if e != nil {
				return e
			}
			for node, allocation := range slotAssignments {
				co := *NewCommandOpts()
				co.AddKIL("slots", allocation.Slots)
				command, e := NewCommand(node, CommandAddSlots, co, keysCount)
				if e != nil {
					return e
				}
				newCommands = append(newCommands, command)
			}

			advancedSlotAssignments, e := planner.AllocateSlotsWithMultipleNodes(keys)
			if e != nil {
				return e
			}
			for node, allocation := range advancedSlotAssignments {
				// TODO: Implement multiple nodes with keys
				log.Infof("[WIP] Multiple nodes with keys: %s %v", node, allocation)
			}

			task.Commands = append(task.Commands, newCommands...)
			log.Infof("New Tasks: %v", newCommands)
		}

		return nil
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, task)
	p.lock.Unlock()

	return nil
}

// NewAddNodeTask add additional nodes to the cluster
func (p *Planner) NewAddNodeTask() error {
	/**
	 * Assign as master or replicate existing master?
	 * Pick master with the least replicas
	 * Send CLUSTER MEET to the node
	 * Tell node to replicate master
	 */
	var commands []*Command

	task, err := NewTask(TaskAddNodeCluster, commands)
	if err != nil {
		return err
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, task)
	p.lock.Unlock()

	return nil
}

// NewRemoveNodeTask add additional nodes to the cluster
func (p *Planner) NewRemoveNodeTask() error {
	/**
	 * Reshard slots away
	 * Loop cluster members
	 * If node replicates this node, replicate another
	 * Send CLUSTER FORGET
	 * Shutdown node
	 */
	var commands []*Command

	task, err := NewTask(TaskRemoveNodeCluster, commands)
	if err != nil {
		return err
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, task)
	p.lock.Unlock()

	return nil
}
