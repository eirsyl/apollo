package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	cliUtils "github.com/eirsyl/apollo/cmd/utils"
	"github.com/eirsyl/apollo/pkg/agent"
	"github.com/eirsyl/apollo/pkg/runtime"

	"github.com/eirsyl/apollo/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cliUtils.StringConfig(agentCmd, "redis", "r", "127.0.0.1:6379", "redis instance that the agent should manage")
	cliUtils.StringConfig(agentCmd, "manager", "m", "127.0.0.1:8080", "manager instance that the agent should report to")
	cliUtils.StringConfig(agentCmd, "hostAnnotations", "", "", "host annotations is used to group nodes into different failure domains. Example dc=eu-central-1 vpc=prod01")
	cliUtils.BoolConfig(agentCmd, "managerTLS", "", true, "use tls when connecting to the manager")
	cliUtils.BoolConfig(agentCmd, "skip-prechecks", "", false, "skip prechecks at startup")
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
		skipPrechecks := viper.GetBool("skip-prechecks")
		hostAnnotations := viper.GetString("hostAnnotations")

		annotations := map[string]string{}
		for _, annotation := range strings.Split(hostAnnotations, " ") {
			if split := strings.Split(annotation, "="); len(split) == 2 {
				annotations[split[0]] = split[1]
			}
		}

		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalf("Could not lookup hostname: %v", err)
		}
		annotations["hostname"] = hostname

		instanceAgent, err := agent.NewAgent(skipPrechecks, annotations)
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
