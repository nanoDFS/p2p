package transport

import (
	"testing"
	"time"

	"github.com/nanoDFS/p2p/p2p/encoder"
)

func getServer(addr string) (*TCPTransport, error) {
	return NewTCPTransport(addr)
}

func TestListen(t *testing.T) {
	port := ":8080"
	server, _ := getServer(port)
	if err := server.Listen(); err != nil {
		t.Errorf("failed to listen, %s, %v", port, err)
	}
}

func TestSend(t *testing.T) {
	port := ":8090"
	server, _ := getServer(port)
	server.Listen()
	node, _ := NewTCPTransport(":8800")
	data := "Some new data"

	go func() {
		for {
			var msg string
			server.Consume(encoder.GOBDecoder{}, &msg)
			if data != msg {
				t.Errorf("Expected %s, found %s", data, msg)
			}
		}
	}()

	if err := node.Send(port, data); err != nil {
		t.Errorf("Failed to send message to server: %s, %v", port, err)
	}
	if err := node.Send(port, data); err != nil {
		t.Errorf("Failed to send message to server: %s, %v", port, err)
	}

	time.Sleep(time.Second * 5)
}

func TestStopServer(t *testing.T) {
	port := ":8000"
	server, _ := getServer(port)
	server.Listen()
	err := server.Stop()
	if err != nil {
		t.Errorf("Failed to stop server: %s, %v", port, err)
	}
}

func TestNewTransport(t *testing.T) {
	port := ":8900"
	_, err := getServer(port)
	if err != nil {
		t.Errorf("Failed to create server at port: %s", port)
	}
}
