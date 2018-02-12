package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eirsyl/apollo/pkg/agent"
	"github.com/eirsyl/apollo/pkg/runtime"
	"github.com/eirsyl/apollo/pkg/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	stringConfig(agentCmd, "redis", "r", "127.0.0.1:6379", "redis instance that the agent should manage")
	stringConfig(agentCmd, "manager", "m", "127.0.0.1:8080", "manager instance that the agent should report to")
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
		instanceAgent, err := agent.NewAgent()
		if err != nil {
			log.Fatalf("Could not initialize agent instance: %v", err)
		}

		// Exit
		go func() {
			for {
				s := <-exitSig
				log.Infof("Signal %s received, shutting down gracefully", s)
				go utils.ForceExit(exitSig, 5*time.Second)
				err = instanceAgent.Exit()
				if err != nil {
					log.Fatalf("Could not gracefully exit: %v", err)
					os.Exit(1)
				}
			}
		}()

		err = instanceAgent.Run()
		if err != nil {
			log.Fatalf("Agent exited unexpectedly: %v", err)
			os.Exit(1)
		}
	},
}
