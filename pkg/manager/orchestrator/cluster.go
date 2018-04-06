package orchestrator

import (
	"time"

	"strconv"

	"fmt"

	"errors"

	"github.com/coreos/bbolt"
	"github.com/eirsyl/apollo/pkg"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/eirsyl/apollo/pkg/manager/orchestrator/planner"
	log "github.com/sirupsen/logrus"
)

type clusterState int

// Int64 returns the value as a int64
func (cs *clusterState) Int64() int64 {
	return int64(*cs)
}

type clusterHealth int

// Int64 returns the value as a int64
func (cs *clusterHealth) Int64() int64 {
	return int64(*cs)
}

const (
	clusterUnknown      clusterState = 0
	clusterUnconfigured clusterState = 1
	clusterConfigured   clusterState = 2

	clusterOK    clusterHealth = 0
	clusterWarn  clusterHealth = 1
	clusterError clusterHealth = 2
)

// Cluster stores the manager observations of the cluster
type Cluster struct {
	state              clusterState
	health             clusterHealth
	desiredReplication int
	minNodesCreate     int
	db                 *bolt.DB
	startTime          time.Time
	planner            *planner.Planner
	nodeManager        *nodeManager
}

// NewCluster creates a new cluster planner
func NewCluster(desiredReplication, minNodesCreate int, db *bolt.DB) (*Cluster, error) {
	// Create buckets
	err := db.Update(func(tx *bolt.Tx) error {
		// Node info bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("nodes")); err != nil {
			return err
		}
		// Cluster member bucket
		if _, err := tx.CreateBucketIfNotExists([]byte("cluster")); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	p, err := planner.NewPlanner()
	if err != nil {
		return nil, err
	}

	nm, err := newNodeManager(db)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		state:              clusterUnknown,
		health:             clusterError,
		desiredReplication: desiredReplication,
		minNodesCreate:     minNodesCreate,
		db:                 db,
		startTime:          time.Now().UTC(),
		planner:            p,
		nodeManager:        nm,
	}, nil
}

// Run starts the manager reconciliation loop
// The loop is responsible for calculating optimal cluster state
func (c *Cluster) Run() error {
	for {
		switch c.state {
		case clusterUnknown:
			c.findClusterConfiguration()
		case clusterUnconfigured:
			c.configureCluster()
		case clusterConfigured:
			c.iteration()
		default:
			log.WithField("state", c.state).Warnf("Unhandled cluster state")
		}
		time.Sleep(10 * time.Second)
	}
}

// findClusterConfiguration is responsible for trying to figure out if the cluster is configured or not
func (c *Cluster) findClusterConfiguration() {
	log.Info("Starting configuration detection")
	nodes, err := c.nodeManager.onlineNodes()
	if err != nil {
		log.Infof("Could not retrieve online nodes from nodeManager: %v", err)
		return
	}

	if len(nodes) < 3 {
		log.Infof("Skipping configuration detection, minimum 3 nodes has to be online and reporting state.")
		return
	}

	var onlineNodesEmpty = true
	for _, node := range nodes {
		if !node.IsEmpty {
			onlineNodesEmpty = false
			break
		}
	}

	if onlineNodesEmpty {
		log.Infof("Online nodes empty, setting cluster state to unconfigured and preparing for cluster initialization")
		c.state = clusterUnconfigured
	} else {
		log.Infof("Nodes not empty, setting cluster to configured")
		c.state = clusterConfigured
	}
}

