package planner

type taskType int
type taskStatus int

var (
	// TaskCreateCluster task is responsible for creating a cluster
	TaskCreateCluster taskType = 1
	// TaskCheckCluster is responsible for checking the configuration of a cluster
	TaskCheckCluster taskType = 2
	// TaskFixSlotAllocation is responsible for fixing slot allocation issues
	TaskFixSlotAllocation taskType = 3
	// TaskRebalanceCluster is responsible for moving slot assignments in order to prevent cluster issues
	TaskRebalanceCluster taskType = 4
	// TaskAddNodeCluster is used to add new nodes to the cluster
	TaskAddNodeCluster taskType = 5
	// TaskRemoveNodeCluster is used to remove a node from the cluster
	TaskRemoveNodeCluster taskType = 6

	// StatusWaiting is used for a task that is in the execution queue
	StatusWaiting taskStatus
	// StatusExecuting is used by tasks that is currently being executed by the nodes.
	StatusExecuting taskStatus = 1
	// StatusExecuted is used on tasks that is done
	StatusExecuted taskStatus = 2
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

// NextCommand returns the next command to be executed by a node with the provided NodeID
func (t *Task) NextCommand(nodeID string) (*Command, error) {
	return nil, nil
}
