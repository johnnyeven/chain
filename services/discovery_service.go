package services

import (
	"git.profzone.net/profzone/chain/messages"
	"git.profzone.net/profzone/chain/global"
	"git.profzone.net/profzone/terra/dht"
)

var _ interface {
	Service
} = (*DiscoveryService)(nil)

type DiscoveryService struct{}

func NewDiscoveryService() Service {
	return &DiscoveryService{}
}

func (s *DiscoveryService) Messages() []messages.MessageHandler {
	return []messages.MessageHandler{
		{
			Type:   global.MESSAGE_TYPE__NEW_PEERS_TO_PEER,
			Runner: s.RunNewPeersToPeer,
		},
		{
			Type:   global.MESSAGE_TYPE__BROADCAST_NEW_PEER_TO_PEERS,
			Runner: s.RunBroadcastNewPeerToPeers,
		},
	}
}

func (s *DiscoveryService) Start() error {
	return nil
}

func (s *DiscoveryService) Stop() error {
	return nil
}

func (s *DiscoveryService) RunNewPeersToPeer(t *dht.Transport, msg *messages.Message) error {
	payload := &messages.NewPeersToPeer{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	//peers := payload.Peers
	//pm := p2p.GetPeerManager()
	//if pm == nil {
	//	return errors.New("PeerManager not ready")
	//}
	//for _, peerInfo := range peers {
	//	if peerInfo.Guid == pm.Guid {
	//		continue
	//	}
	//	if pm.Get(peerInfo.Guid) != nil {
	//		continue
	//	}
	//	logrus.Debugf("new introduced peer: %v", peerInfo.Guid)
	//	peer := p2p.NewPeer(peerInfo.Guid, peerInfo.Ip, peerInfo.Port, peerInfo.Version)
	//	SayHelloAction(peer)
	//}

	return nil
}

func (s *DiscoveryService) RunBroadcastNewPeerToPeers(t *dht.Transport, msg *messages.Message) error {
	payload := &messages.BroadcastNewPeerToPeers{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}
	//
	//pm := p2p.GetPeerManager()
	//if pm == nil {
	//	return errors.New("PeerManager not ready")
	//}
	//if payload.Guid == pm.Guid {
	//	return nil
	//}
	//if pm.Get(payload.Guid) != nil {
	//	return nil
	//}
	//logrus.Debugf("new broadcasting peer: %v", payload.Guid)
	//peer := p2p.NewPeer(0, payload.Ip, payload.Port, 1)
	//SayHelloAction(peer)

	return nil
}

//func NewPeersToPeerAction(target *p2p.Peer) error {
//	pm := p2p.GetPeerManager()
//	if pm == nil {
//		return errors.New("PeerManager not exist")
//	}
//	return target.PrepareSendUDPMessage(func(peer *p2p.Peer) error {
//		message := &messages.NewPeersToPeer{
//			Peers: make([]*messages.NewPeersToPeer_Peer, 0),
//		}
//
//		pm.Iterator(func(guid int64, friend *p2p.Peer) error {
//			message.Peers = append(message.Peers, &messages.NewPeersToPeer_Peer{
//				Ip:      friend.IP,
//				Port:    uint32(friend.Port),
//				Guid:    friend.Guid,
//				Version: friend.Version,
//			})
//			return nil
//		}, false)
//
//		if len(message.Peers) > 0 {
//			err := messages.SendUDPMessage(peer.UDPClient, peer.UDPRemoteAddr, pm.Guid, message)
//			if err != nil {
//				logrus.Errorf("NewPeersToPeerAction module.SendMessage err: %v", err)
//				return err
//			}
//		}
//		return nil
//	})
//}
//
//func BroadcastNewPeerToPeersAction(newPeer *p2p.Peer) error {
//	pm := p2p.GetPeerManager()
//	if pm == nil {
//		return errors.New("PeerManager not exist")
//	}
//	pm.Iterator(func(guid int64, target *p2p.Peer) error {
//		return target.PrepareSendUDPMessage(func(p *p2p.Peer) error {
//			message := &messages.BroadcastNewPeerToPeers{
//				Ip:      newPeer.IP,
//				Port:    uint32(newPeer.Port),
//				Guid:    newPeer.Guid,
//				Version: newPeer.Version,
//			}
//			err := messages.SendUDPMessage(p.UDPClient, p.UDPRemoteAddr, pm.Guid, message)
//			if err != nil {
//				logrus.Errorf("BroadcastNewPeerToPeersAction module.SendMessage err: %v", err)
//				return err
//			}
//			return nil
//		})
//	}, false)
//
//	return nil
//}