// configureCluster configures the redis nodes as a cluster if all the requirements is meet
func (c *Cluster) configureCluster() {
	nodes, err := c.nodeManager.onlineNodes()
	if err != nil {
		log.Infof("Could not retrieve online nodes from node manager: %v", err)
		return
	}

	if len(nodes) < c.minNodesCreate {
		log.Infof("Skipping cluster creation %d nodes is required", c.minNodesCreate)
		return
	}

	log.Infof("Generating cluster creation plan")
	var nodeIds []string
	for _, node := range nodes {
		nodeIds = append(nodeIds, node.ID)
	}

	validConfiguration := validateClusterSize(len(nodeIds), c.desiredReplication)
	if !validConfiguration {
		log.Warnf("invalid cluster, Redis Cluster requires at least 3 master nodes. %d nodes is required with the replication %d", 3*(c.desiredReplication+1), c.desiredReplication)
		return
	}

	// Create the cluster creation task and update cluster state
	err = c.nodeManager.setClusterNodes(nodeIds)
	if err != nil {
		log.Warnf("Could not update cluster members: %v", err)
		return
	}

	slots, err := allocSlots(&nodes, c.desiredReplication)
	if err != nil {
		log.Warnf("Could not calculate slots: %v", err)
	}

	err = c.planner.NewCreateClusterTask(slots)
	if err != nil {
		log.Infof("Could not generate cluster creation plan: %v", err)
		return
	}

	log.Info("Updating cluster state")
	c.health = clusterWarn
	c.state = clusterConfigured
}

// iteration watches the cluster after everything is configured and up an running
func (c *Cluster) iteration() {
	l := log.WithFields(log.Fields{"clusterHealth": HumanizeClusterHealth(c.health.Int64()), "clusterState": HumanizeClusterState(c.state.Int64())})
	l.Info("Running iteration")

	task, err := c.planner.CurrentTask()
	if err != nil {
		l.Infof("Planner error: %v", err)
	}

	if task == nil {
		// Set cluster state to unknown if all nodes becomes empty while the manager is running
		if c.state == clusterConfigured {
			nodes, err := c.nodeManager.onlineNodes()
			if err == nil {
				var onlineNodesEmpty = true
				for _, node := range nodes {
					if !node.IsEmpty {
						onlineNodesEmpty = false
						break
					}
				}
				if onlineNodesEmpty {
					log.Info("All nodes appear to be empty, cluster state is unknown")
					c.state = clusterUnknown
					return
				}
			}
		}

		l.Info("No running task, initializing a cluster check")
		c.checkCluster()
	} else {
		l.Infof("Watching task execution: %s", planner.HumanizeTaskType(task.Type.Int64()))
		c.watchTask(task)
	}

}

// ReportState collects the state from the reporting node
func (c *Cluster) ReportState(node *pb.StateRequest) error {
	n, err := NewNodeFromPb(node)
	if err != nil {
		return err
	}

	return c.nodeManager.updateNode(n)
}

// NextExecution sends the next command to a node. The planner returns a command if
// a step is planned by the manager.
func (c *Cluster) NextExecution(req *pb.NextExecutionRequest) (*pb.NextExecutionResponse, error) {
	task, nextCommands, err := c.planner.NextCommands(req.NodeID)

	if err != nil {
		return nil, err
	}

	var commands []*pb.ExecutionCommand
	for _, command := range nextCommands {
		command.UpdateStatus(planner.CommandRunning)
		var arguments []string

		switch command.Type {
		case planner.CommandAddSlots:
			for _, slot := range command.Opts.GetKIL("slots") {
				arguments = append(arguments, strconv.Itoa(slot))
			}
		case planner.CommandSetReplicate:
			arguments = []string{command.Opts.GetKS("target")}
		case planner.CommandSetEpoch:
			arguments = []string{command.Opts.GetKS("epoch")}
		case planner.CommandJoinCluster:
			arguments = []string{command.Opts.GetKS("node")}
		default:
			log.Warnf("Received command without opt parser: %v", command.Type)
		}

		log.Infof("Sending command to agent: %v %v %v", req.NodeID, command.ID.String(), command.Type)
		commands = append(commands, &pb.ExecutionCommand{
			Id:        command.ID.String(),
			Command:   command.Type.Int64(),
			Arguments: arguments,
		})
	}

	if task != nil {
		(*task).UpdateStatus()
	}

	return &pb.NextExecutionResponse{
		Commands: commands,
	}, nil
}

