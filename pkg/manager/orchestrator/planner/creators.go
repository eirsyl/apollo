package planner

import (
	"strconv"

	"errors"

	"github.com/eirsyl/apollo/pkg/utils"
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
	// TODO: Cluster does not agree on member configuration. This has to be fixed manually by a system administrator.
	log.Warn("Cluster configuration is inconsistent, the manager cannot fix this. Admin interference is required.")
	return nil
}

// NewOpenSlotsFixupTask tries to close open slots
func (p *Planner) NewOpenSlotsFixupTask(clusterNodes []string, openSlots []int, planner SlotCloserPlanner) error {
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
	var masters []string

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
		masters = append(masters, node)
	}

	commands = append(commands, keysCount...)
	log.Infof("Retrieved masters: %v", masters)

	task, err := NewTask(TaskFixOpenSlots, commands)
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
			log.Infof("Calculating slot closings")

			var ownerCommands []*Command
			var newCommands []*Command

			log.Info("Open slots: %v", openSlots)

			for _, slot := range openSlots {
				// Slot owners
				owners, err := planner.SlotOwners(slot)
				if err != nil {
					return err
				}

				// Importing nodes
				importing, err := planner.ImportingNodes(slot)
				if err != nil {
					return err
				}

				// Migrating nodes
				migrating, err := planner.MigratingNodes(slot)
				if err != nil {
					return err
				}

				log.Infof("Owners: %v", owners)
				log.Infof("Importing nodes: %v", importing)
				log.Infof("Migrating nodes: %v", migrating)

				var owner string
				switch len(owners) {
				case 0:
					// Assign owner to node with most keys in slot
					// Cannot fix if no node has keys in slot
					owner, err = nodeWithMostKeysInSlot(keys, slot)
					if err != nil {
						return err
					}
					log.Infof("Using %s as owner, the node has most related keys", owner)
					ownerCommands, err = slotCoverageNoMastersCreator(owner, slot, keysCount)
					if err != nil {
						return err
					}
					importing = utils.RemoveFromStringList(importing, owner)
					migrating = utils.RemoveFromStringList(migrating, owner)
				case 1:
					// We have the owner, continue without actions
					owner = owners[0]
					log.Infof("Using %s as owner, this is the only node that reports slot ownership", owner)
				default:
					// Multiple owners, define a owner and continue
					ownerKeys := map[string]SlotKeyCounts{}
					for _, owner := range owners {
						ownerKeys[owner] = keys[owner]
					}
					owner, err = nodeWithMostKeysInSlot(ownerKeys, slot)
					if err != nil {
						return err
					}
					var nodesWithKeys []string
					for node := range ownerKeys {
						if node == owner {
							continue
						}
						nodesWithKeys = append(nodesWithKeys, node)
						importing = utils.RemoveFromStringList(importing, node)
						importing = append(importing, node)
					}
					ownerCommands, err = slotCoverageMultipleMasters(owner, nodesWithKeys, slot, keysCount)
					if err != nil {
						return err
					}
					log.Info("Using %s as owner, other nodes that reported slot ownership: %v", owner, nodesWithKeys)
				}

				if len(importing) == 1 && len(migrating) == 1 {
					// Case 1: one importing and one migrating node
					migratingNode := migrating[0]
					importingNode := importing[0]
					addr, err := planner.GetAddr(importingNode)
					if err != nil {
						return err
					}
					commands, err := migrateSlot(migratingNode, importingNode, addr, masters, slot, true, false, ownerCommands)
					if err != nil {
						return err
					}
					newCommands = append(newCommands, commands...)
				} else if len(migrating) == 0 && len(importing) > 0 {
					// Case 2: Multiple nodes with the slot in importing state
					for _, source := range importing {
						addr, err := planner.GetAddr(owner)
						if err != nil {
							return err
						}
						commands, err := migrateSlot(source, owner, addr, masters, slot, true, true, ownerCommands)
						if err != nil {
							return err
						}
						newCommands = append(newCommands, commands...)

						// Set slot to stable
						co := *NewCommandOpts()
						co.AddKIL("slots", []int{slot})
						co.AddKS("state", "stable")
						command, err := NewCommand(source, CommandSetSlotState, co, commands)
						if err != nil {
							return err
						}
						newCommands = append(newCommands, command)
					}
				} else if len(importing) == 0 && len(migrating) == 1 {
					// Case 3: One empty node with the slot in migrating state
					migratingNode := migrating[0]
					counts, ok := keys[migratingNode]
					if !ok {
						continue
					}
					keyCount, ok := counts.Counts[slot]
					if !ok {
						continue
					}
					if keyCount != 0 {
						return errors.New("cannot handle case 3 when the migrating node has keys in the slot")
					}
					// Set slot to stable
					co := *NewCommandOpts()
					co.AddKIL("slots", []int{slot})
					co.AddKS("state", "stable")
					command, err := NewCommand(migratingNode, CommandSetSlotState, co, ownerCommands)
					if err != nil {
						return err
					}
					newCommands = append(newCommands, command)
				} else {
					// This case cannot be handled by the manager
					// TODO: Address more cases and try to fix them.
					log.Warnf("Cannot address open slot case, importing: %d, migrating: %d", len(importing), len(migrating))
					return errors.New("cannot handle open slot case")
				}
			}

			if len(newCommands) == 0 {
				return errors.New("could not calculate slot closing commands")
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
			log.Info("Calculating slot allocation strategy")

			var newCommands []*Command

			// Case 1: No nodes have keys belonging to the slot
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

			// Case 2: One node have keys belonging to the slot
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

			// Case 3: Multiple nodes have keys belonging to the slot
			advancedSlotAssignments, e := planner.AllocateSlotsWithMultipleNodes(keys)
			if e != nil {
				return e
			}
			for slot, allocation := range advancedSlotAssignments {
				target := allocation.Master

				addSlotOpts := *NewCommandOpts()
				addSlotOpts.AddKIL("slots", []int{slot})
				addSlotCmd, e := NewCommand(target, CommandAddSlots, addSlotOpts, keysCount)
				if e != nil {
					return e
				}
				newCommands = append(newCommands, addSlotCmd)

				setSlotOpts := *NewCommandOpts()
				setSlotOpts.AddKIL("slots", []int{slot})
				setSlotOpts.AddKS("state", "stable")
				setSlotCmd, e := NewCommand(target, CommandSetSlotState, setSlotOpts, []*Command{addSlotCmd})
				if e != nil {
					return e
				}
				newCommands = append(newCommands, setSlotCmd)

				for _, node := range allocation.Nodes {
					if node == allocation.Master {
						continue
					}

					nodeImportingOpts := *NewCommandOpts()
					nodeImportingOpts.AddKIL("slots", []int{slot})
					nodeImportingOpts.AddKS("state", "importing")
					nodeImportingOpts.AddKS("nodeID", allocation.Master)
					nodeImportingCmd, err := NewCommand(node, CommandSetSlotState, nodeImportingOpts, []*Command{setSlotCmd})
					if err != nil {
						return err
					}

					masterAddr, err := planner.GetAddr(allocation.Master)
					if err != nil {
						return nil
					}

					migrateCommands, err := migrateSlot(node, allocation.Master, masterAddr, nil, slot, true, true, []*Command{nodeImportingCmd})
					if err != nil {
						return err
					}

					nodeStableOpts := *NewCommandOpts()
					nodeStableOpts.AddKIL("slots", []int{slot})
					nodeStableOpts.AddKS("state", "stable")
					nodeStableCmd, err := NewCommand(node, CommandSetSlotState, nodeStableOpts, migrateCommands)
					if err != nil {
						return err
					}

					newCommands = append(newCommands, nodeImportingCmd)
					newCommands = append(newCommands, migrateCommands...)
					newCommands = append(newCommands, nodeStableCmd)
				}
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
func (p *Planner) NewAddNodeTask(nodes []string, planner AddNodePlanner, replication int) error {
	/**
	 * Assign as master or replicate existing master?
	 * Pick master with the least replicas
	 * Send CLUSTER MEET to the node
	 * Tell node to replicate master
	 *
	 * A rebalancing task generated in the next iteration add slots to newly assigned masters
	 */
	var tasks []*Task

	plans, err := planner.GetNodePlans(nodes, replication)
	if err != nil {
		return err
	}

	nodeToJoin, err := planner.GetNodeToJoin()
	if err != nil {
		return nil
	}

	// Create node task per plan
	for _, plan := range plans {
		if plan.IsMaster {
			var commands []*Command

			co := *NewCommandOpts()
			co.AddKS("node", nodeToJoin)
			command, err := NewCommand(plan.NodeID, CommandJoinCluster, co, nil)
			if err != nil {
				return err
			}
			commands = append(commands, command)

			task, err := NewTask(TaskAddNodeCluster, commands)
			if err != nil {
				return err
			}
			tasks = append(tasks, task)
		} else if !plan.IsMaster && plan.ReplicationTarget != "" {
			// Replicate
			var commands []*Command

			co := *NewCommandOpts()
			co.AddKS("node", nodeToJoin)
			command, err := NewCommand(plan.NodeID, CommandJoinCluster, co, nil)
			if err != nil {
				return err
			}
			commands = append(commands, command)

			co = *NewCommandOpts()
			co.AddKS("target", plan.ReplicationTarget)
			command, err = NewCommand(plan.NodeID, CommandSetReplicate, co, commands)
			if err != nil {
				return err
			}
			commands = append(commands, command)

			task, err := NewTask(TaskAddNodeCluster, commands)
			if err != nil {
				return err
			}
			tasks = append(tasks, task)
		} else {
			log.Warnf("Planner generated invalid node plan: wrong parameters")
			return nil
		}
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, tasks...)
	p.lock.Unlock()

	return nil
}

// NewRemoveNodeTask add additional nodes to the cluster
func (p *Planner) NewRemoveNodeTask(nodes []string) error {
	/**
	 * TODO: Node removal isn't a high priority task, implementing this is considered as further work.
	 * Reshard slots away
	 * Loop cluster members
	 * If node replicates this node, replicate another
	 * Send CLUSTER FORGET
	 * Shutdown node
	 */
	var commands []*Command

	log.Warn("The manager should have removed nodes from cluster but functionality is missing: %v", nodes)

	task, err := NewTask(TaskRemoveNodeCluster, commands)
	if err != nil {
		return err
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, task)
	p.lock.Unlock()

	return nil
}

// NewMigrateSlotTask moves a slot from one node to another
func (p *Planner) NewMigrateSlotTask(source string, destination string, destinationAddr string, masters []string, slots []int) error {
	var commands []*Command

	for _, slot := range slots {
		migrateCommands, err := migrateSlot(source, destination, destinationAddr, masters, slot, true, false, nil)
		if err != nil {
			return err
		}
		commands = append(commands, migrateCommands...)
	}

	task, err := NewTask(TaskReshardCluster, commands)
	if err != nil {
		return err
	}

	p.lock.Lock()
	p.tasks = append(p.tasks, task)
	p.lock.Unlock()

	return nil
}
