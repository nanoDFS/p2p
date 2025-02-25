package transport

import "testing"

func TestNewTransport(t *testing.T) {
	_, err := NewTransport(":8000")
	if err != nil {
		t.Error("Failed to create")
	}
}
