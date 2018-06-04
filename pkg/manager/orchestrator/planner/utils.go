package planner

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func shouldExecute(command *Command) bool {
	if command.ShouldExecute != nil {
		shouldExecute, err := (*command.ShouldExecute)(command)
		if err != nil {
			log.Warnf("Could not execute the shouldExecute func: %v", err)
			return false
		}
		if !shouldExecute {
			command.UpdateStatus(CommandFinished)
		}
		return shouldExecute
	}
	return true
}

func migrateSlot(source string, target string, targetAddr string, masters []string, slot int, fix bool, cold bool, dependencies []*Command) ([]*Command, error) {
	var commands []*Command

	if !cold {
		// Target importing
		nodeImportingOpts := *NewCommandOpts()
		nodeImportingOpts.AddKIL("slots", []int{slot})
		nodeImportingOpts.AddKS("state", "importing")
		nodeImportingOpts.AddKS("nodeID", source)
		nodeImporting, err := NewCommand(target, CommandSetSlotState, nodeImportingOpts, dependencies)
		if err != nil {
			return nil, err
		}
		commands = append(commands, nodeImporting)

		// Target importing
		nodeMigratingOpts := *NewCommandOpts()
		nodeMigratingOpts.AddKIL("slots", []int{slot})
		nodeMigratingOpts.AddKS("state", "migrating")
		nodeMigratingOpts.AddKS("nodeID", target)
		nodeMigrating, err := NewCommand(source, CommandSetSlotState, nodeMigratingOpts, dependencies)
		if err != nil {
			return nil, err
		}
		commands = append(commands, nodeMigrating)
	}

	var fixMessage string
	if fix {
		fixMessage = "true"
	} else {
		fixMessage = "false"
	}
	migrateSlotOpts := *NewCommandOpts()
	migrateSlotOpts.AddKIL("slots", []int{slot})
	migrateSlotOpts.AddKS("addr", targetAddr)
	migrateSlotOpts.AddKS("fix", fixMessage)
	migrateSlotCommand, err := NewCommand(source, CommandMigrateSlots, migrateSlotOpts, append(commands, dependencies...))
	if err != nil {
		return nil, err
	}
	commands = append(commands, migrateSlotCommand)

	if !cold {
		var slotUpdates []*Command
		for _, node := range masters {
			updateOpts := *NewCommandOpts()
			updateOpts.AddKIL("slots", []int{slot})
			updateOpts.AddKS("state", "node")
			updateOpts.AddKS("nodeID", target)
			updateCommand, err := NewCommand(node, CommandSetSlotState, updateOpts, commands)
			if err != nil {
				return nil, err
			}
			slotUpdates = append(slotUpdates, updateCommand)
		}
		commands = append(commands, slotUpdates...)
	}

	return commands, nil
}

func nodeWithMostKeysInSlot(counts map[string]SlotKeyCounts, slot int) (string, error) {
	var owner string
	var mostKeys int

	for node, c := range counts {
		count, ok := c.Counts[slot]
		if !ok {
			continue
		}
		if count > mostKeys {
			mostKeys = count
			owner = node
		}
	}

	if owner != "" {
		return owner, nil
	}
	return "", fmt.Errorf("could not find a node with keys form slot: %d", slot)
}

func slotCoverageNoMastersCreator(owner string, slot int, dependencies []*Command) ([]*Command, error) {
	var commands []*Command

	co := *NewCommandOpts()
	co.AddKIL("slots", []int{slot})
	co.AddKS("state", "stable")
	command, err := NewCommand(owner, CommandSetSlotState, co, dependencies)
	if err != nil {
		return nil, err
	}
	commands = append(commands, command)

	co = *NewCommandOpts()
	co.AddKIL("slots", []int{slot})
	command, err = NewCommand(owner, CommandAddSlots, co, commands)
	if err != nil {
		return nil, err
	}
	commands = append(commands, command)

	co = *NewCommandOpts()
	command, err = NewCommand(owner, CommandBumpEpoch, co, commands)
	if err != nil {
		return nil, err
	}
	commands = append(commands, command)

	return commands, nil
}

func slotCoverageMultipleMasters(owner string, nodesWithKeys []string, slot int, dependencies []*Command) ([]*Command, error) {
	var commands []*Command

	for _, node := range nodesWithKeys {
		co1 := *NewCommandOpts()
		co1.AddKIL("slots", []int{slot})
		command1, err := NewCommand(node, CommandDelSlots, co1, dependencies)
		if err != nil {
			return nil, err
		}

		co2 := *NewCommandOpts()
		co2.AddKIL("slots", []int{slot})
		co2.AddKS("state", "importing")
		co2.AddKS("nodeID", owner)
		command2, err := NewCommand(node, CommandSetSlotState, co2, append(dependencies, command1))
		if err != nil {
			return nil, err
		}

		commands = append(commands, command1)
		commands = append(commands, command2)
	}

	co := *NewCommandOpts()
	command, err := NewCommand(owner, CommandBumpEpoch, co, commands)
	if err != nil {
		return nil, err
	}
	commands = append(commands, command)

	return commands, nil
}
