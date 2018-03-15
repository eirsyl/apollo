package get

import "github.com/spf13/cobra"

func init() {

}

// GetCmd is the main cmd entrypoint for the get
var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Display one or many resources",
}
