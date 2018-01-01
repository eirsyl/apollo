package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"apollo/pkg/agent"
	"apollo/pkg/runtime"
	"apollo/pkg/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(agentCmd)
}

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the instance agent functionality",
	Run: func(cmd *cobra.Command, args []string) {
		runtime.OptimizeRuntime()

		var (
			exitSig = make(chan os.Signal, 1)
		)

		signal.Notify(exitSig, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

		var err error
		instanceAgent := &agent.Agent{}

		// Exit
		go func() {
			for {
				s := <-exitSig
				log.Printf("Signal %s received, shutting down gracefully", s)
				log.Printf("Shutdown is forced in 5 sek")
				go utils.ForceExit(exitSig, 5000)
				time.Sleep(10000)
				//err = instanceAgent.Exit()
				if err != nil {
					log.Fatal(err, "Could not gracefully exit")
					os.Exit(1)
				}
			}
		}()

		err = instanceAgent.Run()
		if err != nil {
			log.Fatal(err, "Agent exited unexpectedly")
			os.Exit(1)
		}
	},
}
