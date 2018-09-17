package blockchain

import (
	"bytes"
	"encoding/gob"
	"github.com/sirupsen/logrus"
	"time"
)

type BlockHeader struct {
	PrevBlockHash []byte
	Hash          []byte
	Nonce         uint64
	Timestamp     uint64
	Height        uint64
}

type BlockData struct {
	Data []byte
}

type Block struct {
	Header BlockHeader
	Body   BlockData
}

func NewBlock(data []byte, prevBlockHash []byte, height uint64) *Block {
	block := &Block{
		Header: BlockHeader{
			PrevBlockHash: prevBlockHash,
			Timestamp:     uint64(time.Now().Unix()),
			Height:        height,
		},
		Body: BlockData{
			Data: data,
		},
	}
	block.runPOW()

	return block
}

func (b *Block) runPOW() {
	pow := NewPOW(b)
	nonce, hash := pow.Run()

	b.Header.Nonce = nonce
	b.Header.Hash = hash
}

func (b *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(b)
	if err != nil {
		logrus.Panicf("Block.Serialize err: %v", err)
	}
	return buffer.Bytes()
}

func DeserializeBlock(data []byte) *Block {
	var block Block
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	err := decoder.Decode(&block)
	if err != nil {
		logrus.Panicf("DeserializeBlock err: %v", err)
	}
	return &block
}
