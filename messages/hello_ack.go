package messages

import (
	"github.com/golang/protobuf/proto"
	"git.profzone.net/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*HelloAck)(nil)

func (msg *HelloAck) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__HELLO_ACK
}

func (msg *HelloAck) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *HelloAck) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}
