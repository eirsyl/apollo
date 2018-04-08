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
	CLICmd.AddCommand(forgetSlotCmd)
}

var forgetSlotCmd = &cobra.Command{
	Use:   "forget-slot slot 6000...6005 7000",
	Short: "Forget slot, nodes has to be located at 127.0.0.1",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		slotArg := args[0]
		slot, err := strconv.Atoi(slotArg)
		if err != nil {
			return err
		}
		for _, arg := range args[1:] {

			s := strings.Split(arg, "...")
			if len(s) == 1 {
				client := redis.NewClient(&redis.Options{
					Addr: fmt.Sprintf("127.0.0.1:%s", s),
				})
				_, err := client.ClusterDelSlots(slot).Result()
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
				_, err := client.ClusterDelSlots(slot).Result()
				if err != nil {
					return err
				}
			}

		}
		return nil
	},
}
