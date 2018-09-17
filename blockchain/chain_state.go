package blockchain

import (
	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"git.profzone.net/profzone/chain/global"
)

type ChainState struct {
	BlockChain *BlockChain
}

func (s ChainState) Reindex() error {
	db := s.BlockChain.DB
	bucketName := []byte(global.ChainStateBucketIdentity)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && err != bolt.ErrBucketNotFound {
			logrus.Panicf("ChainState Reindex err: the bucket not found: %s", bucketName)
		}
		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			logrus.Panicf("ChainState Reindex err: %v", err)
		}

		return nil
	})
	if err != nil {
		logrus.Errorf("ChainState Reindex err: %v", err)
		return err
	}

	unspentOutputs := s.findUnspentOutputs()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)

		for tranID, outputs := range unspentOutputs {
			key, err := hex.DecodeString(tranID)
			if err != nil {
				logrus.Panicf("ChainState Reindex err: %v", err)
			}

			var buffer bytes.Buffer
			encoder := gob.NewEncoder(&buffer)
			err = encoder.Encode(outputs)
			if err != nil {
				logrus.Panicf("ChainState Reindex err: %v", err)
			}

			err = bucket.Put(key, buffer.Bytes())
			if err != nil {
				logrus.Panicf("ChainState Reindex err: %v", err)
			}
		}

		return nil
	})

	return nil
}

func (s ChainState) FindSpendableOutputsMapping(publicKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	spendableAmount := 0

	db := s.BlockChain.DB
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.ChainStateBucketIdentity))
		cursor := bucket.Cursor()

	Work:
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			tranID := hex.EncodeToString(key)
			outputs := make([]Output, 0)
			decoder := gob.NewDecoder(bytes.NewReader(value))
			err := decoder.Decode(&outputs)
			if err != nil {
				logrus.Panicf("ChainState FindSpendableOutputsMapping decoder err: %v", err)
			}

			for outIndex, output := range outputs {
				if output.IsLockedWithKey(publicKeyHash) {
					if amount > 0 && spendableAmount >= amount {
						break Work
					}
					spendableAmount += output.Amount
					unspentOutputs[tranID] = append(unspentOutputs[tranID], outIndex)
				}
			}
		}

		return nil
	})
	if err != nil {
		logrus.Panicf("ChainState FindSpendableOutputsMapping transaction err: %v", err)
	}

	return spendableAmount, unspentOutputs
}

func (s ChainState) FindSpendableOutputs(publicKeyHash []byte) (int, []Output) {
	unspentOutputs := make([]Output, 0)
	spendableAmount := 0

	db := s.BlockChain.DB
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.ChainStateBucketIdentity))
		cursor := bucket.Cursor()

		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			outputs := make([]Output, 0)
			decoder := gob.NewDecoder(bytes.NewReader(value))
			err := decoder.Decode(&outputs)
			if err != nil {
				logrus.Panicf("ChainState FindSpendableOutputsMapping decoder err: %v", err)
			}

			for _, output := range outputs {
				if output.IsLockedWithKey(publicKeyHash) {
					spendableAmount += output.Amount
					unspentOutputs = append(unspentOutputs, output)
				}
			}
		}

		return nil
	})
	if err != nil {
		logrus.Panicf("ChainState FindSpendableOutputsMapping transaction err: %v", err)
	}

	return spendableAmount, unspentOutputs
}

func (s ChainState) findUnspentOutputs() map[string][]Output {
	unspentOutputs := make(map[string][]Output)
	spentOutputs := make(map[string][]int)
	it := s.BlockChain.Iterator()

	for {
		block := it.Next()
		if block == nil {
			break
		}

		buffer := bytes.NewReader(block.Body.Data)
		decoder := gob.NewDecoder(buffer)
		trans := make(TransactionContainer, 0)
		err := decoder.Decode(&trans)
		if err != nil {
			logrus.Panicf("FindUnspentTransactions block data decode err: %v", err)
		}

		for _, tran := range trans {
			if !tran.IsCoinBase() {
				for _, in := range tran.Inputs {
					tranID := hex.EncodeToString(in.TransactionID)
					spentOutputs[tranID] = append(spentOutputs[tranID], in.OutputIndex)
				}
			}

			tranID := hex.EncodeToString(tran.ID)

		Outputs:
			for outIndex, out := range tran.Outputs {
				if spentOutputs[tranID] != nil {
					for _, spentOutputIndex := range spentOutputs[tranID] {
						if spentOutputIndex == outIndex {
							continue Outputs
						}
					}
				}

				unspentOutputs[tranID] = append(unspentOutputs[tranID], out)
			}
		}

		if len(block.Header.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentOutputs
}

func (s ChainState) Update(block *Block) {
	db := s.BlockChain.DB

	err := db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(global.ChainStateBucketIdentity))

		decoder := gob.NewDecoder(bytes.NewReader(block.Body.Data))
		trans := make(TransactionContainer, 0)
		err := decoder.Decode(&trans)
		if err != nil {
			logrus.Panicf("ChainState Update err: block data cant be decoded")
		}

		for _, tran := range trans {
			if !tran.IsCoinBase() {
				for _, input := range tran.Inputs {
					updatedOutputs := make([]Output, 0)
					serializedOutputs := bucket.Get(input.TransactionID)
					deserializedOutputs := make([]Output, 0)
					decoder := gob.NewDecoder(bytes.NewReader(serializedOutputs))
					err := decoder.Decode(&deserializedOutputs)
					if err != nil {
						logrus.Panicf("ChainState Update decoder err: %v", err)
					}

					for outputIndex, output := range deserializedOutputs {
						if outputIndex != input.OutputIndex {
							updatedOutputs = append(updatedOutputs, output)
						}
					}

					if len(updatedOutputs) == 0 {
						err := bucket.Delete(input.TransactionID)
						if err != nil {
							logrus.Panicf("ChainState Update bucket.Delete err: %v", err)
						}
					} else {
						var buffer bytes.Buffer
						encoder := gob.NewEncoder(&buffer)
						err = encoder.Encode(updatedOutputs)
						if err != nil {
							logrus.Panicf("ChainState Reindex err: %v", err)
						}
						err := bucket.Put(input.TransactionID, buffer.Bytes())
						if err != nil {
							logrus.Panicf("ChainState Update bucket.Put err: %v", err)
						}
					}
				}
			}

			var buffer bytes.Buffer
			encoder := gob.NewEncoder(&buffer)
			err = encoder.Encode(tran.Outputs)
			if err != nil {
				logrus.Panicf("ChainState Update err: %v", err)
			}

			err := bucket.Put(tran.ID, buffer.Bytes())
			if err != nil {
				logrus.Panicf("ChainState Update bucket.Put err: %v", err)
			}
		}

		return nil
	})
	if err != nil {
		logrus.Panicf("ChainState Update transaction err: %v", err)
	}
}
