package network

import (
	"net"
	"git.profzone.net/profzone/terra/dht"
	"git.profzone.net/profzone/chain/global"
)

var P2P *Server

type Server struct {
	transport     *dht.Transport
	AnnouncedAddr *net.UDPAddr
	dht           *dht.DistributedHashTable
	quitChannel   chan struct{}
}

func NewServer() *Server {
	if P2P != nil {
		return P2P
	}

	table := &dht.DistributedHashTable{
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
		Handler:              DHTPacketHandler,
		HandshakeFunc:        Handshake,
		PingFunc:             Ping,
	}

	P2P = &Server{
		dht:         table,
		quitChannel: make(chan struct{}),
	}

	return P2P
}

func (s *Server) init() {

}

func (s *Server) listen() {

}

func (s *Server) Run() {
	s.init()
	s.listen()

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
	s.transport.Close()
	close(s.quitChannel)
}
