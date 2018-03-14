package redis

import (
	"fmt"
	"strings"
	"sync"

	"strconv"

	"github.com/eirsyl/apollo/pkg/utils"
	goRedis "github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// Client exposes a set of methods used to interact with redis.
type Client struct {
	redis *goRedis.Client
}

// NodeState contains state information about the redis node.
type NodeState struct {
	Up bool
}

// NewClient returns a new redis client that can be used to interact
// with the redis node.
func NewClient(addr string) (*Client, error) {
	redis := goRedis.NewClient(&goRedis.Options{
		Addr: addr,
	})

	return &Client{
		redis: redis,
	}, nil
}

// GetAddr returns the redis listen address
func (c *Client) GetAddr() string {
	return c.redis.Options().Addr
}

// RunPreflightTests runs a set of preflight tests to make sure the
// node is compatible with apollo.
// Checks:
// - Cluster mode enabled
func (c *Client) RunPreflightTests() error {
	log.Infof("Running preflight tests on %v", c.redis)

	var (
		scrapes = make(chan ScrapeResult)
		details []string
		wg      sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		requirements := map[string]float64{
			"cluster_enabled": 1,
		}

		for scrape := range scrapes {
			if val, ok := requirements[scrape.Name]; ok && val != scrape.Value {
				details = append(details, fmt.Sprintf("%s: invalid value", scrape.Name))
			}
		}
		wg.Done()
	}()

	err := c.collectInfo(&scrapes)
	close(scrapes)
	if err != nil {
		return err
	}

	wg.Wait()
	if len(details) > 0 {
		return NewErrNodeIncompatible(details)
	}

	return nil
}

// ScrapeInformation returns collected info from the redis node
// the information is sent into the given channel
func (c *Client) ScrapeInformation(scrapes *chan ScrapeResult) error {
	return c.collectInfo(scrapes)
}

// IsEmpty check if the node is empty or attached to an existing cluster
// The node is empty if no database exists and the instance is'nt aware of other cluster members
func (c *Client) IsEmpty() (bool, error) {
	// info keyspace !contains db0 && cluster info contains cluster_known_nodes:1
	ik, err := c.redis.Info("keyspace").Result()
	if err != nil {
		return false, err
	}

	if strings.Contains(ik, "db0") {
		return false, nil
	}

	cm, err := c.redis.ClusterInfo().Result()
	if err != nil {
		return false, err
	}

	if !strings.Contains(cm, "cluster_known_nodes:1") {
		return false, nil
	}

	return true, nil
}

// ClusterNodes fetches the cluster topology known by the node.
func (c *Client) ClusterNodes() ([]ClusterNode, error) {
	var nodes []ClusterNode

	nodeInfo, err := c.redis.ClusterNodes().Result()
	if err != nil {
		return []ClusterNode{}, err
	}

	for _, l := range strings.Split(nodeInfo, "\n") {
		args := strings.Split(l, " ")
		if len(args) == 1 {
			continue
		} else if len(args) < 8 {
			log.Warn("Cluster node information retrieved from redis is invalid: %v", args)
			continue
		}

		flags := args[2]
		role := "slave"
		if strings.Contains(flags, "master") && args[3] == "-" {
			role = "master"
		}

		pingRecv, err := strconv.Atoi(args[4])
		if err != nil {
			log.Warn("Could not parse PingRecv: %v", args[4])
		}

		pingSent, err := strconv.Atoi(args[5])
		if err != nil {
			log.Warn("Could not parse PingSent: %v", args[5])
		}

		configEpoch, err := strconv.Atoi(args[6])
		if err != nil {
			log.Warn("Could not parse ConfigEpoch: %v", args[6])
		}

		var slots []string
		if len(args) > 8 {
			slots = args[8:]
		}

		node := ClusterNode{
			NodeID:      args[0],
			Addr:        args[1],
			Flags:       flags,
			Role:        role,
			Myself:      strings.Contains(flags, "myself"),
			MasterID:    args[3],
			PingRecv:    pingRecv,
			PingSent:    pingSent,
			ConfigEpoch: configEpoch,
			LinkStatus:  args[7],
			Slots:       slots,
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// AddSlots marks the redis instance responsible for a list if slots.
func (c *Client) AddSlots(slots []int) (string, error) {
	log.Infof("Adding cluster slots: %v", slots)
	return c.redis.ClusterAddSlots(slots...).Result()
}

// JoinCluster
func (c *Client) JoinCluster(addr string) (string, error) {
	log.Infof("Joining cluster: %v", addr)
	host, p := utils.GetHostPort(addr)
	port := strconv.Itoa(p)
	return c.redis.ClusterMeet(host, port).Result()
}

// Replicate configures a slave to replicate a master
func (c *Client) Replicate(nodeID string) (string, error) {
	log.Infof("Setting replication target: %v", nodeID)
	return c.redis.ClusterReplicate(nodeID).Result()
}

// SetEpoch configures node cluster epoch
func (c *Client) SetEpoch(epoch int) (string, error) {
	log.Infof("Setting cluster epoch: %v", epoch)
	return "", nil
}

/*
 * Private functions
 */

func (c *Client) collectInfo(scrapes *chan ScrapeResult) error {
	config, err := c.redis.ConfigGet("*").Result()
	if err != nil {
		return err
	}
	_ = extractConfigMetrics(config, scrapes)

	info, err := c.redis.Info("ALL").Result()
	if err != nil {
		return err
	}
	_ = extractInfoMetrics(info, scrapes)

	if strings.Contains(info, "cluster_enabled:1") {
		info, err = c.redis.ClusterInfo().Result()
		if err != nil {
			return err
		}
		_ = extractInfoMetrics(info, scrapes)
	}
	return nil
}

// Shutdown closes the connection to redis.
func (c *Client) Shutdown() error {
	return c.redis.Close()
}
