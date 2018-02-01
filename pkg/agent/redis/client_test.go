package redis

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if client.redis == nil || err != nil {
		t.Fail()
	}
}
