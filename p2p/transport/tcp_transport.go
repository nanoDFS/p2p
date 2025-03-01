package transport

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/charmbracelet/log"

	"github.com/nanoDFS/p2p/p2p/encoder"
	"github.com/nanoDFS/p2p/p2p/peer"
)

type Message struct {
	Payload []byte
}

type TCPTransport struct {
	ListenAddr        net.Addr
	IncommingMsgQueue chan Message
	Encoder           encoder.Encoder
	OnAcceptingConn   func(conn net.Conn)

	mu       *sync.RWMutex
	PeersMap map[string]peer.Peer
	listener net.Listener
	wg       sync.WaitGroup
}

func NewTCPTransport(addr string) (*TCPTransport, error) {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address type, %v", err)
	}
	return &TCPTransport{

		ListenAddr:        address,
		IncommingMsgQueue: make(chan Message, 10),
		Encoder:           encoder.GOBEncoder{},
		PeersMap:          make(map[string]peer.Peer),
		wg:                sync.WaitGroup{},
		mu:                &sync.RWMutex{},
	}, nil
}

func (t *TCPTransport) Listen() error {
	var err error
	t.listener, err = net.Listen(t.ListenAddr.Network(), t.ListenAddr.String())
	if err != nil {
		return fmt.Errorf("failed to start server, %v", err)
	}
	log.Infof("TCP: Started listening at port %s", t.ListenAddr)
	go t.connectionLoop()
	return nil
}

func (t *TCPTransport) Stop() error {
	return t.listener.Close()
}

func (t *TCPTransport) Send(addr string, data any) error {
	return t.send(addr, data)
}

func (t *TCPTransport) Consume(decoder encoder.Decoder, writer any) error {
	data := <-t.IncommingMsgQueue
	return decoder.Decode(bytes.NewBuffer(data.Payload), writer)
}

func (t *TCPTransport) Close(addr string) error {
	return t.dropConnection(addr)
}

func (t *TCPTransport) send(addr string, data any) error {
	peerNode, err := t.dial(addr)
	if err != nil {
		return err
	}
	var msg bytes.Buffer
	_ = t.Encoder.Encode(data, &msg)

	length := uint32(len(msg.Bytes()))
	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)

	peerNode.GetConnection().Write(lengthBytes)
	n, err := peerNode.GetConnection().Write(msg.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send data to %s, %v", addr, err)
	}
	log.Debugf("successfully wrote %d bytes to %s", n, addr)
	return nil
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
	t.addConnection(conn)

	return t.addConnection(conn), nil
}

func (t *TCPTransport) connectionLoop() {
	defer func() {
		log.Infof("Shutting down server: %s", t.ListenAddr)
		t.wg.Done()
	}()

	t.wg.Add(1)
	for {

		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			log.Warnf("failed to establish connection with %s", t.ListenAddr.String())
		}

		t.addConnection(conn)
		if t.OnAcceptingConn != nil {
			t.OnAcceptingConn(conn)
		}
		go t.handleConnection(conn)

	}
}

func (t *TCPTransport) handleConnection(conn net.Conn) error {

	defer func() {
		t.dropConnection(conn.RemoteAddr().String())
		log.Infof("Dropping connection: %s\n", conn.RemoteAddr())
	}()

	for {
		// first 4 bytes are used for message length
		lengthBytes := make([]byte, 4)
		_, err := conn.Read(lengthBytes)
		length := binary.BigEndian.Uint32(lengthBytes)

		if err == io.EOF {
			return err
		}

		var buffer = make([]byte, length)
		n, err := conn.Read(buffer)

		if err != nil {
			return fmt.Errorf("failed to read from %s", conn.RemoteAddr())
		}

		log.Debugf("Recieved message of length %d from %s", n, conn.RemoteAddr().String())
		t.IncommingMsgQueue <- Message{Payload: buffer[:n]}

	}
}

func (t *TCPTransport) getConnection(addr string) (peer.Peer, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	tcp_addr := t.buildAddress(addr)
	if conn := t.PeersMap[tcp_addr]; conn != nil {
		return conn, nil
	}
	return nil, fmt.Errorf("failed to get connection")
}

func (t *TCPTransport) addConnection(conn net.Conn) peer.Peer {
	t.mu.Lock()
	t.PeersMap[conn.RemoteAddr().String()] = &peer.TCPPeer{Conn: conn}
	res := t.PeersMap[conn.RemoteAddr().String()]
	t.mu.Unlock()
	return res
}

func (t *TCPTransport) dropConnection(addr string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if conn, _ := t.getConnection(addr); conn != nil {
		delete(t.PeersMap, conn.GetAddress().String())
		return conn.Close()
	} else {
		return fmt.Errorf("no existing connections found")
	}
}

func (t *TCPTransport) buildAddress(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	} else {
		return addr
	}
}
