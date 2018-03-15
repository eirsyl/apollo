package get

import (
	"context"

	"fmt"

	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/eirsyl/apollo/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	GetCmd.AddCommand(getStatusCmd)
}

var getStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Return cluster status",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := cli.NewCLIClient(viper.GetString("manager"))
		if err != nil {
			return err
		}

		response, err := (*client).Status(context.Background(), &pb.EmptyMessage{})
		if err != nil {
			return err
		}

		fmt.Printf(`Health: %v
State: %v`, response.Health, response.State)
		return nil
	},
}
