package main

import (
	"bytes"
	"fmt"
	"net"
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
	server.Start()

	time.Sleep(time.Second * 4)

	go func() {
		rec := <-server.Queue
		var g Greet
		encoder.GOBDecoder{}.Decode(bytes.NewBuffer(rec.Payload), &g)
		fmt.Println(g)

	}()

	conn, _ := net.Dial("tcp", ":9000")
	d := Greet{Msg: "Hi from client"}
	var buff bytes.Buffer
	enc := encoder.GOBEncoder{}

	enc.Encode(d, &buff)

	conn.Write(buff.Bytes())

	select {}
}
