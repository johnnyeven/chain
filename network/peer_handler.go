package network

import (
	"github.com/johnnyeven/terra/dht"
	"github.com/johnnyeven/chain/messages"
	"github.com/sirupsen/logrus"
)

func PeerPacketHandler(p *Peer, packet dht.Packet) {
	msg := unpackMessageFromPackage(packet.Data)
	runner, ok := messages.GetMessageManager().GetMessageRunner(msg.Header)

	if !ok {
		logrus.Errorf("[PeerPacketHandler] handleDeserializeData err MsgHeader: %d", msg.Header)
		return
	}

	logrus.Debug("[PeerPacketHandler] Handle message [MsgHeader=", msg.Header.String(), ", MsgID=", msg.MessageID, "] started")

	err := runner(p.transport, msg)
	if err != nil {
		logrus.Errorf("[PeerPacketHandler] Handle message err: %v", err)
	}

	tranID := msg.MessageID
	tran := p.transport.Get(tranID, msg.RemoteAddr)
	if tran != nil {
		defer func() {
			tran.ResponseChannel <- struct{}{}
		}()
	}

	logrus.Debug("[PeerPacketHandler] Handle message [MsgHeader=", msg.Header.String(), ", MsgID=", msg.MessageID, "] ended")
}
