package redis

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func extractVal(s string, separator string) (string, float64, error) {
	split := strings.Split(s, separator)
	if len(split) != 2 {
		return "", 0, fmt.Errorf("Cannot extract value: %s", s)

	}
	val, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		return "", 0, fmt.Errorf("Cannot extract value: %s", s)
	}
	return split[0], val, err
}

func extractConfigMetrics(config []interface{}, scrapes *chan scrapeResult) error {
	if len(config)%2 != 0 {
		return errors.New("Invalid redis config")
	}

	for pos := 0; pos < len(config)/2; pos++ {
		strKey := config[pos*2].(string)
		strVal := config[pos*2+1].(string)

		// This map contains config values that our system is interested in
		if !map[string]bool{
			"maxmemory": true,
		}[strKey] {
			continue
		}

		if val, err := strconv.ParseFloat(strVal, 64); err == nil {
			*scrapes <- scrapeResult{
				Name:  fmt.Sprintf("config_%s", strKey),
				Value: val,
			}
		}
	}
	return nil
}

func extractInfoMetrics(info string, scrapes *chan scrapeResult) error {
	lines := strings.Split(info, "\r\n")
	for _, l := range lines {
		name, value, err := extractVal(l, ":")
		if err != nil {
			log.Debugf("Could not parse: %v", err)
			continue
		}
		*scrapes <- scrapeResult{Name: name, Value: value}
	}
	return nil
}
