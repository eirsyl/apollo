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

func extractConfigMetrics(config []interface{}, scrapes *chan ScrapeResult) error {
	if len(config)%2 != 0 {
		return errors.New("Invalid redis config")
	}

	log.Debug("Extracting config information")

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
			*scrapes <- ScrapeResult{
				Name:  fmt.Sprintf("config_%s", strKey),
				Value: val,
			}
		}
	}
	return nil
}

func extractInfoMetrics(info string, scrapes *chan ScrapeResult) error {
	log.Debug("Extracting info information")

	lines := strings.Split(info, "\r\n")
	for _, l := range lines {
		name, value, err := extractVal(l, ":")
		if err != nil {
			// Try to parse the keyspace value
			split := strings.Split(l, ":")
			if len(split) == 2 {
				if keysTotal, keysEx, avgTTL, ok := parseDBKeyspaceString(split[0], split[1]); ok {
					*scrapes <- ScrapeResult{Name: "db_keys", Value: keysTotal}
					*scrapes <- ScrapeResult{Name: "db_keys_expiring", Value: keysEx}
					if avgTTL > -1 {
						*scrapes <- ScrapeResult{Name: "db_avg_ttl_seconds", Value: avgTTL}
					}
					continue
				}
			}

			log.Debugf("Could not parse: %v", err)
			continue
		}
		*scrapes <- ScrapeResult{Name: name, Value: value}
	}
	return nil
}

func parseDBKeyspaceString(db string, stats string) (keysTotal float64, keysExpiringTotal float64, avgTTL float64, ok bool) {
	ok = false
	if !strings.HasPrefix(db, "db") {
		return
	}

	split := strings.Split(stats, ",")
	if len(split) != 3 {
		return
	}

	var err error
	ok = true
	if _, keysTotal, err = extractVal(split[0], "="); err != nil {
		ok = false
		return
	}
	if _, keysExpiringTotal, err = extractVal(split[1], "="); err != nil {
		ok = false
		return
	}

	avgTTL = -1
	if len(split) > 2 {
		if _, avgTTL, err = extractVal(split[2], "="); err != nil {
			ok = false
			return
		}
		avgTTL /= 1000
	}

	return
}