// ReportExecution reports the status of a command executed on a node
func (c *Cluster) ReportExecution(req *pb.ReportExecutionRequest) error {
	var results []*planner.CommandResult

	for _, result := range req.CommandResults {
		log.Infof("Storing execution result: %v %v", req.NodeID, result.Id)
		cr, err := planner.NewCommandResult(result.Id, result.Result, result.Success)
		if err != nil {
			return err
		}
		results = append(results, cr)
	}

	return c.planner.ReportResult(req.NodeID, results)
}

/**
 * Functions bellow this line is used to generate and manage cluster state
 */

// checkCluster validates the cluster and executes new tasks if required
func (c *Cluster) checkCluster() {
	onlineNodes, clusterNodes, err := c.findClusterAndOnlineNodes()
	if err != nil {
		return
	}

	// Update manager state if the clusterNodes is unset for some reason
	// One requirement: The cluster has to be clean and clusterMembers = onlineNodes
	c.maybeSetClusterNodes(onlineNodes, clusterNodes)

	// Check if cluster is consistent (All nodes have the same observation of members)
	clusterMembers, clusterMemberValid, offlineAgents, err := c.validateClusterMembers(onlineNodes, clusterNodes)
	if err != nil {
		log.Warn("Could not validate cluster members, skipping iteration")
		return
	}
	if len(offlineAgents) > 0 {
		log.Warnf("Some nodes is configured as cluster members, but not agent is reporting state: %v", offlineAgents)
		log.Warn("Skipping iteration, please make sure each redis node has an active apollo agent.")
		return
	}
	if !clusterMemberValid {
		err = c.planner.NewClusterMemberFixupTask()
		if err != nil {
			log.Warnf("Could not create planner task: %v", err)
		}
		c.health = clusterWarn
		return
	}

	nodesWithOpenSlots, openSlotsValid, err := c.validateOpenSlots(clusterMembers)
	if err != nil {
		log.Warn("Could not validate open slots, skipping iteration")
		return
	}
	if !openSlotsValid {
		log.Warnf("Nodes with open slots: %d", len(*nodesWithOpenSlots))

		err = c.planner.NewOpenSlotsFixupTask()
		if err != nil {
			log.Warnf("Could not create planner task: %v", err)
		}
		c.health = clusterWarn
		return
	}

	slotCoverageValid, err := c.validateSlotCoverage(clusterMembers)
	if err != nil {
		log.Warn("Could not validate slot coverage, skipping iteration")
		return
	}
	if !slotCoverageValid {
		err = c.planner.NewSlotCoverageFixupTask()
		if err != nil {
			log.Warnf("Could not create planner task: %v", err)
		}
		c.health = clusterWarn
		return
	}

	{
		c.health = clusterOK
		log.Info("Cluster healthy")
		err = c.nodeManager.garbageCollectNodes()
		if err != nil {
			log.Warnf("Could not remove unused node data: %v", err)
		}
	}

	// Cluster healthy start the process of adding or removing a node from the cluster
	err = c.createNodeManagementTasks()
	if err != nil {
		log.Warnf("Could not initialize node addition/removal tasks: %v", err)
	}
}

// watchTask watches task execution and tries to fix execution errors
func (c *Cluster) watchTask(task *planner.Task) {
	log.Infof("Watches task: %s (NOT IMPLEMENTED)", planner.HumanizeTaskType(task.Type.Int64()))
}

// findClusterAndOnlineNodes looks up online nodes and cluster members from db
func (c *Cluster) findClusterAndOnlineNodes() (*[]Node, *[]string, error) {
	clusterNodes, err := c.nodeManager.getClusterNodes()
	if err != nil {
		log.Warnf("Could get stored cluster nodes, finding nodes from cluster. %v", err)
		clusterNodes = []string{}
	}

	onlineNodes, err := c.nodeManager.onlineNodes()
	if err != nil {
		log.Warnf("Could not retrieve online nodes, loop continues: %v", err)
		return nil, nil, err
	}

	return &onlineNodes, &clusterNodes, nil
}

