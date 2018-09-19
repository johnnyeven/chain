package messages

import (
	"github.com/profzone/chain/global"
	"github.com/profzone/terra/dht"
	"github.com/sirupsen/logrus"
	"net"
	"encoding/gob"
)

type Message struct {
	Header     global.MessageType
	PeerID     []byte
	MessageID  int64
	Payload    []byte
	RemoteAddr net.Addr
}

func init() {
	gob.Register(&net.UDPAddr{})
	gob.Register(&net.TCPAddr{})
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

func (m *MessageManager) GetMessageRunner(messageType global.MessageType) (MessageRunner, bool) {
	v, ok := m.messageMap.Get(messageType)
	return v.(MessageRunner), ok
}
