package utils

import (
	"testing"
)

func TestGetHostPort(t *testing.T) {
	h, p := GetHostPort("")
	if !(h == "localhost" && p == 6379) {
		t.Fatalf("localhost, 6379 !=  %s, %d", h, p)
	}

	h, p = GetHostPort("127.0.0.1")
	if !(h == "127.0.0.1" && p == 6379) {
		t.Fatalf("127.0.0.1, 6379 != %s, %d", h, p)
	}

	h, p = GetHostPort("127.0.0.1:7000")
	if !(h == "127.0.0.1" && p == 7000) {
		t.Fatalf("127.0.0.1, 7000 != %s, %d", h, p)
	}
}
