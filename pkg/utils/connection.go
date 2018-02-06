package utils

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetHostPort returns the host and port separately from a string on the format host:port
func GetHostPort(addr string) (string, int) {
	if addr == "" {
		return "localhost", 6379
	}

	a := strings.Split(addr, ":")
	if len(a) == 1 {
		return a[0], 6379
	}

	p, err := strconv.Atoi(a[1])
	if err != nil {
		log.Warnf("Could not parse address port: %v", err)
		p = 6379
	}
	return a[0], p

}
