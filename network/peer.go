package network

import (
	"git.profzone.net/profzone/terra/dht"
	"net"
	"git.profzone.net/profzone/chain/messages"
)

type Peer struct {
	Guid          []byte
	Node          *dht.Node
	transport     *dht.Transport
	packetChannel chan dht.Packet
	quitChannel   chan struct{}
	Handler       func(*Peer, dht.Packet)
}

func NewPeerWithConnection(id []byte, node *dht.Node, conn *net.TCPConn) *Peer {
	t := NewChainProtobufTransport(conn, 5000)
	p := &Peer{
		Guid:          id,
		Node:          node,
		transport:     t,
		packetChannel: make(chan dht.Packet),
		quitChannel:   make(chan struct{}),
		Handler:       PeerPacketHandler,
	}
	t.GetClient().(*ChainProtobufClient).peer = p

	go p.Run()

	return p
}

func (p *Peer) listen() {
	go p.transport.Receive(p.packetChannel)
}

func (p *Peer) Run() {
	go p.transport.Run()
	p.listen()

Run:
	for {
		select {
		case packet := <-p.packetChannel:
			p.Handler(p, packet)
		case <-p.quitChannel:
			break Run
		}
	}
}

func (p *Peer) GetTransport() *dht.Transport {
	return p.transport
}

func (p *Peer) Handshake() {
	message := &messages.HelloTCP{
		Guid: p.Guid,
		Node: []byte(p.Node.CompactNodeInfo()),
	}
	request := p.transport.MakeRequest(p.Guid, p.Node.Addr, "", message)
	p.transport.Request(request)
}
