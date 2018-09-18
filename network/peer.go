package network

import (
	"git.profzone.net/profzone/terra/dht"
	"net"
)

type Peer struct {
	Guid        []byte
	transport   *dht.Transport
	quitChannel chan struct{}
}

func NewPeerWithConnection(id []byte, conn *net.TCPConn) *Peer {
	t := NewChainProtobufTransport(nil, conn, 5000)
	p := &Peer{
		Guid:        id,
		transport:   t,
		quitChannel: make(chan struct{}),
	}
	return p
}
