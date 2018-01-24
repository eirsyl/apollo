package utils

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

// ForceExit the application when a signal is sent
func ForceExit(exitSig chan os.Signal, maxDelay time.Duration) {
	go func() {
		log.Warnf("Shutdown is forced in %s", maxDelay)
		time.Sleep(maxDelay)
		os.Exit(1)
	}()

	<-exitSig
	log.Warnf("Shutdown forced, exiting")
	os.Exit(1)
}
