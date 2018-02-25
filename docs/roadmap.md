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
