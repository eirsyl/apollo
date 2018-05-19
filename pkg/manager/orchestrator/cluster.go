package orchestrator

import (
	"time"

	"strconv"

	"fmt"

	"errors"

	"sort"

	"math"

	"github.com/coreos/bbolt"
	"github.com/eirsyl/apollo/pkg"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/eirsyl/apollo/pkg/manager/orchestrator/planner"
	"github.com/eirsyl/apollo/pkg/utils"
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
		time.Sleep(pkg.ReconciliationLoopInterval)
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
	log.WithField("duration", time.Since(c.startTime)).Info("Configuration detected")
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
	startTime := time.Now().UTC()
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

	err = c.planner.NewCreateClusterTask(slots, startTime)
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
		case planner.CommandCountKeysInSlots:
			for _, slot := range command.Opts.GetKIL("slots") {
				arguments = append(arguments, strconv.Itoa(slot))
			}
		case planner.CommandSetSlotState:
			arguments = []string{
				command.Opts.GetKS("state"),
				command.Opts.GetKS("nodeID"),
			}
			for _, slot := range command.Opts.GetKIL("slots") {
				arguments = append(arguments, strconv.Itoa(slot))
			}
		case planner.CommandBumpEpoch:
			arguments = []string{}
		case planner.CommandDelSlots:
			for _, slot := range command.Opts.GetKIL("slots") {
				arguments = append(arguments, strconv.Itoa(slot))
			}
		case planner.CommandMigrateSlots:
			arguments = []string{
				command.Opts.GetKS("addr"),
				command.Opts.GetKS("fix"),
			}
			for _, slot := range command.Opts.GetKIL("slots") {
				arguments = append(arguments, strconv.Itoa(slot))
			}
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
	// Update the clusterNodes variable
	clusterNodes = &[]string{}
	for _, node := range *clusterMembers {
		*clusterNodes = append(*clusterNodes, node.ID)
	}

	nodesWithOpenSlots, openSlots, openSlotsValid, err := c.validateOpenSlots(clusterMembers)
	if err != nil {
		log.Warn("Could not validate open slots, skipping iteration")
		return
	}
	if !openSlotsValid {
		log.Warnf("Nodes with open slots: %d", len(*nodesWithOpenSlots))
		closePlanner, e := newSlotClosePlanner(c.nodeManager)
		if e != nil {
			log.Warnf("Could not intialize slot close planner: %v", e)
			return
		}
		err = c.planner.NewOpenSlotsFixupTask(*clusterNodes, openSlots, closePlanner)
		if err != nil {
			log.Warnf("Could not create planner task: %v", err)
		}
		c.health = clusterWarn
		return
	}

	openSlots, slotCoverageValid, err := c.validateSlotCoverage(clusterMembers)
	if err != nil {
		log.Warn("Could not validate slot coverage, skipping iteration")
		return
	}
	if !slotCoverageValid {
		slotPlanner, e := newSlotCoveragePlanner(c.nodeManager)
		if e != nil {
			log.Warnf("Could not initialize slot planner: %v", err)
			return
		}
		e = c.planner.NewSlotCoverageFixupTask(*clusterNodes, openSlots, slotPlanner)
		if e != nil {
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
	processingNodes, err := c.createNodeManagementTasks()
	if err != nil {
		log.Warnf("Could not initialize node addition/removal tasks: %v", err)
	}

	if !processingNodes {
		// Check cluster balance if the cluster is clean and no nodes is pending addition or removal
		err = c.balanceCluster()
		if err != nil {
			log.Warnf("Could not initialize cluster balancing: %v", err)
		}
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
	err = c.nodeManager.setEmptyNodes(utils.MapKeysToString(addableNodes))
	if err != nil {
		log.Warnf("Could not set addable nodes: %v", err)
	}

	return &nodes, clean, offlineAgents, nil
}

// validateOpenSlots checks if the cluster has open slots
func (c *Cluster) validateOpenSlots(clusterMembers *[]Node) (*[]Node, []int, bool, error) {
	log.Infof("Checking for open slots")
	var n []Node
	var slots []int
	slotMap := map[int]bool{}

	for _, node := range *clusterMembers {
		openSlots, err := node.MySelf.openSlots()
		if err != nil {
			return nil, nil, false, err
		}
		if len(*openSlots) != 0 {
			n = append(n, node)
			for _, os := range *openSlots {
				slotMap[os.slot] = true
			}
		}
	}

	for slot := range slotMap {
		slots = append(slots, slot)
	}

	return &n, slots, len(n) == 0, nil
}

// validateSlotCoverage validates the slot coverage (Each slot need a responsible node)
func (c *Cluster) validateSlotCoverage(clusterMembers *[]Node) ([]int, bool, error) {
	log.Info("Checking slot coverage")
	var slots []int
	for _, node := range *clusterMembers {
		nodeSlots, err := node.MySelf.allSlots()
		if err != nil {
			return nil, false, fmt.Errorf("could not read node slots: %v", err)
		}
		slots = append(slots, nodeSlots...)
	}
	return findOpenSlots(slots), len(slots) == pkg.ClusterHashSlots, nil
}

// createNodeManagementTasks is responsible for creating tasks for node addition or removal
func (c *Cluster) createNodeManagementTasks() (bool, error) {
	/**
	 * Consider online nodes only, this works because the cluster validation, or caller
	 * of this function skips the execution if not every agent is online.
	 */
	nodes, err := c.nodeManager.onlineNodes()
	if err != nil {
		return false, err
	}

	// Empty nodes lookup
	em, err := c.nodeManager.getEmptyNodes()
	if err != nil {
		return false, err
	}
	emptyNodes := map[string]bool{}
	for _, n := range em {
		emptyNodes[n] = true
	}

	var addableNodes []string
	var removableNodes []string

	for _, n := range nodes {
		if c.nodeManager.isNodePending(n.ID) {
			log.Info("Node pending, searching for next possible target: %v", n.ID)
			continue
		}
		if emptyNodes[n.ID] {
			addableNodes = append(addableNodes, n.ID)
		} else if n.MarkedForDeletion {
			removableNodes = append(removableNodes, n.ID)
		}
	}

	// Remove nodes first in order to not assign nodes as replicas for nodes marked for removal
	if len(removableNodes) > 0 {
		err = c.planner.NewRemoveNodeTask(removableNodes)
		if err != nil {
			return false, err
		}
	} else if len(addableNodes) > 0 {
		addNodePlanner, err := newAddNodePlanner(c.nodeManager)
		if err != nil {
			return false, err
		}
		err = c.planner.NewAddNodeTask(addableNodes, addNodePlanner, c.desiredReplication)
		if err != nil {
			return false, err
		}
	}

	return len(addableNodes) > 0 || len(removableNodes) > 0, nil
}

func (c *Cluster) balanceCluster() error {
	// Balance cluster checks the distribution of slots between the master modes.
	threshold := 5.0
	// useEmptyMasters := true

	nodes, err := c.nodeManager.allNodes()
	if err != nil {
		return err
	}
	clusterMembers, err := c.nodeManager.getClusterNodes()
	if err != nil {
		return err
	}

	masters := map[string]*Node{}
	for _, clusterNode := range clusterMembers {
		node, ok := nodes[clusterNode]
		if !ok {
			continue
		}

		if node.MySelf.Role != pkg.MasterRole {
			continue
		}

		/*
			Use empty masters is always true
			if !useEmptyMasters && node.IsEmpty {
				continue
			}
		*/
		masters[clusterNode] = &node
	}

	weights, totalWeight := getNodeWeights(masters)
	balances := map[string]int{}
	unevenCluster := false
	for _, node := range masters {
		weight, ok := weights[node.ID]
		if !ok {
			return errors.New("could not retrieve node weight")
		}
		expected := (pkg.ClusterHashSlots / totalWeight) * weight
		slots, err := node.MySelf.allSlots()
		if err != nil {
			return err
		}
		balances[node.ID] = len(slots) - int(expected)

		overTreshold := false
		if len(slots) > 0 {
			errPercent := 100 - (100.0 * expected / float64(len(slots)))
			if errPercent > threshold {
				overTreshold = true
			}
		} else if expected > 0 {
			overTreshold = true
		}
		if overTreshold {
			unevenCluster = true
		}
	}

	if !unevenCluster {
		log.Info("Cluster is balanced, skipping rebalancing")
		return nil
	}

	log.Info("Cluster is uneven balanced, initializing rebalancing tasks")

	totalBalance := 0
	for _, balance := range balances {
		totalBalance += balance
	}
	// It's possible that totalBalance is not zero because of rounding, fix this
	for totalBalance > 0 {
		for node, balance := range balances {
			if balance < 0 && totalBalance > 0 {
				balances[node]--
				totalBalance--
			}
		}
	}

	// Sort by balances
	type nodeBalance struct {
		nodeID  string
		balance int
	}
	var sbalances []nodeBalance
	for node, balance := range balances {
		sbalances = append(sbalances, nodeBalance{nodeID: node, balance: balance})
	}
	sort.Slice(sbalances, func(i, j int) bool {
		return sbalances[i].balance < sbalances[j].balance
	})

	dstIdx := 0
	srcIdx := len(sbalances) - 1
	for dstIdx < srcIdx {
		dst := sbalances[dstIdx]
		src := sbalances[srcIdx]
		var numSlots int
		if math.Abs(float64(dst.balance)) < math.Abs(float64(src.balance)) {
			numSlots = int(math.Abs(float64(dst.balance)))
		} else {
			numSlots = int(math.Abs(float64(src.balance)))
		}

		if numSlots > 0 {
			log.Infof("Moving %d slots from %s to %s", numSlots, src.nodeID, dst.nodeID)
			reshardTable, err := getReshardTable(masters[src.nodeID], numSlots)
			if err != nil {
				return err
			}
			if len(reshardTable) != numSlots {
				return errors.New("reshard table does not contain the required amount of slots")
			}
			var masterList []string
			for _, master := range masters {
				masterList = append(masterList, master.ID)
			}
			err = c.planner.NewMigrateSlotTask(src.nodeID, dst.nodeID, masters[dst.nodeID].Addr, masterList, reshardTable)
			if err != nil {
				return err
			}
		}

		// Update balances
		sbalances[dstIdx].balance += numSlots
		sbalances[srcIdx].balance -= numSlots
		if sbalances[dstIdx].balance == 0 {
			dstIdx++
		}
		if sbalances[srcIdx].balance == 0 {
			srcIdx--
		}
	}

	return nil
}
