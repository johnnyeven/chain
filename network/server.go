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
	}

	return P2P
}

func (s *Server) init() {

}

func (s *Server) listen() {

	//realaddr := conn.LocalAddr().(*net.UDPAddr)
	//if s.NAT != nil {
	//	if !realaddr.IP.IsLoopback() {
	//		go nat.Map(s.NAT, s.quit, "udp", realaddr.Port, realaddr.Port, "terra discovery")
	//	}
	//	// TODO: react to external IP changes over time.
	//	ext, err := s.NAT.ExternalIP()
	//	if err == nil {
	//		realaddr = &net.UDPAddr{IP: ext, Port: realaddr.Port}
	//		logrus.Debugf("nat device found. realaddr detected: %v", realaddr)
	//	} else {
	//		logrus.Debugf("nat device not found. err: %v", err)
	//	}
	//}
	//s.AnnouncedAddr = realaddr
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
