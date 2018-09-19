package network

import (
	"git.profzone.net/profzone/terra/dht"
	"git.profzone.net/profzone/chain/messages"
	"git.profzone.net/profzone/chain/global"
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
