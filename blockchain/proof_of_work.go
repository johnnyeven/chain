package blockchain

import (
	"math/big"
	"bytes"
	"math"
	"crypto/sha256"
	"github.com/sirupsen/logrus"
	"encoding/gob"
	"github.com/johnnyeven/chain/global"
)

var (
	difficulty = 12
	maxNonce uint64 = math.MaxUint64
)

type POW struct {
	block  *Block
	target *big.Int
}

func NewPOW(block *Block) *POW {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))

	pow := &POW{
		block:  block,
		target: target,
	}

	return pow
}

func (pow *POW) prepareData(nonce uint64) []byte {
	buffer := bytes.NewReader(pow.block.Body.Data)
	decoder := gob.NewDecoder(buffer)
	trans := make(TransactionContainer, 0)
	err := decoder.Decode(&trans)
	if err != nil {
		logrus.Panicf("pow prepareData decode err: %v", err)
	}
	data := bytes.Join([][]byte{
		pow.block.Header.PrevBlockHash,
		trans.Hash(),
		global.IntToByte(pow.block.Header.Timestamp),
		global.IntToByte(uint64(difficulty)),
		global.IntToByte(nonce),
	}, []byte{})
	return data
}

func (pow *POW) hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

func (pow *POW) Run() (nonce uint64, hash []byte) {
	var hashInt big.Int

	logrus.Debugf("开始验证区块数据: %s", string(pow.block.Body.Data))
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = pow.hash(data)
		hashInt.SetBytes(hash)
		if hashInt.Cmp(pow.target) == -1 {
			logrus.Debugf("找到合适的 Nonce: %d", nonce)
			logrus.Debugf("%x", hash)
			break
		}
		nonce++
	}
	return
}

func (pow *POW) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Header.Nonce)
	hash := pow.hash(data)
	hashInt.SetBytes(hash)
	if hashInt.Cmp(pow.target) == -1 {
		return true
	}

	return false
}
