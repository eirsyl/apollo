package cli

import (
	"github.com/eirsyl/apollo/cmd/cli/del"
)

func init() {
	CLICmd.AddCommand(del.DeleteCmd)
}
