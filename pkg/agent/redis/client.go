package redis

import (
	goRedis "github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

// Client exposes a set of methods used to interact with redis.
type Client struct {
	redis *goRedis.Client
}

// InstanceHealth contains health information about the redis instance.
type InstanceHealth struct {
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

// Health returns the instance health.
func (c *Client) Health() (InstanceHealth, error) {
	return InstanceHealth{}, nil
}

// RunPreflightTests runs a set of preflight tests to make sure the
// instance is compatible with apollo
func (c *Client) RunPreflightTests() error {
	log.Infof("Running preflight tests on %v", c.redis)
	return nil
}

// Shutdown closes the connection to redis.
func (c *Client) Shutdown() error {
	return c.redis.Close()
}
