# Roadmap

# TODO:

- [x] Lookup node name from `cluster nodes`
- [x] Make usage of the collected information from the info commands
- [ ] Manager should store the cluster topology

* Check slot coverage
* Check cluster configuration consistency (slot owners)
* Use host annotations to optimize for anti-affinity

# System properties

* Node monitoring
* Cluster planner
* Rebalancing
* Node management
* Administrative interface

# System lifecycle

## Components
* Agent
  - Reconciliation loop running each 30 second
  - Fetches node state
  - Reports state to the manager
  - Fetches commands from the master
  - Return results to the manager

* Manager
  - Continuously collect and store node state together with collection timestamp
  - Worker calculating the next operations to execute
  - Send the next command to execute to agents and monitor the outcome

# Functionality steps

## Required Cluster Knowledge

* Connected Nodes
* Cluster configured or not
* If not configured -> hook from user or least number of nodes?
* Cluster lifecycle state

## Create Cluster

1. Nodes reporting status to manager
2. Cluster is not configured
3. Cluster creation is initialized by admin or amount of online nodes
4. Make sure each node is empty
5. Check cluster requirements: number of replicas and masters
6. Allocate slots
7. Flush node config
8. Assign different epoch to each node
9. Join cluster
10. Wait for nodes to join
11. Check cluster

## Check Cluster

Simple task for checking slot allocation and coverage. Running this as
a background task if cluster is configured.

1. Check for cluster configuration inconsistencies
2. Check for open slots (migrating)
3. Is every slot handled by a node?

## Fix Broken Cluster Configuration

This state handles slots configuration issues based on three different cases:
1. No node has keys for this slot
2. A single node has keys for this slot
3. Multiple nodes has keys from this slot

CASE 1:
    Assign slot to random node with available resources.

CASE 2:
    Assign slot to the node with keys belonging to the slot.

CASE 3:
    Assing slot to the node with most keys belonging to the slot

## Reshard Cluster

1. The cluster config has to be valid!
2. Manual operation?


## Rebalance Cluster


## Add Node


## Delete Node