package transport

import (
	"testing"
)

func TestNewTransport(t *testing.T) {
	port := ":8000"
	_, err := NewTCPTransport(port)
	if err != nil {
		t.Errorf("Failed to create server at port: %s", port)
	}
}
