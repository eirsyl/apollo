package cli

import (
	"strings"

	"fmt"

	"errors"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
)

func init() {
	CLICmd.AddCommand(resetNodesCmd)
}

var resetNodesCmd = &cobra.Command{
	Use:   "reset-nodes 6000...6005 7000",
	Short: "Reset node cluster, nodes has to be located at 127.0.0.1",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {

			s := strings.Split(arg, "...")
			if len(s) == 1 {
				client := redis.NewClient(&redis.Options{
					Addr: fmt.Sprintf("127.0.0.1:%s", s),
				})
				_, err := client.ClusterResetHard().Result()
				return err
			}

			start, err := strconv.Atoi(s[0])
			if err != nil {
				return err

			}
			stop, err := strconv.Atoi(s[1])
			if err != nil {
				return err
			}

			if start > stop {
				return errors.New("start has to be lower than stop")
			}

			for i := start; i <= stop; i++ {
				client := redis.NewClient(&redis.Options{
					Addr: fmt.Sprintf("127.0.0.1:%d", i),
				})
				_, err := client.ClusterResetHard().Result()
				if err != nil {
					return err
				}
			}

		}
		return nil
	},
}
