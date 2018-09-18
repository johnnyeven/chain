package dht

import (
	"net"
	"errors"
	"github.com/sirupsen/logrus"
	"time"
	"git.profzone.net/profzone/terra/dht/util"
)

const (
	PingType         = "ping"
	FindNodeType     = "find_node"
	GetPeersType     = "get_peers"
	AnnouncePeerType = "announce_peer"
)

const (
	GeneralError = 201 + iota
	ServerError
	ProtocolError
	UnknownError
)

var _ interface {
	TransportDriver
} = (*KRPCClient)(nil)

type KRPCClient struct {
	conn *net.UDPConn
	dht  *DistributedHashTable
}

func NewKRPCTransport(dht *DistributedHashTable, conn net.Conn, maxCursor uint64) *Transport {
	trans := &Transport{}
	trans.Init(dht, &KRPCClient{
		dht:  dht,
		conn: conn.(*net.UDPConn),
	}, maxCursor)

	return trans
}

func (c *KRPCClient) MakeRequest(id interface{}, remoteAddr net.Addr, requestType string, data interface{}) *Request {
	params := MakeQuery(c.dht.transport.GenerateTranID(), requestType, data.(map[string]interface{}))
	return &Request{
		ClientID:   id,
		CMD:        requestType,
		Data:       params,
		RemoteAddr: remoteAddr,
	}
}

func (c *KRPCClient) MakeResponse(id interface{}, remoteAddr net.Addr, tranID interface{}, data interface{}) *Request {
	params := MakeResponse(tranID.(string), data.(map[string]interface{}))
	return &Request{
		Data:       params,
		RemoteAddr: remoteAddr,
	}
}

func (c *KRPCClient) MakeError(id interface{}, remoteAddr net.Addr, tranID interface{}, errCode int, errMsg string) *Request {
	params := MakeError(tranID.(string), errCode, errMsg)
	return &Request{
		Data:       params,
		RemoteAddr: remoteAddr,
	}
}

func (c *KRPCClient) Request(request *Request) {}

func (c *KRPCClient) SendRequest(request *Request, retry int) {
	tranID := request.Data.(map[string]interface{})["t"].(string)
	tran := c.dht.transport.NewTransaction(tranID, request, retry)
	c.dht.transport.InsertTransaction(tran)
	defer c.dht.transport.DeleteTransaction(tran.ID)

	success := false
Run:
	for i := 0; i < retry; i++ {
		logrus.Debugf("[KRPCClient].Request c.conn.WriteToUDP try %d", i+1)
		err := c.Send(request)
		if err != nil {
			logrus.Warningf("[KRPCClient].Request c.conn.WriteToUDP err: %v", err)
			break
		}

		select {
		case <-tran.ResponseChannel:
			success = true
			break Run
		case <-time.After(time.Second * 15):
		}
	}

	if !success {
		c.dht.GetRoutingTable().RemoveByAddr(request.RemoteAddr.String())
	}
}

func (c *KRPCClient) Send(request *Request) error {
	count, err := c.conn.WriteToUDP([]byte(util.Encode(request.Data)), request.RemoteAddr.(*net.UDPAddr))
	if err != nil {
		return err
	}

	logrus.Debugf("[KRPCClient].Sent %d bytes", count)

	return nil
}

func (c *KRPCClient) Receive(receiveChannel chan Packet) {
	buff := make([]byte, 8192)
	for {
		n, raddr, err := c.conn.ReadFromUDP(buff)
		if err != nil {
			continue
		}

		receiveChannel <- Packet{buff[:n], raddr}
	}
}

func (c *KRPCClient) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *KRPCClient) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *KRPCClient) Close() error {
	return c.conn.Close()
}

func (c *KRPCClient) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *KRPCClient) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *KRPCClient) SetDeadline(time time.Time) error {
	return c.conn.SetDeadline(time)
}

func (c *KRPCClient) SetReadDeadline(time time.Time) error {
	return c.conn.SetReadDeadline(time)
}

func (c *KRPCClient) SetWriteDeadline(time time.Time) error {
	return c.conn.SetWriteDeadline(time)
}

func ParseMessage(data interface{}) (map[string]interface{}, error) {
	response, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.New("response is not dict")
	}

	if err := ParseKeys(
		response, [][]string{{"t", "string"}, {"y", "string"}}); err != nil {
		return nil, err
	}

	return response, nil
}

func ParseKeys(data map[string]interface{}, pairs [][]string) error {
	for _, args := range pairs {
		key, t := args[0], args[1]
		if err := ParseKey(data, key, t); err != nil {
			return err
		}
	}
	return nil
}

func ParseKey(data map[string]interface{}, key string, t string) error {
	val, ok := data[key]
	if !ok {
		return errors.New("lack of key")
	}

	switch t {
	case "string":
		_, ok = val.(string)
	case "int":
		_, ok = val.(int)
	case "map":
		_, ok = val.(map[string]interface{})
	case "list":
		_, ok = val.([]interface{})
	default:
		panic("invalid type")
	}

	if !ok {
		return errors.New("invalid key type")
	}

	return nil
}

// makeQuery returns a query-formed data.
func MakeQuery(t, q string, a map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"t": t,
		"y": "q",
		"q": q,
		"a": a,
	}
}

// makeResponse returns a response-formed data.
func MakeResponse(t string, r map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"t": t,
		"y": "r",
		"r": r,
	}
}

// makeError returns a err-formed data.
func MakeError(t string, errCode int, errMsg string) map[string]interface{} {
	return map[string]interface{}{
		"t": t,
		"y": "e",
		"e": []interface{}{errCode, errMsg},
	}
}
