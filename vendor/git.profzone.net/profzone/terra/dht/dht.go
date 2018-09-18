package dht

import (
	"net"
	"github.com/sirupsen/logrus"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"time"
	"git.profzone.net/profzone/terra/dht/util"
	"math"
)

type DistributedHashTable struct {
	// the kbucket expired duration
	BucketExpiredAfter time.Duration
	// the node expired duration
	NodeExpiredAfter time.Duration
	// how long it checks whether the bucket is expired
	CheckBucketPeriod time.Duration
	// the max transaction id
	MaxTransactionCursor uint64
	// how many nodes routing table can hold
	MaxNodes int
	// in mainline dht, k = 8
	K int
	// for crawling mode, we put all nodes in one bucket, so BucketSize may
	// not be K
	BucketSize int
	// the nodes num to be fresh in a kbucket
	RefreshNodeCount int
	// udp, udp4, udp6
	Network string
	// local network address
	LocalAddr string
	// initialized node list
	SeedNodes []string
	// the constructor func for transport
	TransportConstructor func(dht *DistributedHashTable, conn net.Conn, maxCursor uint64) *Transport
	// the Transport communicating component
	transport *Transport
	// node storage engin
	routingTable *routingTable
	// NAT
	nat nat.Interface
	// self node
	Self *Node
	// received packet channel
	packetChannel chan Packet
	// system shutdown channel
	quitChannel chan struct{}
	// packet handler
	Handler func(table *DistributedHashTable, packet Packet)
	// join method
	HandshakeFunc func(node *Node, t *Transport, target []byte)
	// ping method
	PingFunc func(node *Node, t *Transport)
}

type Config struct {
	BucketExpiredAfter   time.Duration
	NodeExpiredAfter     time.Duration
	CheckBucketPeriod    time.Duration
	MaxTransactionCursor uint64
	MaxNodes             int
	K                    int
	BucketSize           int
	RefreshNodeCount     int
	Network              string
	LocalAddr            string
	SeedNodes            []string
	TransportConstructor func(dht *DistributedHashTable, conn net.Conn, maxCursor uint64) *Transport
	Handler              func(table *DistributedHashTable, packet Packet)
	HandshakeFunc        func(node *Node, t *Transport, target []byte)
	PingFunc             func(node *Node, t *Transport)
}

func GetNormalConfig() *Config {
	return &Config{
		BucketExpiredAfter:   0,
		NodeExpiredAfter:     0,
		CheckBucketPeriod:    5 * time.Second,
		MaxTransactionCursor: math.MaxUint32,
		MaxNodes:             5000,
		K:                    8,
		BucketSize:           math.MaxInt32,
		RefreshNodeCount:     256,
		Network:              "udp4",
		LocalAddr:            ":6881",
	}
}

func NewDHT(config *Config) *DistributedHashTable {
	if config == nil {
		logrus.Panic("config is empty")
	}
	table := &DistributedHashTable{
		BucketExpiredAfter:   config.BucketExpiredAfter,
		NodeExpiredAfter:     config.NodeExpiredAfter,
		CheckBucketPeriod:    config.CheckBucketPeriod,
		MaxTransactionCursor: config.MaxTransactionCursor,
		MaxNodes:             config.MaxNodes,
		K:                    config.K,
		BucketSize:           config.BucketSize,
		RefreshNodeCount:     config.RefreshNodeCount,
		Network:              config.Network,
		LocalAddr:            config.LocalAddr,
		SeedNodes:            config.SeedNodes,
		TransportConstructor: config.TransportConstructor,
		Handler:              config.Handler,
		HandshakeFunc:        config.HandshakeFunc,
		PingFunc:             config.PingFunc,
	}

	return table
}

func (dht *DistributedHashTable) Run() {
	dht.init()
	dht.listen()
	dht.join()

	tick := time.Tick(dht.CheckBucketPeriod)

Run:
	for {
		select {
		case packet := <-dht.packetChannel:
			dht.Handler(dht, packet)
		case <-tick:
			if dht.routingTable.Len() == 0 {
				dht.join()
			} else if dht.transport.TransactionLength() == 0 {
				//go dht.routingTable.Fresh()
			}
		case <-dht.quitChannel:
			break Run
		}
	}
}

func (dht *DistributedHashTable) GetTransport() *Transport {
	return dht.transport
}

func (dht *DistributedHashTable) GetRoutingTable() *routingTable {
	return dht.routingTable
}

func (dht *DistributedHashTable) init() {
	if dht.TransportConstructor == nil {
		logrus.Panic("dht TransportConstructor not set")
	}
	if dht.HandshakeFunc == nil {
		logrus.Panic("dht HandshakeFunc not set")
	}
	if dht.PingFunc == nil {
		logrus.Panic("dht PingFunc not set")
	}
	if dht.Handler == nil {
		logrus.Panic("dht Handler not set")
	}
	listener, err := net.ListenPacket(dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init net.ListenPacket err: %v", err)
	}

	dht.transport = dht.TransportConstructor(dht, listener.(net.Conn), dht.MaxTransactionCursor)
	if dht.transport != nil {
		go dht.transport.Run()
	} else {
		logrus.Panic("dht.transport is nil")
	}

	dht.routingTable = newRoutingTable(dht.BucketSize, dht)
	dht.nat = nat.Any()
	dht.packetChannel = make(chan Packet)
	dht.quitChannel = make(chan struct{})

	dht.Self, err = NewNode(util.RandomString(20), dht.Network, dht.LocalAddr)
	if err != nil {
		logrus.Panicf("[DistributedHashTable].init NewNode err: %v", err)
	}
	logrus.Info("initialized")
}

func (dht *DistributedHashTable) join() {
	for _, addr := range dht.SeedNodes {
		udpAddr, err := net.ResolveUDPAddr(dht.Network, addr)
		if err != nil {
			logrus.Warningf("[DistributedHashTable].join net.ResolveUDPAddr err: %v", err)
			continue
		}

		dht.HandshakeFunc(&Node{Addr: udpAddr}, dht.transport, dht.Self.ID.RawData())
	}
}

func (dht *DistributedHashTable) listen() {
	realAddr := dht.transport.LocalAddr().(*net.UDPAddr)
	if dht.nat != nil {
		if !realAddr.IP.IsLoopback() {
			go nat.Map(dht.nat, dht.quitChannel, "udp4", realAddr.Port, realAddr.Port, "terra discovery")
		}
	}
	go dht.transport.Receive(dht.packetChannel)
	logrus.Info("listened")
}

func (dht *DistributedHashTable) ID(target string) string {
	if target == "" {
		return dht.Self.ID.RawString()
	}
	return target[:15] + dht.Self.ID.RawString()[15:]
}

func (dht *DistributedHashTable) Close() {
	dht.quitChannel <- struct{}{}
	dht.transport.Close()
	close(dht.packetChannel)
	close(dht.quitChannel)
}
