package cli

import (
	"github.com/eirsyl/apollo/cmd/utils"
	"github.com/spf13/cobra"
)

func init() {
	utils.StringConfig(CLICmd, "manager", "m", "127.0.0.1:8080", "manager instance address")
}

// CLICmd is the main cmd entrypoint for the cli
var CLICmd = &cobra.Command{
	Use:   "cli",
	Short: "Interact with the manager",
}
