package network

import (
	"git.profzone.net/profzone/terra/dht"
	"github.com/sirupsen/logrus"
	"git.profzone.net/profzone/chain/messages"
	"net"
	"strconv"
	"time"
)

func DHTPacketHandler(table *dht.DistributedHashTable, packet dht.Packet) {
	msg := unpackMessageFromPackage(packet.Data)
	runner, ok := messages.GetMessageManager().GetMessageRunner(msg.Header)

	if !ok {
		logrus.Errorf("[DHTPacketHandler] handleDeserializeData err MsgHeader: %d", msg.Header)
		return
	}

	logrus.Debug("[DHTPacketHandler] Handle message [MsgHeader=", msg.Header.String(), ", MsgID=", msg.MessageID, "] started")

	err := runner(table.GetTransport(), msg)
	if err != nil {
		logrus.Errorf("[DHTPacketHandler] Handle message err: %v", err)
	}

	logrus.Debug("[DHTPacketHandler] Handle message [MsgHeader=", msg.Header.String(), ", MsgID=", msg.MessageID, "] ended")
}

func NewNodeHandler(peerID []byte, node *dht.Node) {
	logrus.Infof("new node recorded, id: %x, ip: %s, port: %d", []byte(node.ID.RawString()), node.Addr.IP.String(), node.Addr.Port)

	//TODO PeerManager count limit
	if P2P.peerManager.Has(peerID) {
		return
	}

	remote := net.JoinHostPort(node.Addr.IP.String(), strconv.FormatUint(uint64(node.Addr.Port), 10))
	conn, err := net.DialTimeout("tcp", remote, 10*time.Second)
	if err != nil {
		logrus.Errorf("[NewNodeHandler] net.DialTimeout error: %v", err)
	}

	p := NewPeerWithConnection(peerID, node, conn.(*net.TCPConn))
	P2P.peerManager.Set(p)
	logrus.Infof("new peer connect id: %x, address: %s", p.Guid, conn.RemoteAddr().String())

	p.Handshake()
}
