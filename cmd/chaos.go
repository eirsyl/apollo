package cmd

import (
	"github.com/eirsyl/apollo/pkg/chaos"
	"github.com/eirsyl/apollo/pkg/runtime"
	"github.com/eirsyl/apollo/pkg/utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(chaosCmd)
}

var chaosCmd = &cobra.Command{
	Use:   "chaos",
	Short: "Destroy Redis instances to test the cluster manager",
	Long: `
The chaos tool destroys instances in the Redis Cluster and
watches the cleanup executed by the cluster Chaos. This
tool is used to test the cluster Chaos implementation.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if !viper.GetBool("debug") {
			log.Fatal("The chaos feature is only available when the command is ran in debug mode")
		}

		runtime.OptimizeRuntime()

		var (
			exitSig = make(chan os.Signal, 1)
		)

		signal.Notify(exitSig, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

		var err error
		instanceChaos, err := chaos.NewChaos()
		if err != nil {
			log.Fatalf("Could not initialize chaos instance: %v", err)
		}

		// Exit
		go func() {
			for {
				s := <-exitSig
				log.Infof("Signal %s received, shutting down gracefully", s)
				go utils.ForceExit(exitSig, 5*time.Second)
				err = instanceChaos.Exit()
				if err != nil {
					log.Fatalf("Could not gracefully exit: %v", err)
					os.Exit(1)
				}
			}
		}()

		err = instanceChaos.Run()
		if err != nil {
			log.Fatalf("Chaos exited unexpectedly: %v", err)
			os.Exit(1)
		}
	},
}
