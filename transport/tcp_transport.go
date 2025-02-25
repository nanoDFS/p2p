package transport

import (
	"fmt"
	"log"
	"net"

	"github.com/nanoDFS/p2p/encoder"
)

type Message struct {
	Payload []byte
}

type TCPTransport struct {
	Addr    net.Addr
	Queue   chan Message
	Encoder encoder.Encoder
}

func NewTCPTransport(addr string) (*TCPTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address type, %v", err)
	}
	return &TCPTransport{
		Addr:    address,
		Queue:   make(chan Message),
		Encoder: encoder.GOBEncoder{},
	}, nil
}

func (t *TCPTransport) Start() error {
	listener, err := net.Listen(t.Addr.Network(), t.Addr.String())
	if err != nil {
		return fmt.Errorf("failed to start server, %v", err)
	}
	log.Printf("Started listening at port %s", t.Addr)

	go t.connectionLoop(listener)
	return nil
}

func (t *TCPTransport) connectionLoop(listener net.Listener) {
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to establish connection with %s", t.Addr.String())
		}
		go t.handleConnection(conn)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn) error {
	defer func() {
		log.Printf("Dropping connection: %s\n", conn.RemoteAddr())
		conn.Close()
	}()

	var buffer = make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to read from %s", conn.RemoteAddr())
		}
		fmt.Printf("Recieved message of length %d from %s\n", n, conn.RemoteAddr().String())
		t.Queue <- Message{Payload: buffer[:n]}
		log.Printf("Recieved data form %s", conn.RemoteAddr())
	}
}
