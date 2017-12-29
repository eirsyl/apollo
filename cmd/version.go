package cmd

import (
	"apollo/pkg"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Apollo version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Apollo cluster manager version %s\n", pkg.Version)
	},
}
