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
	"github.com/eirsyl/apollo/pkg/manager/orchestrator/planner"
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
			if _, err = fmt.Fprintf(status, "Manager\tHealth\tState\n"); err != nil {
				return err
			}
			var health orchestrator.ClusterHealth
			var state orchestrator.ClusterState
			{
				health = orchestrator.ClusterHealth(response.Health)
				state = orchestrator.ClusterState(response.State)
			}
			if _, err = fmt.Fprintf(status, "%s\t%s\t%s\n", manager, health, state); err != nil {
				return err
			}
			if _, err = tm.Println(status); err != nil {
				return err
			}

			tasks := tm.NewTable(0, 10, 5, ' ', 0)
			commands := tm.NewTable(0, 10, 5, ' ', 0)

			if _, err = fmt.Fprintf(tasks, "Task ID\tType\tStatus\n"); err != nil {
				return err
			}
			if _, err = fmt.Fprintf(commands, "Task ID\tCommand ID\tType\tStatus\tNode\tRetries\tDependencies\n"); err != nil {
				return err
			}

			for _, task := range response.Tasks {
				var taskType planner.TaskType
				var taskStatus planner.TaskStatus
				{
					taskType = planner.TaskType(task.Type)
					taskStatus = planner.TaskStatus(task.Status)
				}
				if _, err = fmt.Fprintf(tasks, "%d\t%s\t%s\n", task.Id, taskType, taskStatus); err != nil {
					return err
				}

				var commandType planner.CommandType
				var commandStatus planner.CommandStatus
				for _, command := range task.Commands {
					{
						commandType = planner.CommandType(command.Type)
						commandStatus = planner.CommandStatus(command.Status)
					}
					if _, err = fmt.Fprintf(commands, "%d\t%s\t%s\t%s\t%s\t%d\t%d\n", task.Id, command.Id, commandType, commandStatus, command.NodeID, command.Retries, command.Dependencies); err != nil {
						return err
					}
				}
			}

			if _, err = tm.Println(tasks); err != nil {
				return err
			}

			if _, err = tm.Println(commands); err != nil {
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
