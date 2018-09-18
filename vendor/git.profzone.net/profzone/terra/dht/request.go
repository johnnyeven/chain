package dht

import (
	"net"
)

type Request struct {
	RemoteAddr net.Addr
	CMD        string
	ClientID   interface{}
	Data       interface{}
}