// maybeClusterNodes sets the clusterNodes if it is unset and the cluster is clean
func (c *Cluster) maybeSetClusterNodes(onlineNodes *[]Node, clusterNodes *[]string) {
	discoveredNodeIds, clean, err := findClusterNodes(onlineNodes)
	if err != nil {
		log.Warnf("Could not find cluster nodes: %v", err)
		return
	}

	if len(*clusterNodes) != 0 {
		// cluster nodes set, just continue
		return
	}

	if !clean {
		log.Warnf("Cluster is not clean, nodes report different cluster members, skipping clusterNode config")
		return
	}

	if clean {
		log.Info("Cluster nodes is empty, forcing update")
		err := c.nodeManager.setClusterNodes(discoveredNodeIds)
		if err != nil {
			log.Warnf("Could not update cluster members: %v", err)
		} else {
			*clusterNodes = discoveredNodeIds
		}
	}
}

// validateClusterMembers validates the node cluster membership state
func (c *Cluster) validateClusterMembers(onlineNodes *[]Node, clusterNodes *[]string) (*[]Node, bool, []string, error) {
	// onlineNodes: Nodes with an agent that pushes metrics and fetches tasks
	// clusterNodes: Nodes that the manager consider as cluster members
	// Returns (clusterMembers, consistentCluster, offlineNodes, error)

	// Validate the nodes that actually is a member of the cluster
	// Detect nodes that want to become a member if the cluster
	// Splits or multiple clusters?
	// Is a node agent offline?

	log.Info("Validating cluster members")

	onlineNodesMap := map[string]Node{}
	for _, node := range *onlineNodes {
		onlineNodesMap[node.ID] = node
	}

	clusterNodesMap := map[string]bool{}
	for _, node := range *clusterNodes {
		clusterNodesMap[node] = true
	}

	clusterNodeIds, clean, err := findClusterNodes(onlineNodes)
	if err != nil {
		return nil, false, nil, err
	}

	var offlineAgents []string
	for _, clusterNode := range clusterNodeIds {
		_, ok := onlineNodesMap[clusterNode]
		if !ok {
			offlineAgents = append(offlineAgents, clusterNode)
		}
	}

	var nodes []Node
	addableNodes := map[string]bool{}
	if clean {
		// Update cluster members if the cluster is clean
		err = c.nodeManager.setClusterNodes(clusterNodeIds)
		if err != nil {
			return nil, false, nil, err
		}
		// Use all the nodes returned by findClusterNodes
		for _, nodeID := range clusterNodeIds {
			node, ok := onlineNodesMap[nodeID]
			if !ok {
				continue
			}
			nodes = append(nodes, node)
		}
	} else {
		/**
		* This is the hard case, nodes appear to be member of different clusters.
		* This could either be a clean node or a configured node that is member of
		* another cluster.
		* Cases:
		* - Functional cluster and empty nodes
		* - Divided cluster, not every node is aware of the whole cluster
		* - To separate clusters is reporting state to the manager (causes crash, apollo cannot handle this case)
		 */
		signatures, e := nodeSignatures(onlineNodes)
		if e != nil {
			return nil, false, nil, e
		}

		if len(signatures) < 2 {
			// No nodes or clean cluster, should not happen.
			log.Warnf("The node validator reported the cluster as unclean, but the signature count is lower than 2")
			return nil, false, nil, errors.New("invalid signature count")
		}

		type clusterCase int
		var emptyNodes clusterCase = 1
		var partition clusterCase = 2
		var unknownClusters clusterCase = 3
		discoveredCases := map[clusterCase]bool{}
		emptySignatures := map[string]bool{}

		for signature, nodes := range signatures {
			// New empty node
			if len(nodes) == 1 {
				node := nodes[0]
				if node.IsEmpty && len(node.Nodes) == 1 {
					discoveredCases[emptyNodes] = true
					emptySignatures[signature] = true
					continue
				}
			}

			knownMembers := 0
			fullMatch := true
			for _, node := range nodes {
				for _, member := range node.Nodes {
					if clusterNodesMap[member.ID] {
						knownMembers++
					} else {
						fullMatch = false
					}
				}
			}

			// Overlapping signature (cluster partition)
			if knownMembers > 0 && !fullMatch {
				discoveredCases[partition] = true
				continue
			}

			// Completely unknown members (two different clusters is reporting to apollo)
			if !fullMatch {
				discoveredCases[unknownClusters] = true
			}
		}

		if discoveredCases[unknownClusters] {
			return nil, false, nil, errors.New("signature contains unknown cluster, cannot continue")
		}

		// Don't start cluster repair if the unknown signature is an empty node
		if !discoveredCases[partition] && discoveredCases[emptyNodes] {
			clean = true
		}

		// Merge partitioned signatures and filter away empty nodes
		mergedNodes := map[string]Node{}
		for signature, clusterMembers := range signatures {
			if !emptySignatures[signature] {
				for _, node := range clusterMembers {
					mergedNodes[node.ID] = node
				}
			} else {
				for _, node := range clusterMembers {
					addableNodes[node.ID] = true
				}
			}
		}
		for _, node := range mergedNodes {
			nodes = append(nodes, node)
		}

	}

	// Update addable nodes for later usage by the addNode task
	err = c.nodeManager.setEmptyNodes(mapKeysToString(addableNodes))
	if err != nil {
		log.Warnf("Could not set addable nodes: %v", err)
	}

	return &nodes, clean, offlineAgents, nil
}

