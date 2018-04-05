package del

import (
	"context"
	"fmt"

	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/eirsyl/apollo/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	DeleteCmd.AddCommand(deleteNodeCmd)
}

var deleteNodeCmd = &cobra.Command{
	Use:   "node [nodeID]",
	Short: "Remove a node from the cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := viper.GetString("manager")
		client, e := cli.NewCLIClient(manager)
		if e != nil {
			return e
		}

		nodeID := args[0]
		response, err := (*client).DeleteNode(context.Background(), &pb.DeleteNodeRequest{
			NodeID: nodeID,
		})

		if err != nil {
			return err
		}

		if response.Success {
			fmt.Printf("Node %s is marked for removal, the cluster manager is going to remove this node from the cluster when the cluster us ready.", nodeID)
		} else {
			fmt.Printf("Could not mark node %s for deletion", nodeID)
		}

		return nil
	},
}
