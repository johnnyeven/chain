package network

import (
	"git.profzone.net/profzone/terra/dht"
	"github.com/sirupsen/logrus"
	"git.profzone.net/profzone/chain/messages"
	"git.profzone.net/profzone/chain/global"
)

func Ping(node *dht.Node, t *dht.Transport) {

}

func Handshake(node *dht.Node, t *dht.Transport, target string) {
	logrus.Debugf("[Client] SayHelloAction to RemoteAddr: %v", node.Addr)

	message := &messages.Hello{
		Guid:    global.Config.Guid,
		Version: global.Config.Version,
		Ip:      nil,
		Port:    0,
	}

	request := t.MakeRequest(global.Config.Guid, node.Addr, "", message)
	t.Request(request)
}
