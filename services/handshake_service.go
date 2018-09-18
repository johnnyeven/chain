package services

import (
	"github.com/sirupsen/logrus"
	"git.profzone.net/profzone/chain/messages"
	"git.profzone.net/profzone/chain/global"
	"git.profzone.net/profzone/terra/dht"
)

var _ interface {
	Service
} = (*HandshakeService)(nil)

type HandshakeService struct{}

func NewHandshakeService() Service {
	return &HandshakeService{}
}

func (s *HandshakeService) Messages() []messages.MessageHandler {
	return []messages.MessageHandler{
		{
			Type: global.MESSAGE_TYPE__HELLO,
			Runner:  s.RunHello,
		},
		{
			Type: global.MESSAGE_TYPE__HELLO_ACK,
			Runner:  s.RunHelloAck,
		},
		{
			Type: global.MESSAGE_TYPE__HELLO_TCP,
			Runner:  s.RunHelloTCP,
		},
	}
}

func (s *HandshakeService) Start() error {
	return nil
}

func (s *HandshakeService) Stop() error {
	return nil
}

func (s *HandshakeService) RunHello(t *dht.Transport, msg *messages.Message) error {
	payload := &messages.Hello{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	//pm := p2p.GetPeerManager()
	//if payload.Guid == pm.Guid {
	//	logrus.Warn("Self connected. ignored.")
	//	return nil
	//}
	//
	//message := messages.HelloAck{
	//	Version: pm.Version,
	//	Guid:    pm.Guid,
	//	Ip:      pm.AnnouncedAddr.IP,
	//	Port:    uint32(pm.AnnouncedAddr.Port),
	//}
	//resultBytes, err := message.EncodeFromSource()
	//if err != nil {
	//	return err
	//}
	//
	//response := constants.Message{
	//	Header:    constants.MESSAGE_TYPE__HELLO_ACK,
	//	MessageID: guid.GetGUID(),
	//	PeerID:    pm.Guid,
	//	Message:   resultBytes,
	//}
	//
	//count, err := conn.(*net.UDPConn).WriteToUDP(p2p.Pack(response), msg.UDPAddr)
	//
	//if err != nil {
	//	logrus.Fatal(err.Error())
	//}
	//
	//if count <= 0 {
	//	logrus.Fatal("Send [hello ack] failed")
	//}
	//
	//peer := p2p.NewPeer(payload.Guid, payload.Ip, payload.Port, payload.Version)
	//peer.UDPClient = conn.(*net.UDPConn)
	//peer.UDPRemoteAddr = msg.UDPAddr
	//
	//BroadcastNewPeerToPeersAction(peer)
	//NewPeersToPeerAction(peer)
	//
	//pm.Put(peer)

	return nil
}

func (s *HandshakeService) RunHelloAck(t *dht.Transport, msg *messages.Message) error {
	payload := &messages.HelloAck{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}
	logrus.Debug("[", payload.Guid, "(V:", payload.Version, ")] Hello Ack message")

	//peer := p2p.NewPeer(payload.Guid, payload.Ip, payload.Port, payload.Version)
	//peer.UDPRemoteAddr = conn.(*net.UDPConn).RemoteAddr().(*net.UDPAddr)
	//peer.UDPClient = conn.(*net.UDPConn)
	//
	//if err := peer.ConnectTCPServer(); err != nil {
	//	logrus.Errorf("HelloAck peer.ConnectTCPServer err: %v", err)
	//	return err
	//}
	//
	//p2p.GetPeerManager().Put(peer)
	//SayHelloTCP(peer)

	return nil
}

func (s *HandshakeService) RunHelloTCP(t *dht.Transport, msg *messages.Message) error {
	//peer := p2p.GetPeerManager().Get(msg.PeerID)
	//peer.TCPClient = conn.(*net.TCPConn)

	return nil
}
//
//func SayHelloAction(p *p2p.Peer) {
//	p.PrepareSendUDPMessage(func(peer *p2p.Peer) error {
//		logrus.Debugf("[Client] SayHelloAction to RemoteAddr: %v", p.UDPClient.RemoteAddr().(*net.UDPAddr).IP)
//		pm := p2p.GetPeerManager()
//		message := &messages.Hello{
//			Guid:    pm.Guid,
//			Version: pm.Version,
//			Ip:      pm.AnnouncedAddr.IP,
//			Port:    uint32(pm.AnnouncedAddr.Port),
//		}
//		messages.SendMessage(peer.UDPClient, pm.Guid, message)
//		return nil
//	})
//}

//func SayHelloTCP(p *p2p.Peer) {
//	pm := p2p.GetPeerManager()
//
//	message := &messages.HelloTCP{}
//	messages.SendMessage(p.TCPClient, pm.Guid, message)
//}
