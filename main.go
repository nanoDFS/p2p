package main

import (
	"github.com/charmbracelet/log"
	"github.com/nanoDFS/p2p/p2p/transport/tcp"
)

func main() {

	port := ":8081"
	server, _ := tcp.NewTCPTransport(port)
	if err := server.Listen(); err != nil {
		log.Errorf("failed to listen, %s, %v", port, err)
	}

	client, _ := tcp.NewTCPTransport(":8078")
	msg := "Hey this is a sample message"
	client.Send(port, msg)

	select {}
}
