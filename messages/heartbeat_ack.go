package messages

import (
	"github.com/golang/protobuf/proto"
	"git.profzone.net/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*HeartbeatAck)(nil)

func (msg *HeartbeatAck) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__HEARTBEAT_ACK
}

func (msg *HeartbeatAck) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *HeartbeatAck) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}