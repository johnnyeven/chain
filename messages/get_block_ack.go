package messages

import (
	"github.com/golang/protobuf/proto"
	"github.com/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*GetBlockAck)(nil)

func (msg *GetBlockAck) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__GET_BLOCK_ACK
}

func (msg *GetBlockAck) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *GetBlockAck) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}