package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/nanoDFS/p2p/encoder"
	"github.com/nanoDFS/p2p/transport"
)

type Greet struct {
	Msg string
}

func main() {
	fmt.Println("Hello from P2P ")
	server, err := transport.NewTCPTransport(":9000")
	if err != nil {
		fmt.Println("Got error")
	}
	server.Listen()

	time.Sleep(time.Second * 4)

	go func() {
		rec := <-server.IncommingMsgQueue
		var g Greet
		encoder.GOBDecoder{}.Decode(bytes.NewBuffer(rec.Payload), &g)
		fmt.Println(g)

	}()

	d := Greet{Msg: "Hi from client"}

	client, _ := transport.NewTCPTransport(":8990")

	client.Send(":9000", d)

	select {}
}
