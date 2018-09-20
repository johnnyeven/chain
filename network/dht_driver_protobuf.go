package network

import (
	"net"
	"github.com/johnnyeven/terra/dht"
	"time"
	"github.com/johnnyeven/chain/messages"
	"github.com/sirupsen/logrus"
	"reflect"
	"github.com/johnnyeven/chain/global"
	"io"
	"github.com/marpie/goguid"
)

var _ interface {
	dht.TransportDriver
} = (*ProtobufClient)(nil)

type ProtobufClient struct {
	conn  *net.UDPConn
	table *dht.DistributedHashTable
}

func NewProtobufTransport(table *dht.DistributedHashTable, conn net.Conn, maxCursor uint64) *dht.Transport {
	transport := &dht.Transport{}
	transport.Init(table, &ProtobufClient{
		table: table,
		conn:  conn.(*net.UDPConn),
	}, maxCursor)

	return transport
}

func (c *ProtobufClient) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *ProtobufClient) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *ProtobufClient) Close() error {
	return c.conn.Close()
}

func (c *ProtobufClient) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *ProtobufClient) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *ProtobufClient) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *ProtobufClient) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *ProtobufClient) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *ProtobufClient) MakeRequest(id interface{}, remoteAddr net.Addr, requestType string, data interface{}) *dht.Request {
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

func (c *ProtobufClient) MakeResponse(id interface{}, remoteAddr net.Addr, tranID interface{}, data interface{}) *dht.Request {
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

func (c *ProtobufClient) MakeError(id interface{}, remoteAddr net.Addr, tranID interface{}, errCode int, errMsg string) *dht.Request {
	panic("not implements")
}

func (c *ProtobufClient) Request(request *dht.Request) {}

func (c *ProtobufClient) SendRequest(request *dht.Request, retry int) {
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

func (c *ProtobufClient) Send(request *dht.Request) error {
	count, err := c.conn.WriteToUDP(packMessage(request.Data.(*messages.Message)), request.RemoteAddr.(*net.UDPAddr))
	if err != nil {
		return err
	}

	logrus.Debugf("[ProtobufClient].Sent %d bytes", count)

	return nil
}

func (c *ProtobufClient) Receive(receiveChannel chan dht.Packet) {
	readedUDPMessage := make(chan *messages.Message)
	go c.handleDeserializeData(readedUDPMessage, receiveChannel, c.conn)

	truncatedData := make([]byte, 0)
	for {
		buffer := make([]byte, 8192)
		count, udpAddr, err := c.conn.ReadFromUDP(buffer)
		if err != nil {
			if err != io.EOF {
				logrus.Error(err.Error())
			}
			continue
		}
		if count == 0 {
			continue
		}
		truncatedData = unpackMessage(append(truncatedData, buffer[:count]...), readedUDPMessage, udpAddr)
	}
}

func (c *ProtobufClient) handleDeserializeData(readMessage <-chan *messages.Message, receiveChannel chan<- dht.Packet, conn net.Conn) {
	for {
		msg, isOpened := <-readMessage
		if !isOpened || msg == nil {
			break
		}

		receiveChannel <- dht.Packet{packMessage(msg), msg.RemoteAddr}
	}
}
