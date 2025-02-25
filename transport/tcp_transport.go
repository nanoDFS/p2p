package transport

import (
	"fmt"
	"log"
	"net"
)

type TCPTransport struct {
	Addr net.Addr
}

func NewTransport(addr string) (*TCPTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address type, %v", err)
	}
	return &TCPTransport{
		Addr: address,
	}, nil
}

func (t *TCPTransport) Start() error {
	fmt.Printf("Started listening at port %s", t.Addr)
	listener, err := net.Listen(t.Addr.Network(), t.Addr.String())
	if err != nil {
		return fmt.Errorf("failed to start server, %v", err)
	}

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
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to read from %s", conn.RemoteAddr())
		}
		fmt.Printf("Recieved %s form %s", string(buffer[:n]), conn.RemoteAddr())

		_, err = conn.Write([]byte("Message received " + string(buffer[:n])))
		if err != nil {
			return fmt.Errorf("faile to write to client: %v", err)
		}
	}
}
