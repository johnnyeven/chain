package messages

import (
	"github.com/golang/protobuf/proto"
	"git.profzone.net/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*Heartbeat)(nil)

func (msg *Heartbeat) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__HEARTBEAT
}

func (msg *Heartbeat) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *Heartbeat) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}
