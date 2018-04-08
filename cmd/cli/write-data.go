package cli

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	CLICmd.AddCommand(writeDataCmd)
}

var writeDataCmd = &cobra.Command{
	Use:   "write-data server",
	Short: "Write data to a redis cluster",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := args[0]
		client := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{server},
		})

		for {
			key := strconv.Itoa(rand.Int())
			value := strconv.Itoa(rand.Int())
			log.Infof("Writing to redis: %s=%s", key, value)
			_, err := client.Set(key, value, 5*time.Minute).Result()
			if err != nil {
				return err
			}
		}
	},
}
