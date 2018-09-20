package blockchain

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"sync"
	"github.com/johnnyeven/chain/global"
)

type BlockChain struct {
	// 最后一个区块的Hash
	tip []byte
	// 数据库连接
	DB *bolt.DB

	NewGenesisBlockFunc func(to string, data string) *Block
	// 同步锁
	lock *sync.Mutex
}

var GeneralChain *BlockChain

func (c *BlockChain) PackageBlock(data []byte) *Block {
	//c.lock.Lock()
	//defer c.lock.Unlock()

	var lastHash []byte
	var lastHeight uint64

	c.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.BlockBucketIdentity))
		lastHash = bucket.Get([]byte(global.ChainLastIndexKey))

		lastBlock := DeserializeBlock(bucket.Get(lastHash))
		lastHeight = lastBlock.Header.Height

		return nil
	})

	block := NewBlock(data, lastHash, lastHeight+1)
	c.AddBlock(block)

	chainState := ChainState{c}
	chainState.Update(block)

	return block
}

func (c *BlockChain) AddBlock(block *Block) {
	//c.lock.Lock()
	//defer c.lock.Unlock()

	pow := NewPOW(block)
	if !pow.Validate() {
		logrus.Warnf("AddBlock pow invalid!")
		return
	}
	err := c.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.BlockBucketIdentity))
		blockExist := bucket.Get(block.Header.Hash)
		if blockExist != nil {
			return nil
		}

		err := bucket.Put(block.Header.Hash, block.Serialize())
		if err != nil {
			logrus.Errorf("BlockChain.AddBlock bucket.Put block err: %v", err)
			return err
		}

		bestHeight := c.GetBestHeight()
		if block.Header.Height > bestHeight {
			err = bucket.Put([]byte(global.ChainLastIndexKey), block.Header.Hash)
			if err != nil {
				logrus.Errorf("BlockChain.AddBlock bucket.Put last hash err: %v", err)
				return err
			}
			c.tip = block.Header.Hash
		}

		return nil
	})

	if err != nil {
		logrus.Panicf("BlockChain.AddBlock bucket transaction err: %v", err)
	}
}

func (c *BlockChain) Iterator() *BlockChainIterator {
	it := &BlockChainIterator{
		currentHash: c.tip,
		db:          c.DB,
	}

	return it
}

func (c *BlockChain) GetBestHeight() uint64 {
	//c.lock.Lock()
	//defer c.lock.Unlock()

	var lastHeight uint64
	err := c.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.BlockBucketIdentity))
		lastHash := bucket.Get([]byte(global.ChainLastIndexKey))

		if lastHash != nil && len(lastHash) > 0 {
			lastBlock := DeserializeBlock(bucket.Get(lastHash))
			lastHeight = lastBlock.Header.Height
			return nil
		}

		return nil
	})

	if err != nil {
		logrus.Panicf("BlockChain.GetBestHeight bucket transaction err: %v", err)
	}

	return lastHeight
}

func (c *BlockChain) GetBlock(blockHash []byte) *Block {
	//c.lock.Lock()
	//defer c.lock.Unlock()

	var block *Block
	err := c.DB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.BlockBucketIdentity))
		blockData := bucket.Get(blockHash)
		if blockData == nil || len(blockData) == 0 {
			return nil
		}
		block = DeserializeBlock(blockData)
		return nil
	})

	if err != nil {
		logrus.Panicf("BlockChain.GetBlock bucket transaction err: %v", err)
	}

	return block
}

func InitBlockChain(address string, config Config) *BlockChain {
	var (
		db  *bolt.DB
		tip []byte
		err error
	)
	if GeneralChain != nil {
		db = GeneralChain.DB
	} else {
		db, err = bolt.Open(global.GetBlockFilePath(), 0600, nil)
		if err != nil {
			logrus.Panicf("InitBlockChain bolt.Open err: %v", err)
		}
		GeneralChain = &BlockChain{
			DB:                  db,
			NewGenesisBlockFunc: config.NewGenesisBlockFunc,
			lock:                new(sync.Mutex),
		}
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.BlockBucketIdentity))
		if bucket == nil || bucket.Get([]byte(global.ChainLastIndexKey)) == nil {
			// 数据库未初始化，创建初始区块
			if address == "" {
				logrus.Panic("chain not initialized")
			}

			block := GeneralChain.NewGenesisBlockFunc(address, "This is the CoinBase transaction")

			if bucket == nil {
				bucket, err = tx.CreateBucket([]byte(global.BlockBucketIdentity))
				if err != nil {
					logrus.Errorf("InitBlockChain tx.CreateBucket err: %v", err)
					return err
				}
			}

			err = bucket.Put(block.Header.Hash, block.Serialize())
			if err != nil {
				logrus.Errorf("InitBlockChain bucket.Put block err: %v", err)
				return err
			}
			err = bucket.Put([]byte(global.ChainLastIndexKey), block.Header.Hash)
			if err != nil {
				logrus.Errorf("InitBlockChain bucket.Put last hash err: %v", err)
				return err
			}
			tip = block.Header.Hash
		} else {
			tip = bucket.Get([]byte(global.ChainLastIndexKey))
		}
		return nil
	})

	if err != nil {
		logrus.Panicf("InitBlockChain bucket transaction err: %v", err)
	}

	GeneralChain.tip = tip

	return GeneralChain
}

func NewBlockChain(config Config) *BlockChain {
	if GeneralChain != nil {
		return GeneralChain
	}
	var tip []byte
	db, err := bolt.Open(global.GetBlockFilePath(), 0600, nil)
	if err != nil {
		logrus.Panicf("NewBlockChain bolt.Open err: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.BlockBucketIdentity))
		if bucket == nil {
			_, err := tx.CreateBucket([]byte(global.BlockBucketIdentity))
			if err != nil {
				logrus.Panicf("NewBlockChain CreateBucket err: %v", err)
			}
		} else {
			tip = bucket.Get([]byte(global.ChainLastIndexKey))
		}
		return nil
	})

	if err != nil {
		logrus.Panicf("NewBlockChain bucket transaction err: %v", err)
	}

	GeneralChain = &BlockChain{
		tip:                 tip,
		DB:                  db,
		NewGenesisBlockFunc: config.NewGenesisBlockFunc,
		lock:                new(sync.Mutex),
	}

	return GeneralChain
}

type Config struct {
	NewGenesisBlockFunc func(to string, data string) *Block
}
