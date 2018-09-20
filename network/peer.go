package network

import (
	"github.com/johnnyeven/terra/dht"
	"net"
	"github.com/johnnyeven/chain/messages"
	"github.com/johnnyeven/chain/global"
	"time"
	"github.com/marpie/goguid"
	"github.com/sirupsen/logrus"
)

type Peer struct {
	Guid          []byte
	Node          *dht.Node
	Heartbeat     *Heartbeat
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
		Heartbeat:     NewHeartbeat(),
		transport:     t,
		packetChannel: make(chan dht.Packet),
		quitChannel:   make(chan struct{}),
		Handler:       PeerPacketHandler,
	}
	t.GetClient().(*ChainProtobufClient).peer = p

	go p.Run()

	return p
}

func (p *Peer) Run() {
	go p.transport.Run()
	go p.transport.Receive(p.packetChannel)

	tick := time.Tick(global.Config.CheckPeerPeriod)

Run:
	for {
		select {
		case packet := <-p.packetChannel:
			p.Handler(p, packet)
		case <-tick:
			p.Ping()
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
		Guid: global.Config.Guid,
		Node: []byte(p.Node.CompactNodeInfo()),
	}
	request := p.transport.MakeRequest(p.Guid, p.Node.Addr, "", message)
	p.transport.Request(request)
}

func (p *Peer) Ping() {
	sequence := guid.GetGUID()
	message := &messages.Heartbeat{
		Sequence: sequence,
	}
	t := p.transport
	request := t.MakeRequest(p.Guid, p.Node.Addr, "", message)
	t.Request(request)

	p.Heartbeat.NewMessage(sequence)
}

func (p *Peer) Close() {
	p.quitChannel <- struct{}{}
	p.transport.Close()
	close(p.packetChannel)
	close(p.quitChannel)

	if P2P.peerManager.Has(p.Guid) {
		P2P.peerManager.Delete(p.Guid)
	}
}

type Heartbeat struct {
	Delay       float64
	Health      uint8
	messageList map[int64]time.Time
}

func NewHeartbeat() *Heartbeat {
	return &Heartbeat{
		Delay:       0,
		Health:      0,
		messageList: make(map[int64]time.Time),
	}
}

func (h *Heartbeat) NewMessage(sequence int64) {
	h.messageList[sequence] = time.Now()
}

func (h *Heartbeat) ResponseMessage(sequence int64) {
	if startTime, ok := h.messageList[sequence]; ok {
		d := time.Since(startTime)
		h.Delay = d.Seconds() * 1000.0
		delete(h.messageList, sequence)
		logrus.Debugf("Heartbeat ack: %s, Milliseconds: %.2f", d.String(), d.Seconds()*1000.0)
	}
}
