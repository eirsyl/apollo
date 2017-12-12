package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd is the main entrypoint for the CLI application.
var RootCmd = &cobra.Command{
	Use:   "apollo",
	Short: "Apollo is a cluster manager running as a sidecar in your Redis Cluster deployment.",
	Long: `
Apollo is a Redis Cluster manager that aims to lighten the operational burden
on cluster operators. The cluster manager watches the Redis cluster for possible
issues or reduced performance and tries to fix these in the best possible way.
`,
}
