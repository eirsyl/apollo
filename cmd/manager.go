package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/eirsyl/apollo/pkg/runtime"
	"github.com/spf13/cobra"
)

func init() {
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

		// Exit
		go func() {
			for {
				<-exitSig
			}
		}()

	},
}
