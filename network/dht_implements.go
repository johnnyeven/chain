package network

import (
	"github.com/johnnyeven/terra/dht"
	"github.com/johnnyeven/chain/messages"
	"github.com/johnnyeven/chain/global"
)

func Ping(node *dht.Node, t *dht.Transport) {

}

func Handshake(node *dht.Node, t *dht.Transport, target []byte) {
	message := &messages.FindNode{
		Guid:       []byte(t.GetDHT().ID(string(target))),
		Version:    global.Config.Version,
		//Ip:         nil,
		//Port:       0,
		TargetGuid: target,
	}

	request := t.MakeRequest(node.ID, node.Addr, "", message)
	t.Request(request)
}
