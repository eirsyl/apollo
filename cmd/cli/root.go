package cli

import "github.com/spf13/cobra"

func init() {

}

// CLICmd is the main cmd entrypoint for the cli
var CLICmd = &cobra.Command{
	Use:   "cli",
	Short: "Interact with the manager",
}
