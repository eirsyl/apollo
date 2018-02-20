package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/eirsyl/apollo/pkg/manager"
	"github.com/eirsyl/apollo/pkg/runtime"
	"github.com/eirsyl/apollo/pkg/utils"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	stringConfig(managerCmd, "managerAddr", "", ":8080", "manger listen address")
	stringConfig(managerCmd, "debugAddr", "", ":8081", "debug server listen address")
	stringConfig(managerCmd, "databaseFile", "", "apollo.db", "database path for internal state")
	RootCmd.AddCommand(managerCmd)
}

var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start the cluster manager functionality",
	Run: func(cmd *cobra.Command, args []string) {
		runtime.OptimizeRuntime()

		var (
			exitSig = make(chan os.Signal, 1)
		)

		signal.Notify(exitSig, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

		managerAddr := viper.GetString("managerAddr")
		httpAddr := viper.GetString("debugAddr")
		databaseFile := viper.GetString("databaseFile")
		instanceManager, err := manager.NewManager(managerAddr, httpAddr, databaseFile)
		if err != nil {
			log.Fatalf("Could not initialize manager instance: %v", err)
		}

		// Exit
		go func() {
			for {
				s := <-exitSig
				log.Infof("Signal %s received, shutting down gracefully", s)
				go utils.ForceExit(exitSig, 5*time.Second)
				err = instanceManager.Exit()
				if err != nil {
					log.Fatalf("Could not gracefully exit: %v", err)
					os.Exit(1)
				}
			}
		}()

		err = instanceManager.Run()
		if err != nil {
			log.Fatalf("Manager exited unexpectedly: %v", err)
			os.Exit(1)
		}

	},
}
