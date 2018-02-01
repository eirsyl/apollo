package redis

import (
	goRedis "github.com/go-redis/redis"
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
func NewClient() (*Client, error) {
	redis := goRedis.NewClient(&goRedis.Options{
		Addr: "localhost:6379",
	})

	return &Client{
		redis: redis,
	}, nil
}

// Health returns the instance health.
func (c *Client) Health() (InstanceHealth, error) {
	return InstanceHealth{}, nil
}
