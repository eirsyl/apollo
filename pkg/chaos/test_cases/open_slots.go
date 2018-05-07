package test_cases

import (
	"errors"

	"strings"

	"time"

	"strconv"

	"regexp"

	"sync"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type OpenSlots struct {
}

func NewOpenSlots() (*OpenSlots, error) {
	return &OpenSlots{}, nil
}

func (wd *OpenSlots) GetName() string {
	return "open-slots"
}

func (wd *OpenSlots) Run(args []string) error {
	if len(args) != 2 {
		return errors.New("please provide the following flags: [redis-node] [slot-to-manage]")
	}

	server := args[0]
	slotToManage := args[1]
	slotAsInt, err := strconv.Atoi(slotToManage)
	if err != nil {
		return err
	}

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{server},
	})

	// Masters
	log.Infof("Finding masters")
	type master struct {
		addr string
		id   string
	}
	var masters []master
	res, err := client.ClusterNodes().Result()
	if err != nil {
		return err
	}
	for _, l := range strings.Split(res, "\n") {
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

		if role != "master" {
			continue
		}

		addr := strings.Split(args[1], "@")[0]

		m := master{
			id:   args[0],
			addr: addr,
		}
		masters = append(masters, m)
	}

	// Slot owner
	log.Infof("Finding slot nodes")
	var slotNodes = map[string]bool{}
	slots, err := client.ClusterSlots().Result()
	if err != nil {
		return err
	}
	for _, s := range slots {
		slot := s.Start
		for slot <= s.End {
			if slot == slotAsInt {
				for _, node := range s.Nodes {
					slotNodes[node.Addr] = true
				}
				break
			}
			slot += 1
		}
	}
	if len(slotNodes) == 0 {
		return errors.New("could not find slot hosts")
	}
	log.Info("Finding slot master")
	var slotMaster string
	var slotMasterId string
	for _, master := range masters {
		if slotNodes[master.addr] {
			slotMaster = master.addr
			slotMasterId = master.id
		}
	}
	if slotMaster == "" {
		return errors.New("no slot master found")
	}

	// Non slot master
	log.Info("Find random other master")
	var slotDestination string
	for _, node := range masters {
		if node.addr != slotMaster {
			slotDestination = node.id
			break
		}
	}
	if slotDestination == "" {
		return errors.New("no slot destination")
	}

	// CASE 1: Slot is in migrating state on one node and importing on another node.
	startTime := time.Now().UTC()
	lock := sync.Mutex{}
	destinationSet := false
	err = client.ForEachMaster(func(c *redis.Client) error {
		lock.Lock()
		defer lock.Unlock()
		var cmd *redis.StringCmd
		if c.Options().Addr == slotMaster {
			cmd = redis.NewStringCmd("CLUSTER", "SETSLOT", slotToManage, "MIGRATING", slotDestination)
		} else {
			if destinationSet {
				return nil
			}
			cmd = redis.NewStringCmd("CLUSTER", "SETSLOT", slotToManage, "IMPORTING", slotMasterId)
			destinationSet = true
		}
		err = c.Process(cmd)
		if err != nil {
			return err
		}

		_, err = cmd.Result()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = watchOpenSlots(client)
	if err != nil {
		return err
	}

	log.Infof("Slot is in migrating state on one node and importing on another node: %v", time.Since(startTime))
	time.Sleep(30 * time.Second)

	// Masters
	log.Infof("Finding masters")
	masters = []master{}
	res, err = client.ClusterNodes().Result()
	if err != nil {
		return err
	}
	for _, l := range strings.Split(res, "\n") {
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

		if role != "master" {
			continue
		}

		addr := strings.Split(args[1], "@")[0]

		m := master{
			id:   args[0],
			addr: addr,
		}
		masters = append(masters, m)
	}

	// Slot owner
	log.Infof("Finding slot nodes")
	slotNodes = map[string]bool{}
	slots, err = client.ClusterSlots().Result()
	if err != nil {
		return err
	}
	for _, s := range slots {
		slot := s.Start
		for slot <= s.End {
			if slot == slotAsInt {
				for _, node := range s.Nodes {
					slotNodes[node.Addr] = true
				}
				break
			}
			slot += 1
		}
	}
	if len(slotNodes) == 0 {
		return errors.New("could not find slot hosts")
	}
	log.Info("Finding slot master")
	for _, master := range masters {
		if slotNodes[master.addr] {
			slotMaster = master.addr
			slotMasterId = master.id
		}
	}
	if slotMaster == "" {
		return errors.New("no slot master found")
	}

	// Non slot master
	log.Info("Find random other master")
	for _, node := range masters {
		if node.addr != slotMaster {
			slotDestination = node.id
			break
		}
	}
	if slotDestination == "" {
		return errors.New("no slot destination")
	}

	// CASE 2: Multiple nodes has the slot in importing state.
	startTime = time.Now().UTC()
	err = client.ForEachMaster(func(c *redis.Client) error {
		if c.Options().Addr == slotMaster {
			return nil
		}

		cmd := redis.NewStringCmd("CLUSTER", "SETSLOT", slotToManage, "IMPORTING", slotMasterId)
		err = c.Process(cmd)
		if err != nil {
			return err
		}

		_, err = cmd.Result()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = watchOpenSlots(client)
	if err != nil {
		return err
	}

	log.Infof("Multiple nodes with slot in importing state: %v", time.Since(startTime))
	time.Sleep(30 * time.Second)

	// Masters
	log.Infof("Finding masters")
	masters = []master{}
	res, err = client.ClusterNodes().Result()
	if err != nil {
		return err
	}
	for _, l := range strings.Split(res, "\n") {
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

		if role != "master" {
			continue
		}

		addr := strings.Split(args[1], "@")[0]

		m := master{
			id:   args[0],
			addr: addr,
		}
		masters = append(masters, m)
	}

	// Slot owner
	log.Infof("Finding slot nodes")
	slotNodes = map[string]bool{}
	slots, err = client.ClusterSlots().Result()
	if err != nil {
		return err
	}
	for _, s := range slots {
		slot := s.Start
		for slot <= s.End {
			if slot == slotAsInt {
				for _, node := range s.Nodes {
					slotNodes[node.Addr] = true
				}
				break
			}
			slot += 1
		}
	}
	if len(slotNodes) == 0 {
		return errors.New("could not find slot hosts")
	}
	log.Info("Finding slot master")
	for _, master := range masters {
		if slotNodes[master.addr] {
			slotMaster = master.addr
			slotMasterId = master.id
		}
	}
	if slotMaster == "" {
		return errors.New("no slot master found")
	}

	// Non slot master
	log.Info("Find random other master")
	for _, node := range masters {
		if node.addr != slotMaster {
			slotDestination = node.id
			break
		}
	}
	if slotDestination == "" {
		return errors.New("no slot destination")
	}

	// CASE 3: One empty node with the slot in migrating state.
	startTime = time.Now().UTC()
	err = client.ForEachMaster(func(c *redis.Client) error {
		if c.Options().Addr == slotMaster {
			_, err = c.FlushAll().Result()
			if err != nil {
				return err
			}
			cmd := redis.NewStringCmd("CLUSTER", "SETSLOT", slotToManage, "MIGRATING", slotDestination)
			err = c.Process(cmd)
			if err != nil {
				return err
			}
			_, err = cmd.Result()
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = watchOpenSlots(client)
	if err != nil {
		return err
	}

	log.Infof("One empty node with slot in importing state: %v", time.Since(startTime))

	return nil
}

func watchOpenSlots(c *redis.ClusterClient) error {
	for {
		allClosed := true
		err := c.ForEachNode(func(c *redis.Client) error {
			members, err := c.ClusterNodes().Result()
			if err != nil {
				return err
			}

			reg := regexp.MustCompile(`^\[(?P<slot>\d+)\-(?P<operation>[<|>])\-(?P<destination>\w+)\]$`)
			for _, l := range strings.Split(members, "\n") {
				var slots []string
				args := strings.Split(l, " ")
				if len(args) > 8 {
					slots = args[8:]
				}
				for _, slot := range slots {
					match := reg.FindStringSubmatch(slot)
					if len(match) == 4 {
						allClosed = false
					}
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		if allClosed == true {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
