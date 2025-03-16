package transport

import (
	"log"
	"testing"
	"time"

	"github.com/nanoDFS/p2p/p2p/encoder"
	"github.com/nanoDFS/p2p/p2p/transport/tcp"
)

func getServer(addr string) (*tcp.TCPTransport, error) {
	return tcp.NewTCPTransport(addr)
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
	node, _ := tcp.NewTCPTransport(":8800")
	data := "TCP is really amazing"
	go func() {
		for {
			var msg string
			addr, _ := server.Consume(encoder.GOBDecoder{}, &msg)
			log.Println(msg, addr)
			if data != msg {
				t.Errorf("Expected %s, found %s", data, msg)
			}
		}
	}()

	if err := node.Send(port, data); err != nil {
		t.Errorf("Failed to send message to server: %s, %v", port, err)
	}

	server.Stop()

	if err := node.Send(port, data); err != nil {
		t.Errorf("Failed to send message to server: %s, %v", port, err)
	}

	time.Sleep(time.Second * 2)
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