// validateOpenSlots checks if the cluster has open slots
func (c *Cluster) validateOpenSlots(clusterMembers *[]Node) (*[]Node, bool, error) {
	log.Infof("Checking for open slots")
	var n []Node
	for _, node := range *clusterMembers {
		openSlots, err := node.MySelf.openSlots()
		if err != nil {
			return nil, false, err
		}
		if len(*openSlots) != 0 {
			n = append(n, node)
		}
	}
	return &n, len(n) == 0, nil
}

// validateSlotCoverage validates the slot coverage (Each slot need a responsible node)
func (c *Cluster) validateSlotCoverage(clusterMembers *[]Node) (bool, error) {
	log.Info("Checking slot coverage")
	var slots []int
	for _, node := range *clusterMembers {
		nodeSlots, err := node.MySelf.allSlots()
		if err != nil {
			return false, fmt.Errorf("could not read node slots: %v", err)
		}
		slots = append(slots, nodeSlots...)
	}
	return len(slots) == pkg.ClusterHashSlots, nil
}

// createNodeManagementTasks is responsible for creating tasks for node addition or removal
func (c *Cluster) createNodeManagementTasks() error {
	/**
	 * Consider online nodes only, this works because the cluster validation, or caller
	 * of this function skips the execution if not every agent is online.
	 */
	nodes, err := c.nodeManager.onlineNodes()
	if err != nil {
		return err
	}

	// Empty nodes lookup
	em, err := c.nodeManager.getEmptyNodes()
	if err != nil {
		return err
	}
	emptyNodes := map[string]bool{}
	for _, n := range em {
		emptyNodes[n] = true
	}

	var addableNodes []Node
	var removableNodes []Node

	for _, n := range nodes {
		if c.nodeManager.isNodePending(n.ID) {
			log.Info("Node pending, searching for next possible target: %v", n.ID)
			continue
		}
		if emptyNodes[n.ID] {
			addableNodes = append(addableNodes, n)
		} else if n.MarkedForDeletion {
			removableNodes = append(removableNodes, n)
		}
	}

	if len(addableNodes) > 0 {
		err = c.planner.NewAddNodeTask()
		if err != nil {
			return err
		}
	} else if len(removableNodes) > 0 {
		err = c.planner.NewRemoveNodeTask()
		if err != nil {
			return err
		}
	}

	return nil
}
