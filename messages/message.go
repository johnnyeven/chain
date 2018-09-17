package messages

import (
	"git.profzone.net/profzone/chain/global"
	"git.profzone.net/profzone/terra/dht"
	"github.com/sirupsen/logrus"
)

type Message struct {
	Header    global.MessageType
	MessageID int64
	PeerID    int64
	Payload   []byte
}

type MessageSerializable interface {
	GetMessageType() global.MessageType
	DecodeFromSource([]byte) error
	EncodeFromSource() ([]byte, error)
}

type MessageRunner func(t *dht.Transport, msg *Message) error

type MessageHandler struct {
	Type   global.MessageType
	Runner MessageRunner
}

type MessageManager struct {
	messageMap *dht.SyncedMap
}

var messageManager *MessageManager

func GetMessageManager() *MessageManager {
	if messageManager == nil {
		messageManager = &MessageManager{
			messageMap: dht.NewSyncedMap(),
		}
	}

	return messageManager
}

func (m *MessageManager) RegisterMessage(handler MessageHandler) {
	if m.messageMap.Has(handler.Type) {
		logrus.Panicf("[MessageManager] %s already registered", handler.Type.String())
	}

	m.messageMap.Set(handler.Type, handler.Runner)
}

func (m *MessageManager) GetMessageRunner(messageType global.MessageType) MessageRunner {
	v, _ := m.messageMap.Get(messageType)
	return v.(MessageRunner)
}
