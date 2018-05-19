package testcases

import (
	"errors"

	"time"

	"strconv"

	"fmt"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// SlotCoverage represents the slot coverage test case
type SlotCoverage struct {
}

// NewSlotCoverage creates a new slot coverage test case
func NewSlotCoverage() (*SlotCoverage, error) {
	return &SlotCoverage{}, nil
}

// GetName returns the name of the clot coverage test case
func (wd *SlotCoverage) GetName() string {
	return "slot-coverage"
}

// Run performs the slot coverage test case
func (wd *SlotCoverage) Run(args []string) error {
	if len(args) != 2 {
		return errors.New("please provide the following flags: [redis-node] [slot-to-delete]")
	}

	server := args[0]
	slotToDelete, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{server},
	})

	// Delete slots on all nodes (One node has data is slot, preloaded with write-data)
	delSlot := func(c *redis.Client) error {
		_, delSlotErr := c.ClusterDelSlots(slotToDelete).Result()
		return delSlotErr
	}
	startTime := time.Now().UTC()
	err = client.ForEachNode(delSlot)
	if err != nil {
		return err
	}

	// Watch for slot allocation by the manager
	err = watchSlots(client, slotToDelete)
	if err != nil {
		return err
	}
	log.Infof("One node has keys belonging to the open slot: %v", time.Since(startTime))

	// Delete slots on all nodes (Multiple nodes have data belonging to slot)
	key := "test"
	keyItr := 1
	for {
		var slot int64
		slot, err = client.ClusterKeySlot(key).Result()
		if err != nil {
			return err
		}
		if int(slot) == slotToDelete {
			break
		} else {
			key = fmt.Sprintf("test_%d", keyItr)
			keyItr++
		}
	}
	startTime = time.Now().UTC()
	err = client.ForEachNode(delSlot)
	if err != nil {
		return err
	}

	err = client.ForEachMaster(func(c *redis.Client) error {
		_, err = c.ClusterAddSlots(slotToDelete).Result()
		if err != nil {
			return err
		}
		cmd := redis.NewStringCmd("cluster", "setslot", slotToDelete, "stable")
		err = c.Process(cmd)
		if err != nil {
			return err
		}
		_, err = cmd.Result()
		if err != nil {
			return err
		}
		_, err = c.Set(key, "test-value", 0).Result()
		if err != nil {
			return err
		}

		_, err = c.ClusterDelSlots(slotToDelete).Result()
		return err
	})
	if err != nil {
		return err
	}

	err = watchSlots(client, slotToDelete)
	if err != nil {
		return err
	}
	log.Infof("Multiple nodes has keys belonging to the open slot: %v", time.Since(startTime))
	time.Sleep(30 * time.Second)

	// Delete all keys belonging to slot
	deleteKeys := func(c *redis.Client) error {
		_, err = c.FlushAll().Result()
		return err
	}

	// Delete slots on all nodes (No nodes has data is slot)
	err = client.ForEachMaster(deleteKeys)
	if err != nil {
		return err
	}

	startTime = time.Now().UTC()
	err = client.ForEachNode(delSlot)
	if err != nil {
		return err
	}

	// Watch for slot allocation by the manager
	err = watchSlots(client, slotToDelete)
	if err != nil {
		return err
	}
	log.Infof("No node has keys belonging to the open slot: %v", time.Since(startTime))

	return nil
}

func watchSlots(c *redis.ClusterClient, slot int) error {
	for {
		allKnows := true
		err := c.ForEachNode(func(c *redis.Client) error {
			hasSlot := false
			res, err := c.ClusterSlots().Result()
			if err != nil {
				return err
			}

			for _, alloc := range res {
				s := alloc.Start
				for s <= alloc.End {
					if s == slot {
						hasSlot = true
						break
					}
					s++
				}
			}

			if !hasSlot {
				allKnows = false
			}
			return nil
		})
		if err != nil {
			return err
		}
		if allKnows {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
