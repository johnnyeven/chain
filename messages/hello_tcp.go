package messages

import (
	"github.com/golang/protobuf/proto"
	"git.profzone.net/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*HelloTCP)(nil)

func (msg *HelloTCP) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__HELLO_TCP
}

func (msg *HelloTCP) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *HelloTCP) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}
