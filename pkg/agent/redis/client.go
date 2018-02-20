package redis

import (
	"fmt"
	goRedis "github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

// Client exposes a set of methods used to interact with redis.
type Client struct {
	redis *goRedis.Client
}

// InstanceState contains state information about the redis instance.
type InstanceState struct {
	Up bool
}

// NewClient returns a new redis client that can be used to interact
// with the redis instance.
func NewClient(addr string) (*Client, error) {
	redis := goRedis.NewClient(&goRedis.Options{
		Addr: addr,
	})

	return &Client{
		redis: redis,
	}, nil
}

// GetAddr returns the redis listen address
func (c *Client) GetAddr() string {
	return c.redis.Options().Addr
}

// RunPreflightTests runs a set of preflight tests to make sure the
// instance is compatible with apollo.
// Checks:
// - Cluster mode enabled
func (c *Client) RunPreflightTests() error {
	log.Infof("Running preflight tests on %v", c.redis)

	var (
		scrapes = make(chan ScrapeResult)
		details []string
		wg      sync.WaitGroup
	)

	wg.Add(1)
	go func() {
		requirements := map[string]float64{
			"cluster_enabled": 1,
		}

		for scrape := range scrapes {
			if val, ok := requirements[scrape.Name]; ok && val != scrape.Value {
				details = append(details, fmt.Sprintf("%s: invalid value", scrape.Name))
			}
		}
		wg.Done()
	}()

	err := c.collectInfo(&scrapes)
	close(scrapes)
	if err != nil {
		return err
	}

	wg.Wait()
	if len(details) > 0 {
		return NewErrInstanceIncompatible(details)
	}

	return nil
}

// ScrapeInformation returns collected info from the redis instance
// the information is sent into the given channel
func (c *Client) ScrapeInformation(scrapes *chan ScrapeResult) error {
	return c.collectInfo(scrapes)
}

/*
 * Private functions
 */

func (c *Client) collectInfo(scrapes *chan ScrapeResult) error {
	config, err := c.redis.ConfigGet("*").Result()
	if err != nil {
		return err
	}
	_ = extractConfigMetrics(config, scrapes)

	info, err := c.redis.Info("ALL").Result()
	if err != nil {
		return err
	}
	_ = extractInfoMetrics(info, scrapes)

	if strings.Contains(info, "cluster_enabled:1") {
		info, err = c.redis.ClusterInfo().Result()
		if err != nil {
			return err
		}
		_ = extractInfoMetrics(info, scrapes)
	}
	return nil
}

// Shutdown closes the connection to redis.
func (c *Client) Shutdown() error {
	return c.redis.Close()
}
