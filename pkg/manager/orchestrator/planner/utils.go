package planner

import (
	log "github.com/sirupsen/logrus"
)

// HumanizeTaskType returns a human readable string based on a taskType
func HumanizeTaskType(t int64) string {
	switch taskType(t) {
	case TaskCreateCluster:
		return "CreateCluster"
	case TaskMemberFixup:
		return "MemberFixup"
	case TaskFixOpenSlots:
		return "FixOpenSlots"
	case TaskFixSlotAllocation:
		return "FixSlotAllocation"
	case TaskAddNodeCluster:
		return "AddNode"
	case TaskRemoveNodeCluster:
		return "RemoveMode"
	default:
		return "TaskType_NOT_SET"
	}
}

// HumanizeTaskStatus a human readable string based on a taskStatus
func HumanizeTaskStatus(status int64) string {
	switch taskStatus(status) {
	case StatusWaiting:
		return "Waiting"
	case StatusExecuting:
		return "Executing"
	case StatusExecuted:
		return "Executed"
	default:
		return "TaskStatus_NOT_SET"
	}
}

// HumanizeCommandType returns a human readable string based on a commandType
func HumanizeCommandType(t int64) string {
	switch commandType(t) {
	case CommandAddSlots:
		return "AddSlots"
	case CommandSetReplicate:
		return "SetReplicate"
	case CommandSetEpoch:
		return "SetEpoch"
	case CommandJoinCluster:
		return "JoinCluster"
	case CommandCountKeysInSlots:
		return "CountKeysInSlots"
	default:
		return "CommandType_NOT_SET"
	}
}

// HumanizeCommandStatus returns a human readable string based on a commandStatus
func HumanizeCommandStatus(status int64) string {
	switch commandStatus(status) {
	case CommandWaiting:
		return "Waiting"
	case CommandRunning:
		return "Running"
	case CommandFinished:
		return "Finished"
	default:
		return "CommandStatus_NOT_SET"
	}
}

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
