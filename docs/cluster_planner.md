# Cluster Planner

* Execute commands to multiple nodes based command dependencies
* Watch execution results

## Use cases

The planner manages a list of planned tasks:

Tasks [
    create_cluster
    add_node
    rebalance_cluster
]

### Cluster creation

    Parent task:    create_cluster
    Status:         running                 (pending, running, failed)
    Steps:

        Set of command that has to be executed in order to
        complete the task. Each task may have a dependency.
        The manager needs to watch the status of each step.
    