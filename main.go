package main

import (
	"github.com/charmbracelet/log"
	"github.com/nanoDFS/p2p/p2p/transport"
)

func main() {

	port := ":8081"
	server, _ := transport.NewTCPTransport(port)
	if err := server.Listen(); err != nil {
		log.Errorf("failed to listen, %s, %v", port, err)
	}

	client, _ := transport.NewTCPTransport(":8078")
	msg := "Hey this is a sample message"
	client.Send(port, msg)

	select {}
}
