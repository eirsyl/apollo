package cmd

import (
	"fmt"
	"github.com/eirsyl/apollo/pkg"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug mode")
	err := viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		log.Fatalf("Could not bind flag: %v", err)
	}
	err = viper.BindEnv("debug", "DEBUG")
	if err != nil {
		log.Fatalf("Could not bind config from env: %v", err)
	}
}

// RootCmd is the main entrypoint for the CLI application.
var RootCmd = &cobra.Command{
	Use:     "apollo",
	Short:   "Apollo is a cluster manager running as a sidecar in your Redis Cluster deployment.",
	Version: fmt.Sprintf("%s %s", pkg.Version, pkg.BuildDate),
	Long: `
Apollo is a Redis Cluster manager that aims to lighten the operational burden
on cluster operators. The cluster manager watches the Redis cluster for possible
issues or reduced performance and tries to fix these in the best possible way.
`,
}
