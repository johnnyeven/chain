package messages

import (
	"github.com/golang/protobuf/proto"
	"github.com/johnnyeven/chain/global"
)

var _ interface {
	MessageSerializable
} = (*GetBlock)(nil)

func (msg *GetBlock) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__GET_BLOCK
}

func (msg *GetBlock) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *GetBlock) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}