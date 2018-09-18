package messages

import (
	"github.com/golang/protobuf/proto"
	"git.profzone.net/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*BlocksHash)(nil)

func (msg *BlocksHash) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__BLOCKS_HASH
}

func (msg *BlocksHash) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *BlocksHash) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}