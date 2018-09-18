package dht

import "net"

type Packet struct {
	Data       []byte
	RemoteAddr net.Addr
}

type TransportDriver interface {
	net.Conn
	MakeRequest(id interface{}, remoteAddr net.Addr, requestType string, data interface{}) *Request
	MakeResponse(id interface{}, remoteAddr net.Addr, tranID interface{}, data interface{}) *Request
	MakeError(id interface{}, remoteAddr net.Addr, tranID interface{}, errCode int, errMsg string) *Request
	Request(request *Request)
	SendRequest(request *Request, retry int)
	Receive(receiveChannel chan Packet)
}
