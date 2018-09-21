package services

import (
	"github.com/johnnyeven/chain/blockchain"
	"github.com/johnnyeven/chain/messages"
	"github.com/johnnyeven/chain/global"
	"github.com/johnnyeven/terra/dht"
	"github.com/johnnyeven/chain/network"
	"github.com/sirupsen/logrus"
	"github.com/boltdb/bolt"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
)

var _ interface {
	Service
} = (*BlockChainService)(nil)

var blockChainService *BlockChainService

type blockInTransport struct {
	blocks                   *dht.SyncedMap
	heightIndexedBlockHashed map[uint64][][]byte
	sortedHeight             []uint64
}

func (b *blockInTransport) Get(hash []byte) (*blockchain.Block, bool) {
	val, ok := b.blocks.Get(string(hash))
	return val.(*blockchain.Block), ok
}

func (b *blockInTransport) Has(hash []byte) bool {
	return b.blocks.Has(string(hash))
}

func (b *blockInTransport) Set(block *blockchain.Block) {
	b.blocks.Set(string(block.Header.Hash), block)

	if _, ok := b.heightIndexedBlockHashed[block.Header.Height]; !ok {
		b.sortedHeight = append(b.sortedHeight, block.Header.Height)
	}
	b.heightIndexedBlockHashed[block.Header.Height] = append(b.heightIndexedBlockHashed[block.Header.Height], block.Header.Hash)
}

func (b *blockInTransport) Delete(hash []byte) {
	block, blockExist := b.Get(hash)

	if hashes, ok := b.heightIndexedBlockHashed[block.Header.Height]; ok {
		for i, h := range hashes {
			if bytes.Compare(hash, h) == 0 {
				b.heightIndexedBlockHashed[block.Header.Height] = append(b.heightIndexedBlockHashed[block.Header.Height][:i], b.heightIndexedBlockHashed[block.Header.Height][i+1:]...)
			}
		}
	}

	if blockExist {
		b.blocks.Delete(string(hash))
	}
}

func (b *blockInTransport) DeleteMulti(hashes [][]byte) {
	for _, hash := range hashes {
		b.Delete(hash)
	}
}

func (b *blockInTransport) Clear() {
	b.blocks.Clear()
}

func (b *blockInTransport) Iterator(iterator func(block *blockchain.Block) error, errorContinue bool) {
Run:
	for _, height := range b.sortedHeight {
		hashes := b.heightIndexedBlockHashed[height]
		for _, hash := range hashes {
			block, ok := b.Get(hash)
			if ok {
				err := iterator(block)
				if err != nil {
					if errorContinue {
						continue
					} else {
						break Run
					}
				}
			}
		}
	}
}

func (b *blockInTransport) Len() int {
	return b.blocks.Len()
}

func newBlockInTransport() *blockInTransport {
	return &blockInTransport{
		blocks:                   dht.NewSyncedMap(),
		heightIndexedBlockHashed: make(map[uint64][][]byte),
		sortedHeight:             make([]uint64, 0),
	}
}

type BlockChainService struct {
	c                     *blockchain.BlockChain
	blockInTransport      *blockInTransport
	signalQuit            chan struct{}
	signalRequestHeight   chan struct{}
	signalSendTransaction chan *blockchain.Transaction
}

func NewBlockChainService() Service {
	if blockChainService == nil {
		blockChainService = &BlockChainService{
			c: blockchain.NewBlockChain(blockchain.Config{
				NewGenesisBlockFunc: blockchain.NewGenesisBlock,
			}),
			blockInTransport:      newBlockInTransport(),
			signalQuit:            make(chan struct{}),
			signalRequestHeight:   make(chan struct{}),
			signalSendTransaction: make(chan *blockchain.Transaction),
		}
		chainState := blockchain.ChainState{BlockChain: blockChainService.c}
		chainState.Reindex()
	}
	return blockChainService
}

func GetBlockChainService() *BlockChainService {
	if blockChainService == nil {
		NewBlockChainService()
	}
	return blockChainService
}

