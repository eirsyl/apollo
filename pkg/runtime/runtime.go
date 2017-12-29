package runtime

import (
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// OptimizeRuntime extends the open files limit in order to function properly.
func OptimizeRuntime() {
	var rLimit syscall.Rlimit
	e := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if e == nil {
		rLimit.Cur = 65536
		log.Infof("Setting RLimit to %d", rLimit.Cur)
		e = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if e != nil {
			log.Warn("Could not set RLimit", e)
		}
	}

	nuCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nuCPU)
	log.Infof("Running with %d CPUs", nuCPU)
}
