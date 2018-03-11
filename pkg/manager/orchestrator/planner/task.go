package planner

type taskType int
type taskStatus int

var (
	TaskCreateCluster     taskType = 1
	TaskCheckCluster      taskType = 2
	TaskFixSlotAllocation taskType = 3
	TaskRebalanceCluster  taskType = 4
	TaskAddNodeCluster    taskType = 5
	TaskRemoveNodeCluster taskType = 6

	StatusWaiting   taskStatus = 0
	StatusExecuting taskStatus = 1
	StatusExecuted  taskStatus = 2
)

// Task represents a task that the cluster manager should execute
type Task struct {
	Type     taskType
	Status   taskStatus
	Commands []*Command
}

// NewTask creates a new task
func NewTask(t taskType, commands []*Command) (*Task, error) {
	task := &Task{
		Type:     t,
		Status:   StatusWaiting,
		Commands: commands,
	}
	return task, nil
}

func (t *Task) NextCommand(nodeId string) (*Command, error) {
	return nil, nil
}
