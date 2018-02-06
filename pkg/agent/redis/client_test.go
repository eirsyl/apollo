package redis

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("127.0.0.1:6379")
	if client.redis == nil || err != nil {
		t.Fail()
	}
}