func (s *BlockChainService) GetTransChannel() chan<- *blockchain.Transaction {
	return s.signalSendTransaction
}

func (s *BlockChainService) Messages() []messages.MessageHandler {
	return []messages.MessageHandler{
		{
			Type:   global.MESSAGE_TYPE__REQUEST_HEIGHT,
			Runner: s.RunRequestHeight,
		},
		{
			Type:   global.MESSAGE_TYPE__BLOCKS_HASH,
			Runner: s.RunBlocksHash,
		},
		{
			Type:   global.MESSAGE_TYPE__GET_BLOCK,
			Runner: s.RunGetBlock,
		},
		{
			Type:   global.MESSAGE_TYPE__GET_BLOCK_ACK,
			Runner: s.RunGetBlockAck,
		},
		{
			Type:   global.MESSAGE_TYPE__NEW_TRANSACTION,
			Runner: s.RunNewTransaction,
		},
	}
}

func (s *BlockChainService) Start() error {
	go func() {
	Run:
		for {
			select {
			case <-s.signalQuit:
				break Run
			case <-s.signalRequestHeight:
				go RequestHeight(s.c, 0)
			case tran := <-s.signalSendTransaction:
				BroadcastTran(tran)
			}
		}
	}()
	return nil
}

func (s *BlockChainService) Stop() error {
	s.signalQuit <- struct{}{}
	close(s.signalQuit)
	return nil
}

func (s *BlockChainService) RunRequestHeight(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.RequestHeight{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	peer := t.GetClient().(*network.ChainProtobufClient).GetPeer()
	currentHeight := s.c.GetBestHeight()
	if payload.Height > currentHeight {

		// 对方区块比我方更新，请求对方的区块
		message := &messages.RequestHeight{
			Height:  currentHeight,
			Version: global.Config.Version,
		}
		request := t.MakeResponse(peer.Guid, peer.Node.Addr, msg.MessageID, message)
		t.Request(request)

	} else if payload.Height < currentHeight {

		// 我方区块比对方更新，发送给对方缺失的区块哈希
		blockHashes := make([][]byte, 0)
		it := s.c.Iterator()

		// TODO 优化算法，不用遍历整条链
		for {
			block := it.Next()
			if block == nil {
				break
			}

			if block.Header.Height >= payload.Height {
				blockHashes = append(blockHashes, block.Header.Hash)
			}

			if block.Header.PrevBlockHash == nil || len(block.Header.PrevBlockHash) == 0 {
				break
			}
		}
		message := &messages.BlocksHash{
			Hashes: blockHashes,
		}
		request := t.MakeResponse(peer.Guid, peer.Node.Addr, msg.MessageID, message)
		t.Request(request)
	}

	return nil
}

func (s *BlockChainService) RunBlocksHash(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.BlocksHash{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	peer := t.GetClient().(*network.ChainProtobufClient).GetPeer()
	for _, hash := range payload.Hashes {

		blockExist := s.c.GetBlock(hash)
		if blockExist != nil {
			continue
		}
		message := &messages.GetBlock{
			Hash: hash,
		}
		request := t.MakeRequest(peer.Guid, peer.Node.Addr, "", message)
		t.Request(request)

	}

	return nil
}

func (s *BlockChainService) RunGetBlock(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.GetBlock{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		logrus.Errorf("[RunGetBlock] payload.DecodeFromSource err: %v", err)
		return err
	}

	block := s.c.GetBlock(payload.Hash)
	message := &messages.GetBlockAck{
		Block: block.Serialize(),
	}

	peer := t.GetClient().(*network.ChainProtobufClient).GetPeer()
	request := t.MakeResponse(peer.Guid, peer.Node.Addr, msg.MessageID, message)
	t.Request(request)

	return nil
}

func (s *BlockChainService) RunGetBlockAck(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.GetBlockAck{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	block := blockchain.DeserializeBlock(payload.Block)
	logrus.Debugf("received a new block: %x", block.Header.Hash)

	s.verifyAndAddBlock(block, msg)

	s.blockInTransport.Iterator(func(b *blockchain.Block) error {
		if b == block {
			return nil
		}
		s.verifyAndAddBlock(b, msg)
		return nil
	}, true)

	return nil
}

func (s *BlockChainService) verifyAndAddBlock(block *blockchain.Block, msg *messages.Message) {
	err := s.c.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.ChainStateBucketIdentity))

		decoder := gob.NewDecoder(bytes.NewReader(block.Body.Data))
		trans := make(blockchain.TransactionContainer, 0)
		err := decoder.Decode(&trans)
		if err != nil {
			logrus.Panicf("RunGetBlockAck error: block data cant be decoded")
		}

		for _, tran := range trans {
			if tran.IsCoinBase() {
				continue
			}
			for _, input := range tran.Inputs {
				serializedOutputs := bucket.Get(input.TransactionID)
				if serializedOutputs == nil || len(serializedOutputs) == 0 {
					if !s.blockInTransport.Has(block.Header.Hash) {
						s.blockInTransport.Set(block)
					}
					return errors.New(fmt.Sprintf("%x block's trans not found, set into map", block.Header.Hash))
				}
			}
		}

		return nil
	})

	if err != nil {
		logrus.Warningf("RunGetBlockAck error: %v", err)
		return
	}

	if s.blockInTransport.Has(block.Header.Hash) {
		s.blockInTransport.Delete(block.Header.Hash)
	}

	ok := s.c.AddBlock(block)
	if ok {

		chainState := blockchain.ChainState{BlockChain: s.c}
		chainState.Update(block)

		BroadcastBlock(block, msg)
	}
}

