package network

import (
	"net"
	"git.profzone.net/profzone/terra/dht"
	"time"
	"git.profzone.net/profzone/chain/messages"
	"github.com/sirupsen/logrus"
	"reflect"
	"git.profzone.net/profzone/chain/global"
	"io"
	"github.com/marpie/goguid"
)

var _ interface {
	dht.TransportDriver
} = (*ChainProtobufClient)(nil)

type ChainProtobufClient struct {
	conn  *net.TCPConn
	table *dht.DistributedHashTable
}

func NewChainProtobufTransport(table *dht.DistributedHashTable, conn net.Conn, maxCursor uint64) *dht.Transport {
	transport := &dht.Transport{}
	transport.Init(table, &ChainProtobufClient{
		table: table,
		conn:  conn.(*net.TCPConn),
	}, maxCursor)

	return transport
}

func (c *ChainProtobufClient) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *ChainProtobufClient) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *ChainProtobufClient) Close() error {
	return c.conn.Close()
}

func (c *ChainProtobufClient) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *ChainProtobufClient) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *ChainProtobufClient) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *ChainProtobufClient) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *ChainProtobufClient) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *ChainProtobufClient) MakeRequest(id interface{}, remoteAddr net.Addr, requestType string, data interface{}) *dht.Request {
	msg, ok := data.(messages.MessageSerializable)
	if !ok {
		logrus.Panicf("%s is not implement messages.MessageSerializable", reflect.TypeOf(data).String())
	}

	payload, err := msg.EncodeFromSource()
	if err != nil {
		logrus.Panicf("msg.EncodeFromSource error: %v", err)
	}
	message := &messages.Message{
		Header:    msg.GetMessageType(),
		PeerID:    global.Config.Guid,
		MessageID: guid.GetGUID(),
		Payload:   payload,
	}

	return &dht.Request{
		RemoteAddr: remoteAddr,
		CMD:        msg.GetMessageType().String(),
		ClientID:   id,
		Data:       message,
	}
}

func (c *ChainProtobufClient) MakeResponse(id interface{}, remoteAddr net.Addr, tranID interface{}, data interface{}) *dht.Request {
	msg, ok := data.(messages.MessageSerializable)
	if !ok {
		logrus.Panicf("%s is not implement messages.MessageSerializable", reflect.TypeOf(data).String())
	}

	payload, err := msg.EncodeFromSource()
	if err != nil {
		logrus.Panicf("msg.EncodeFromSource error: %v", err)
	}
	message := &messages.Message{
		Header:    msg.GetMessageType(),
		PeerID:    global.Config.Guid,
		MessageID: tranID.(int64),
		Payload:   payload,
	}

	return &dht.Request{
		RemoteAddr: remoteAddr,
		CMD:        msg.GetMessageType().String(),
		ClientID:   id,
		Data:       message,
	}
}

func (c *ChainProtobufClient) MakeError(id interface{}, remoteAddr net.Addr, tranID interface{}, errCode int, errMsg string) *dht.Request {
	panic("not implements")
}

func (c *ChainProtobufClient) Request(request *dht.Request) {}

func (c *ChainProtobufClient) SendRequest(request *dht.Request, retry int) {
	tranID := request.Data.(*messages.Message).MessageID
	tran := c.table.GetTransport().NewTransaction(tranID, request, retry)
	c.table.GetTransport().InsertTransaction(tran)
	defer c.table.GetTransport().DeleteTransaction(tran.ID)

	success := false
Run:
	for i := 0; i < retry; i++ {
		logrus.Debugf("[ProtobufClient].Request c.conn.WriteToUDP try %d", i+1)
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
		c.table.GetRoutingTable().RemoveByAddr(request.RemoteAddr.String())
	}
}

func (c *ChainProtobufClient) Send(request *dht.Request) error {
	count, err := c.conn.Write(packMessage(request.Data.(*messages.Message)))
	if err != nil {
		return err
	}

	logrus.Debugf("[ChainProtobufClient].Sent %d bytes", count)

	return nil
}

func (c *ChainProtobufClient) Receive(receiveChannel chan dht.Packet) {
	readedUDPMessage := make(chan *messages.Message)
	go c.handleDeserializeData(readedUDPMessage, receiveChannel, c.conn)
	for {
		buffer := make([]byte, 8192)
		count, err := c.conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				logrus.Error(err.Error())
			}
			continue
		}
		if count == 0 {
			continue
		}
		unpackMessage(buffer[:count], readedUDPMessage, c.conn.RemoteAddr())
	}
}

func (c *ChainProtobufClient) handleDeserializeData(readedMessage chan *messages.Message, receiveChannel chan dht.Packet, conn net.Conn) {
	for {
		msg, isOpened := <-readedMessage
		if !isOpened || msg == nil {
			break
		}

		receiveChannel <- dht.Packet{packMessage(msg), msg.RemoteAddr}
	}
}
