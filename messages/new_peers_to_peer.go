package messages

import (
	"github.com/golang/protobuf/proto"
	"github.com/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*NewPeersToPeer)(nil)

func (msg *NewPeersToPeer) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__NEW_PEERS_TO_PEER
}

func (msg *NewPeersToPeer) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *NewPeersToPeer) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}
