package network

import (
	"git.profzone.net/profzone/terra/dht"
	"github.com/sirupsen/logrus"
	"git.profzone.net/profzone/chain/messages"
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
