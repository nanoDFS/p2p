package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/nanoDFS/p2p/p2p/peer"
)

func (t *TCPTransport) handleConnection(conn net.Conn) error {

	defer func() {
		t.dropConnection(conn.RemoteAddr().String())
		log.Debugf("Dropping connection: %s\n", conn.RemoteAddr())
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
		t.IncommingMsgQueue <- Message{RemoteAddr: conn.RemoteAddr().String(), Payload: buffer[:n]}

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
	tcp_addr := t.buildAddress(addr)
	if conn := t.PeersMap[tcp_addr]; conn != nil {
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
