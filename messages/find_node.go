package messages

import (
	"github.com/golang/protobuf/proto"
	"github.com/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*FindNode)(nil)

func (msg *FindNode) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__FIND_NODE
}

func (msg *FindNode) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *FindNode) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}