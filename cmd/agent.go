package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"apollo/pkg/proxy"
	"apollo/pkg/runtime"
	"apollo/pkg/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(proxyCmd)
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Start the instance agent functionality",
	Run: func(cmd *cobra.Command, args []string) {
		runtime.OptimizeRuntime()

		var (
			exitSig = make(chan os.Signal, 1)
		)

		signal.Notify(exitSig, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

		var err error
		instanceProxy := &proxy.Proxy{}

		// Exit
		go func() {
			for {
				s := <-exitSig
				log.Printf("Signal %s received, shutting down gracefully", s)
				log.Printf("Shutdown is forced in 5 sek")
				go utils.ForceExit(exitSig, 5000)
				time.Sleep(10000)
				//err = instanceProxy.Exit()
				if err != nil {
					log.Fatal(err, "Could not gracefully exit")
					os.Exit(1)
				}
			}
		}()

		err = instanceProxy.Run()
		if err != nil {
			log.Fatal(err, "Proxy exited unexpectedly")
			os.Exit(1)
		}
	},
}
