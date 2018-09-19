package network

import (
	"net"
	"git.profzone.net/profzone/terra/dht"
	"git.profzone.net/profzone/chain/global"
	"github.com/sirupsen/logrus"
)

var P2P *Server

type Server struct {
	listener    net.Listener
	dht         *dht.DistributedHashTable
	quitChannel chan struct{}
	peerManager *PeerManager
}

func GetChainDHTConfig() *dht.Config {
	return &dht.Config{
		BucketExpiredAfter:   global.Config.BucketExpiredAfter,
		NodeExpiredAfter:     global.Config.NodeExpriedAfter,
		CheckBucketPeriod:    global.Config.CheckBucketPeriod,
		MaxTransactionCursor: global.Config.MaxTransactionCursor,
		MaxNodes:             global.Config.MaxNodes,
		K:                    global.Config.K,
		BucketSize:           global.Config.BucketSize,
		RefreshNodeCount:     global.Config.RefreshNodeCount,
		Network:              global.Config.Network,
		LocalAddr:            global.Config.LocalAddr.String(),
		SeedNodes:            global.Config.SeedNodes,
		TransportConstructor: NewProtobufTransport,
		NewNodeHandler:       NewNodeHandler,
		Handler:              DHTPacketHandler,
		HandshakeFunc:        Handshake,
		PingFunc:             Ping,
	}
}

func NewServer() *Server {
	if P2P != nil {
		return P2P
	}

	table := dht.NewDHT(GetChainDHTConfig())

	P2P = &Server{
		dht:         table,
		quitChannel: make(chan struct{}),
		peerManager: NewPeerManager(),
	}

	return P2P
}

func (s *Server) init() {
	var err error
	s.listener, err = net.Listen("tcp", s.dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[Server] net.Listen error: %v", err.Error())
	}
	logrus.Infof("[Server] created and listened at: %s ...", s.dht.LocalAddr)
}

func (s *Server) listen() {

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			logrus.Fatalf("[Server] listener.Accept error: %v", err)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	p := NewPeerWithConnection(nil, nil, conn.(*net.TCPConn))
	logrus.Infof("new peer connected id: %x, address: %s", p.Guid, conn.RemoteAddr().String())
}

func (s *Server) Run() {
	s.init()
	go s.listen()

	go s.dht.Run()

Run:
	for {
		select {
		case <-s.quitChannel:
			break Run
		}
	}
}

func (s *Server) Close() {
	s.quitChannel <- struct{}{}
	s.dht.Close()
	close(s.quitChannel)
}

func (s *Server) GetPeerManager() *PeerManager {
	return s.peerManager
}
