package cli

import (
	"github.com/eirsyl/apollo/cmd/cli/get"
)

func init() {
	CLICmd.AddCommand(get.GetCmd)
}
