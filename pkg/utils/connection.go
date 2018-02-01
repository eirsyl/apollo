package utils

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetHostPort(addr string) (string, int) {
	if addr == "" {
		return "localhost", 6379
	}

	a := strings.Split(addr, ":")
	if len(a) == 0 {
		return "localhost", 6379
	} else if len(a) == 1 {
		return a[0], 6379
	} else {
		p, err := strconv.Atoi(a[1])
		if err != nil {
			log.Warnf("Could not parse address port: %v", err)
			p = 6379
		}
		return a[0], p
	}

}
