package services

import (
	"git.profzone.net/profzone/chain/messages"
	"git.profzone.net/profzone/chain/global"
	"git.profzone.net/profzone/terra/dht"
	"bytes"
	"errors"
	"strings"
	"git.profzone.net/profzone/chain/network"
	"github.com/sirupsen/logrus"
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
			Type:   global.MESSAGE_TYPE__HELLO_TCP,
			Runner: s.RunHelloTCP,
		},
		{
			Type:   global.MESSAGE_TYPE__FIND_NODE,
			Runner: s.RunFindNode,
		},
		{
			Type:   global.MESSAGE_TYPE__FIND_NODE_ACK,
			Runner: s.RunFindNodeAck,
		},
	}
}

func (s *DiscoveryService) Start() error {
	return nil
}

func (s *DiscoveryService) Stop() error {
	return nil
}

func (s *DiscoveryService) RunHelloTCP(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.HelloTCP{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	node, err := dht.NewNodeFromCompactInfo(string(payload.Node), "udp4")
	if err != nil {
		return err
	}

	p := t.GetClient().(*network.ChainProtobufClient).GetPeer()
	p.Node = node
	p.Guid = payload.Guid

	network.P2P.GetPeerManager().Set(p)

	logrus.Infof("peer handshake, peerID: %x", p.Guid)

	return nil
}

func (s *DiscoveryService) RunFindNode(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.FindNode{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	targetID := dht.NewIdentityFromBytes(payload.TargetGuid)

	var nodes string
	node, _ := t.GetDHT().GetRoutingTable().GetNodeBucketByID(targetID)
	if node != nil {
		nodes = node.CompactNodeInfo()
	} else {
		nodes = strings.Join(t.GetDHT().GetRoutingTable().GetNeighborCompactInfos(targetID, t.GetDHT().K), "")
	}

	message := &messages.FindNodeAck{
		Guid:    []byte(t.GetDHT().ID(targetID.RawString())),
		Version: global.Config.Version,
		//Ip:      network.P2P.AnnouncedAddr.IP,
		//Port:    uint32(network.P2P.AnnouncedAddr.Port),
		Nodes: []byte(nodes),
	}

	request := t.MakeResponse(nil, msg.RemoteAddr, msg.MessageID, message)
	t.GetClient().(*network.ProtobufClient).Send(request)

	if bytes.Compare(msg.PeerID, global.Config.Guid) == 0 {
		return nil
	}

	n, _ := dht.NewNode(string(payload.Guid), msg.RemoteAddr.Network(), msg.RemoteAddr.String())
	t.GetDHT().GetRoutingTable().Insert(n)

	return nil
}

func (s *DiscoveryService) RunFindNodeAck(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.FindNodeAck{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	tranID := msg.MessageID
	tran := t.GetDHT().GetTransport().Get(tranID, msg.RemoteAddr)
	if tran == nil {
		return errors.New("error trans")
	}

	defer func() {
		tran.ResponseChannel <- struct{}{}
	}()

	guid := payload.Guid

	//if tran.ClientID.(*dht.Identity) != nil && tran.ClientID.(*dht.Identity).RawString() != string(guid) {
	//	t.GetDHT().GetRoutingTable().RemoveByAddr(msg.RemoteAddr.String())
	//	return nil
	//}

	node, err := dht.NewNode(string(guid), msg.RemoteAddr.Network(), msg.RemoteAddr.String())
	if err != nil {
		return err
	}

	if err := findOrContinueRequestTarget(t.GetDHT(), payload.Guid, payload); err != nil {
		return err
	}

	if bytes.Compare(msg.PeerID, global.Config.Guid) == 0 {
		return nil
	}

	if t.GetDHT().GetRoutingTable().Insert(node) {
		if t.GetDHT().NewNodeHandler != nil {
			t.GetDHT().NewNodeHandler(msg.PeerID, node)
		}
	}

	return nil
}

func findOrContinueRequestTarget(table *dht.DistributedHashTable, targetID []byte, data *messages.FindNodeAck) error {

	nodes := string(data.Nodes)
	if len(nodes)%26 != 0 {
		return errors.New("the length of nodes should can be divided by 26")
	}

	hasNew, found := false, false
	for i := 0; i < len(nodes)/26; i++ {
		node, _ := dht.NewNodeFromCompactInfo(string(nodes[i*26:(i+1)*26]), table.Network)
		if bytes.Compare(node.ID.RawData(), targetID) == 0 {
			found = true
		}

		if table.GetRoutingTable().Insert(node) {
			hasNew = true
		}
	}
	if found || !hasNew {
		return nil
	}

	id := dht.NewIdentityFromBytes(targetID)
	for _, node := range table.GetRoutingTable().GetNeighbors(id, table.K) {
		network.Handshake(node, table.GetTransport(), targetID)
	}

	return nil
}
