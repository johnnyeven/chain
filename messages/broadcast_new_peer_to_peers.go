package messages

import (
	"github.com/golang/protobuf/proto"
	"github.com/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*BroadcastNewPeerToPeers)(nil)

func (msg *BroadcastNewPeerToPeers) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__BROADCAST_NEW_PEER_TO_PEERS
}

func (msg *BroadcastNewPeerToPeers) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *BroadcastNewPeerToPeers) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}