package transport

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/nanoDFS/p2p/encoder"
)

func TestNewTransport(t *testing.T) {
	port := ":9000"
	server, err := NewTCPTransport(port)
	if err != nil {
		t.Errorf("Failed to create server at port: %s", port)
	}
	if err := server.Listen(); err != nil {
		t.Errorf("Failed to start server at port: %s", port)
	}
	node, _ := NewTCPTransport(":8800")
	data := "Hi sample data"

	go func() {
		rec := <-server.IncommingMsgQueue
		var msg string
		encoder.GOBDecoder{}.Decode(bytes.NewBuffer(rec.Payload), &msg)
		log.Println(msg)
		if data != msg {
			t.Errorf("Expected %s, found %s", data, msg)
		}
		//server.Stop()
	}()

	if err := node.Send(port, data); err != nil {
		t.Errorf("Failed to send message to server: %s, %v", port, err)
	}

	// err = server.Stop()
	// if err != nil {
	// 	t.Errorf("Failed to stop server: %s, %v", port, err)
	// }

	time.Sleep(time.Microsecond * 5)
}
