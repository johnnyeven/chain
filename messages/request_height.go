package messages

import (
	"github.com/golang/protobuf/proto"
	"git.profzone.net/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*RequestHeight)(nil)

func (msg *RequestHeight) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__REQUEST_HEIGHT
}

func (msg *RequestHeight) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *RequestHeight) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}