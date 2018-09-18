package network

import (
	"net"
	"git.profzone.net/profzone/terra/dht"
	"time"
	"git.profzone.net/profzone/chain/messages"
	"github.com/sirupsen/logrus"
	"reflect"
	"bytes"
	"encoding/gob"
	"encoding/binary"
	"git.profzone.net/profzone/chain/global"
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
		PeerID:    id.(int64),
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
		PeerID:    id.(int64),
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
			break
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
		unpackMessage(buffer[:count], readedUDPMessage, udpAddr)
	}
}

func (c *ProtobufClient) handleDeserializeData(readedMessage chan *messages.Message, receiveChannel chan dht.Packet, conn net.Conn) {
	for {
		msg, isOpened := <-readedMessage
		if !isOpened || msg == nil {
			break
		}

		receiveChannel <- dht.Packet{packMessage(msg), msg.RemoteAddr}
	}
}

func packMessage(msg *messages.Message) []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(msg)
	if err != nil {
		logrus.Fatal(err)
	}
	resultBytes := result.Bytes()
	dataLength := uint32(len(resultBytes))
	dataLengthBuffer := bytes.NewBuffer([]byte{})
	binary.Write(dataLengthBuffer, binary.BigEndian, dataLength)
	return append(dataLengthBuffer.Bytes(), resultBytes[:]...)
}

func unpackMessageFromPackage(source []byte) *messages.Message {

	if len(source) <= int(global.HeaderLength) {
		logrus.Panicf("len(source) <= %d", global.HeaderLength)
	}

	source = source[global.HeaderLength:]

	var msg messages.Message
	decoder := gob.NewDecoder(bytes.NewReader(source))
	err := decoder.Decode(&msg)
	if err != nil {
		logrus.Panic(err)
	}

	return &msg
}

func unpackMessage(source []byte, readedMessage chan *messages.Message, remoteAddr *net.UDPAddr) []byte {
	length := len(source)
	if length == 0 {
		return source
	}

	// 分包
	var i uint32
	for i = 0; i < uint32(length); i = i + 1 {
		// 获取包长信息
		data := source[0:global.HeaderLength]
		packageLength := binary.BigEndian.Uint32(data)

		// 获取包数据
		offset := i + global.HeaderLength
		if offset+packageLength > uint32(length) {
			break
		}
		packageData := source[offset : offset+packageLength]

		var msg messages.Message
		decoder := gob.NewDecoder(bytes.NewReader(packageData))
		err := decoder.Decode(&msg)
		if err != nil {
			logrus.Panic(err)
		}

		if remoteAddr != nil {
			msg.RemoteAddr = remoteAddr
		}

		readedMessage <- &msg
		i += global.HeaderLength + packageLength - 1
	}

	if i == uint32(length) {
		return make([]byte, 0)
	}
	return source[i:]
}
