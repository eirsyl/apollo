package utils

import (
	"os"
	"time"
)

// ForceExit the application when a signal is sent
func ForceExit(exitSig chan os.Signal, maxDelay time.Duration) {
	go func() {
		time.Sleep(maxDelay)
		os.Exit(1)
	}()

	<-exitSig
	os.Exit(1)
}
