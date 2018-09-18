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

	//pm := GetPeerManager()
	//peer := pm.Get(msg.PeerID)
	//if peer != nil && peer.TCPClient == nil {
	//	if conn, ok := conn.(*net.TCPConn); ok {
	//		peer.TCPClient = conn
	//	}
	//}
	logrus.Debug("[DHTPacketHandler] Handle message [MsgHeader=", msg.Header.String(), ", MsgID=", msg.MessageID, "] started")

	err := runner(table.GetTransport(), msg)
	if err != nil {
		logrus.Errorf("[DHTPacketHandler] Handle message err: %v", err)
	}

	logrus.Debug("[DHTPacketHandler] Handle message [MsgHeader=", msg.Header.String(), ", MsgID=", msg.MessageID, "] ended")
}

func NewNodeHandler(node *dht.Node) {
	logrus.Infof("new node recorded, id: %x, ip: %s, port: %d", []byte(node.ID.RawString()), node.Addr.IP.String(), node.Addr.Port)

	remote := net.JoinHostPort(node.Addr.IP.String(), strconv.FormatUint(uint64(node.Addr.Port), 10))
	conn, err := net.DialTimeout("tcp", remote, 10*time.Second)
	if err != nil {
		logrus.Errorf("[NewNodeHandler] net.DialTimeout error: %v", err)
	}

	p := NewPeerWithConnection(node.ID.RawData(), conn.(*net.TCPConn))
	//TODO p.Run() to start receiving request
	P2P.peerManager.Set(p)
}
