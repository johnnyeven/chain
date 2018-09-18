package messages

import (
	"github.com/golang/protobuf/proto"
	"git.profzone.net/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*Hello)(nil)

func (msg *Hello) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__HELLO
}

func (msg *Hello) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *Hello) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}
