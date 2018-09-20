package services

import (
	"github.com/johnnyeven/chain/messages"
	"github.com/johnnyeven/chain/global"
	"github.com/johnnyeven/terra/dht"
	"github.com/johnnyeven/chain/network"
)

var _ interface {
	Service
} = (*HeartbeatService)(nil)

type HeartbeatService struct{}

func NewHeartbeatService() Service {
	return &HeartbeatService{}
}

func (s *HeartbeatService) Messages() []messages.MessageHandler {
	return []messages.MessageHandler{
		{
			Type:   global.MESSAGE_TYPE__HEARTBEAT,
			Runner: s.RunHeartbeat,
		},
		{
			Type:   global.MESSAGE_TYPE__HEARTBEAT_ACK,
			Runner: s.RunHeartbeatAck,
		},
	}
}

func (s *HeartbeatService) Start() error {
	return nil
}

func (s *HeartbeatService) Stop() error {
	return nil
}

func (s *HeartbeatService) RunHeartbeat(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.Heartbeat{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	ack := &messages.HeartbeatAck{
		Sequence: payload.Sequence,
	}

	peer := t.GetClient().(*network.ChainProtobufClient).GetPeer()
	request := t.MakeResponse(peer.Guid, peer.Node.Addr, msg.MessageID, ack)
	t.Request(request)

	return nil
}

func (s *HeartbeatService) RunHeartbeatAck(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.HeartbeatAck{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	peer := t.GetClient().(*network.ChainProtobufClient).GetPeer()
	peer.Heartbeat.ResponseMessage(payload.Sequence)

	return nil
}
