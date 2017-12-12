package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	err := viper.BindEnv("debug", "DEBUG")
	if err != nil {
		log.Info("Could not bind env")
	}
	if viper.GetBool("debug") {
		RootCmd.AddCommand(chaosCmd)
	}
}

var chaosCmd = &cobra.Command{
	Use:   "chaos",
	Short: "Destroy Redis instances to test the cluster manager",
	Long: `
The chaos tool destroys instances in the Redis Cluster and
watches the cleanup executed by the cluster manager. This
tool is used to test the cluster manager implementation.
`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
