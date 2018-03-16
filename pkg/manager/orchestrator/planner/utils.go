package planner

// HumanizeTaskType returns a human readable string based on a taskType
func HumanizeTaskType(t int64) string {
	switch taskType(t) {
	case TaskCreateCluster:
		return "CreateCluster"
	case TaskCheckCluster:
		return "CheckCluster"
	case TaskFixSlotAllocation:
		return "FixSlotAllocation"
	case TaskRebalanceCluster:
		return "Rebalance"
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
