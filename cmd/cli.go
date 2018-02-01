package cmd

import "github.com/eirsyl/apollo/cmd/cli"

func init() {
	RootCmd.AddCommand(cli.CLICmd)
}
