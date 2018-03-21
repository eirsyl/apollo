package get

import (
	"context"

	"fmt"

	"time"

	"sort"
	"strings"

	tm "github.com/buger/goterm"
	"github.com/eirsyl/apollo/cmd/utils"
	pb "github.com/eirsyl/apollo/pkg/api"
	"github.com/eirsyl/apollo/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	utils.BoolConfig(getNodeCmd, "watch", "w", false, "watch nodes")
	GetCmd.AddCommand(getNodeCmd)
}

var getNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Return cluster nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := viper.GetString("manager")
		client, e := cli.NewCLIClient(manager)
		if e != nil {
			return e
		}

		node := func(clear bool) error {
			response, err := (*client).Nodes(context.Background(), &pb.EmptyMessage{})
			if err != nil {
				return err
			}

			if clear {
				tm.MoveCursor(1, 1)
				tm.Clear()
			}

			status := tm.NewTable(0, 10, 5, ' ', 0)
			if _, err = fmt.Fprintf(status, "ID\tAddr\tIsEmpty\tOnline\tNodes\tAnnotations\n"); err != nil {
				return err
			}

			normalize := func(items []string) string {
				sort.Strings(items)
				return strings.Join(items, " ")
			}

			for _, node := range response.Nodes {
				if _, err = fmt.Fprintf(status, "%s\t%s\t%t\t%t\t%d\t%s\n", node.Id, node.Addr, node.IsEmpty, node.Online, len(node.Nodes), normalize(node.HostAnnotations)); err != nil {
					return err
				}
			}

			if _, err = tm.Println(status); err != nil {
				return err
			}

			tm.Flush()
			return nil
		}

		if viper.GetBool("watch") {
			for {
				e = node(true)
				if e != nil {
					return e
				}
				time.Sleep(2 * time.Second)
			}
		} else {
			return node(false)
		}
	},
}
