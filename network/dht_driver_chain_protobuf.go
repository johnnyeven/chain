package network

import (
	"net"
	"github.com/profzone/terra/dht"
	"time"
	"github.com/profzone/chain/messages"
	"github.com/sirupsen/logrus"
	"reflect"
	"github.com/profzone/chain/global"
	"io"
	"github.com/marpie/goguid"
)

const TransactionExpiredTime = 2 * time.Minute

var _ interface {
	dht.TransportDriver
} = (*ChainProtobufClient)(nil)

type ChainProtobufClient struct {
	conn      *net.TCPConn
	transport *dht.Transport
	peer      *Peer
}

func NewChainProtobufTransport(conn net.Conn, maxCursor uint64) *dht.Transport {
	transport := &dht.Transport{}
	transport.Init(nil, &ChainProtobufClient{
		conn:      conn.(*net.TCPConn),
		transport: transport,
	}, maxCursor)

	return transport
}

func (c *ChainProtobufClient) GetPeer() *Peer {
	return c.peer
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
	tran := c.transport.NewTransaction(tranID, request, retry)
	c.transport.InsertTransaction(tran)
	defer c.transport.DeleteTransaction(tran.ID)

	response := false

	err := c.Send(request)
	if err != nil {
		logrus.Warningf("[ChainProtobufClient].Request c.Send err: %v", err)
	}
Run:
	for {
		select {
		case <-tran.ResponseChannel:
			response = true
			break Run
		case <-time.After(TransactionExpiredTime):
			break Run
		}
	}

	if !response {
		logrus.Warningf("[ChainProtobufClient] response timeout, tranID: %d", tranID)
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
