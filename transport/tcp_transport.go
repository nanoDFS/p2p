package transport

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/nanoDFS/p2p/encoder"
	"github.com/nanoDFS/p2p/peer"
)

type Message struct {
	Payload []byte
}

type TCPTransport struct {
	ListenAddr        net.Addr
	IncommingMsgQueue chan Message
	Encoder           encoder.Encoder
	OnAcceptingConn   func(conn net.Conn)
	PeersMap          map[net.Addr]peer.Peer

	quitChan chan struct{}
	wg       sync.WaitGroup
}

func NewTCPTransport(addr string) (*TCPTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address type, %v", err)
	}
	return &TCPTransport{

		ListenAddr:        address,
		IncommingMsgQueue: make(chan Message),
		Encoder:           encoder.GOBEncoder{},
		PeersMap:          make(map[net.Addr]peer.Peer),
		quitChan:          make(chan struct{}),
		wg:                sync.WaitGroup{},
	}, nil
}

func (t *TCPTransport) Listen() error {
	listener, err := net.Listen(t.ListenAddr.Network(), t.ListenAddr.String())
	if err != nil {
		return fmt.Errorf("failed to start server, %v", err)
	}
	log.Printf("Started listening at port %s", t.ListenAddr)

	go t.connectionLoop(listener)
	return nil
}

func (t *TCPTransport) Stop() error {
	t.quitChan <- struct{}{}
	return nil
}

func (t *TCPTransport) Send(addr string, data any) error {
	peerNode, err := t.dial(addr)
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	err = t.Encoder.Encode(data, &buff)
	if err != nil {
		return fmt.Errorf("failed to encode data: %s, %v", data, err)
	}
	n, err := peerNode.GetConnection().Write(buff.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send data to %s, %v", addr, err)
	}
	log.Printf("successfully wrote %d bytes to %s", n, addr)
	return nil
}

func (t *TCPTransport) Close(addr string) error {
	peerNode, _ := t.getConnection(addr)
	if peerNode != nil {
		return peerNode.Close()
	}
	return fmt.Errorf("failed to close connection")
}

func (t *TCPTransport) dial(addr string) (peer.Peer, error) {
	peerNode, _ := t.getConnection(addr)
	if peerNode != nil {
		return peerNode, nil
	}
	conn, err := net.Dial(t.ListenAddr.Network(), addr)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection with %s", addr)
	}

	return t.addConnection(conn), nil
}

func (t *TCPTransport) connectionLoop(listener net.Listener) {
	defer func() {
		log.Printf("Shutting down server: %s", t.ListenAddr)
		t.wg.Done()
		listener.Close()
	}()

	t.wg.Add(1)
	for {
		select {
		case <-t.quitChan:
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("failed to establish connection with %s", t.ListenAddr.String())
			}

			t.PeersMap[conn.RemoteAddr()] = &peer.TCPPeer{Conn: conn}
			if t.OnAcceptingConn != nil {
				t.OnAcceptingConn(conn)
			}
			go t.handleConnection(conn)
		}
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
		t.IncommingMsgQueue <- Message{Payload: buffer[:n]}
		log.Printf("Recieved data form %s", conn.RemoteAddr())

	}
}

func (t *TCPTransport) getConnection(addr string) (peer.Peer, error) {
	tcp_addr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	if conn := t.PeersMap[tcp_addr]; conn != nil {
		return conn, nil
	}
	return nil, fmt.Errorf("failed to get connection")
}

func (t *TCPTransport) addConnection(conn net.Conn) peer.Peer {
	t.PeersMap[conn.RemoteAddr()] = &peer.TCPPeer{Conn: conn}
	return t.PeersMap[conn.RemoteAddr()]
}
