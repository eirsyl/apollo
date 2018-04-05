package del

import (
	"github.com/spf13/cobra"
)

// DeleteCmd is the main cmd entrypoint for the del command
var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a resource",
}