func (s *BlockChainService) RunNewTransaction(t *dht.Transport, msg *messages.Message) error {

	payload := &messages.NewTransaction{}
	err := payload.DecodeFromSource(msg.Payload)
	if err != nil {
		return err
	}

	tran := blockchain.DeserializeTransaction(payload.Transaction)

	go func() {
		trans := make(blockchain.TransactionContainer, 0)
		trans = append(trans, tran)
		trans = append(trans, blockchain.NewCoinbaseTransaction(global.Config.ReceiveAddress, ""))

		for _, tran := range trans {
			if !blockchain.VerifyTransaction(s.c, &tran) {
				logrus.Warningf("invalid transaction: %s", tran.ID)
			}
		}
		block := s.c.PackageBlock(trans.Serialize())
		BroadcastBlock(block, msg)
	}()

	return nil
}

func BroadcastBlock(block *blockchain.Block, msg *messages.Message) {
	network.P2P.GetPeerManager().Iterator(func(peer *network.Peer) error {

		message := &messages.BlocksHash{
			Hashes: [][]byte{block.Header.Hash},
		}
		t := peer.GetTransport()
		request := t.MakeResponse(peer.Guid, peer.Node.Addr, msg.MessageID, message)
		t.Request(request)

		return nil

	}, true)
}

func BroadcastTran(tran *blockchain.Transaction) {
	network.P2P.GetPeerManager().Iterator(func(peer *network.Peer) error {

		message := &messages.NewTransaction{
			Transaction: tran.Serialize(),
		}
		t := peer.GetTransport()
		request := t.MakeRequest(peer.Guid, peer.Node.Addr, "", message)
		t.Request(request)

		return nil

	}, true)
}

func RequestHeight(c *blockchain.BlockChain, height uint64) {
	network.P2P.GetPeerManager().Iterator(func(peer *network.Peer) error {

		if height == 0 {
			height = c.GetBestHeight()
		}

		message := &messages.RequestHeight{
			Height:  height,
			Version: global.Config.Version,
		}
		t := peer.GetTransport()
		request := t.MakeRequest(peer.Guid, peer.Node.Addr, "", message)
		t.Request(request)

		return nil

	}, true)
}

func RequestHeightTask() {
	service := GetBlockChainService()
	if service != nil {
		service.signalRequestHeight <- struct{}{}
	}
}
