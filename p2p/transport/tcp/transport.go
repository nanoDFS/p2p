package tcp

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/charmbracelet/log"

	"github.com/nanoDFS/p2p/p2p/encoder"
	"github.com/nanoDFS/p2p/p2p/peer"
)

type Message struct {
	RemoteAddr string
	Payload    []byte
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

func (t *TCPTransport) Close(addr string) error {
	return t.dropConnection(addr)
}

func (t *TCPTransport) Consume(decoder encoder.Decoder, writer any) (string, error) {
	data := <-t.IncommingMsgQueue
	err := decoder.Decode(bytes.NewBuffer(data.Payload), writer)
	if err != nil {
		return "", err
	}
	return data.RemoteAddr, nil
}

func (t *TCPTransport) BroadCast(data any) error {
	for _, peer := range t.PeersMap {
		go t.send(peer.GetAddress().String(), data)
	}
	return nil
}
