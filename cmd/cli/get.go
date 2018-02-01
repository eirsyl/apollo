package cli

import "github.com/spf13/cobra"

func init() {
	CLICmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display one or many resources",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
