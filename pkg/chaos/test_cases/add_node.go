package test_cases

import (
	"os/exec"
	"strings"
	"time"

	"errors"
	"strconv"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type AddNode struct {
}

func NewAddNode() (*AddNode, error) {
	return &AddNode{}, nil
}

func (wd *AddNode) GetName() string {
	return "add-node"
}

func (wd *AddNode) Run(args []string) error {
	if len(args) != 1 {
		return errors.New("please provide the following flags: [new-cluster-size]")
	}

	clusterSize := args[0]
	size, err := strconv.Atoi(clusterSize)
	if err != nil {
		return err
	}

	serverCmd := "redis-server --port 6100 --cluster-enabled yes --dbfilename node100.rdb --cluster-config-file node100.conf"
	agentCmd := "bin/apollo agent --redis=127.0.0.1:6100 --managerTLS=false"

	// Run services
	go func() {
		err := runCmd(serverCmd)
		if err != nil {
			log.Warnf("server error: %v", err)
		}
	}()
	go func() {
		err := runCmd(agentCmd)
		if err != nil {
			log.Warnf("agent error: %v", err)
		}
	}()

	client := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6100",
	})

	var startTime time.Time
	for {
		_, err := client.Ping().Result()
		if err == nil {
			startTime = time.Now().UTC()
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	for {
		nodes, err := client.ClusterNodes().Result()
		if err != nil {
			return err
		}

		n := len(strings.Split(nodes, "\n")) - 1
		log.Infof("desired size: %d, actual size: %d", size, n)

		if n == size {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Infof("Empty node join cluster: %v", time.Since(startTime))

	return nil
}

func runCmd(command string) error {
	run := strings.Split(command, " ")
	cmd := exec.Command(run[0], run[1:]...)
	return cmd.Run()
}
