package p2p

import (
	"fmt"
	"net"
	"sync"
)

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}
func (t *TCPTransportOpts) Close() error{
	
}
type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

type TCPTransport struct {
	TCPTransportOpts
	listenAddress string
	listener      net.Listener
	rpcch         chan RPC
	decoder       Decoder
	mu            sync.RWMutex
	peers         map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}
	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)
	if err := t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("TCP handshake error: %s\n", err)
		return
	}
	rpc := &RPC{}
	for {
		if err := t.Decoder.Decode(conn, rpc); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}
