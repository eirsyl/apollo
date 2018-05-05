package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirsyl/apollo/pkg/chaos"
	"github.com/eirsyl/apollo/pkg/runtime"
	"github.com/eirsyl/apollo/pkg/utils"

	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(chaosCmd)
}

var chaosCmd = &cobra.Command{
	Use:   "chaos [operation]",
	Short: "Apollo test tool used to test the error detection implemented in the cluster manager",
	Long: `
Chaos introduces configuration errors in a healthy deployment of Redis Cluster. 
The tool tracks the work performed by the cluster manager in order to fix
the issue.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !viper.GetBool("debug") {
			log.Fatal("The chaos feature is only available when the command is ran in debug mode")
		}

		if len(args) < 1 {
			return errors.New("operation argument is required")
		}

		runtime.OptimizeRuntime()

		var (
			exitSig = make(chan os.Signal, 1)
		)

		signal.Notify(exitSig, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

		var err error
		chaosTest, err := chaos.NewChaos()
		if err != nil {
			log.Fatalf("Could not initialize chaos instance: %v", err)
		}

		// Exit
		go func() {
			for {
				s := <-exitSig
				log.Infof("Signal %s received, shutting down gracefully", s)
				go utils.ForceExit(exitSig, 5*time.Second)
				err = chaosTest.Exit()
				if err != nil {
					log.Fatalf("Could not gracefully exit: %v", err)
					os.Exit(1)
				}
			}
		}()

		return chaosTest.Run(args[0], args[1:])
	},
}
