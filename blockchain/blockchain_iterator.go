package blockchain

import (
	"github.com/boltdb/bolt"
	"git.profzone.net/profzone/chain/global"
)

type BlockChainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (it *BlockChainIterator) Next() *Block {
	var block *Block

	it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.BlockBucketIdentity))
		data := bucket.Get(it.currentHash)
		if data != nil && len(data) > 0 {
			block = DeserializeBlock(data)
		}

		return nil
	})

	if block != nil {
		it.currentHash = block.Header.PrevBlockHash
	}

	return block
}
