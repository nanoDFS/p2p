package main

import (
	"fmt"

	"github.com/nanoDFS/p2p/transport"
)

func main() {
	fmt.Println("Hello from P2P ")
	server, err := transport.NewTransport(":9000")
	if err != nil {
		fmt.Println("Got error")
	}
	server.Start()
	select {}
}
