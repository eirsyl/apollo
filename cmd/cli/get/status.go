package get

import (
	"context"

	"fmt"

	"time"

	tm "github.com/buger/goterm"
	"github.com/eirsyl/apollo/cmd/utils"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/eirsyl/apollo/pkg/cli"
	"github.com/eirsyl/apollo/pkg/manager/orchestrator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	utils.BoolConfig(getStatusCmd, "watch", "w", false, "watch status")
	GetCmd.AddCommand(getStatusCmd)
}

var getStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Return cluster status",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := viper.GetString("manager")
		client, e := cli.NewCLIClient(manager)
		if e != nil {
			return e
		}

		status := func(clear bool) error {
			response, err := (*client).Status(context.Background(), &pb.EmptyMessage{})
			if err != nil {
				return err
			}

			if clear {
				tm.MoveCursor(1, 1)
				tm.Clear()
			}

			status := tm.NewTable(0, 10, 5, ' ', 0)
			fmt.Fprintf(status, "Manager\tHealth\tState\n")
			fmt.Fprintf(status, "%s\t%s\t%s\n", manager, orchestrator.HumanizeClusterHealth(response.Health), orchestrator.HumanizeClusterState(response.State))
			_, err = tm.Println(status)
			if err != nil {
				return err
			}
			tm.Flush()
			return nil
		}

		if viper.GetBool("watch") {
			for {
				e = status(true)
				if e != nil {
					return e
				}
				time.Sleep(2 * time.Second)
			}
		} else {
			return status(false)
		}
	},
}
