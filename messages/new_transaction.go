package messages

import (
	"github.com/golang/protobuf/proto"
	"git.profzone.net/profzone/chain/global"
)

var _ interface {
	MessageSerializable
} = (*NewTransaction)(nil)

func (msg *NewTransaction) GetMessageType() global.MessageType {
	return global.MESSAGE_TYPE__NEW_TRANSACTION
}

func (msg *NewTransaction) EncodeFromSource() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *NewTransaction) DecodeFromSource(source []byte) error {
	return proto.Unmarshal(source, msg)
}
