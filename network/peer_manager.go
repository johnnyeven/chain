package network

import (
	"git.profzone.net/profzone/terra/dht"
	"reflect"
)

type PeerManager struct {
	peers *dht.SyncedMap
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: dht.NewSyncedMap(),
	}
}

func (pm *PeerManager) Get(peerID []byte) (*Peer, bool) {
	val, ok := pm.peers.Get(string(peerID))
	return val.(*Peer), ok
}

func (pm *PeerManager) Has(peerID []byte) bool {
	return pm.peers.Has(string(peerID))
}

func (pm *PeerManager) Set(val *Peer) {
	pm.peers.Set(string(val.Guid), val)
}

func (pm *PeerManager) Delete(peerID []byte) {
	pm.peers.Delete(string(peerID))
}

func (pm *PeerManager) DeleteMulti(peerIDs [][]byte) {
	interfaces := make([]interface{}, 0)
	for _, peerID := range peerIDs {
		v := reflect.ValueOf(string(peerID))
		interfaces = append(interfaces, v.Interface())
	}
	pm.peers.DeleteMulti(interfaces)
}

func (pm *PeerManager) Clear() {
	pm.peers.Clear()
}

func (pm *PeerManager) Iterator(iterator func(peer *Peer) error, errorContinue bool) {
	ch := pm.peers.Iter()
	for item := range ch {
		p := item.Value.(*Peer)
		err := iterator(p)
		if err != nil {
			if errorContinue {
				continue
			} else {
				break
			}
		}
	}
}

func (pm *PeerManager) Len() int {
	return pm.peers.Len()
}
