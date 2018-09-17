package blockchain

import (
	"encoding/gob"
	"bytes"
	"github.com/sirupsen/logrus"
	"encoding/hex"
	"crypto/sha256"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/elliptic"
	"math/big"
	"fmt"
)

var (
	RewardAmount = 50
)

type Input struct {
	TransactionID []byte
	OutputIndex   int
	Signature     []byte
	PublicKey     []byte
	ScriptSig     string
}

type Output struct {
	Amount        int
	PublicKeyHash []byte
	ScriptPubKey  string
}

func (in *Input) IsLockedWithKey(publicKeyHash []byte) bool {
	lockingHash := HashPublicKey(in.PublicKey)
	return bytes.Compare(publicKeyHash, lockingHash) == 0
}

func (out *Output) Lock(address []byte) {
	publicKeyHash := Base58Decode(address)
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-AddressChecksumLen]
	out.PublicKeyHash = publicKeyHash
}

func (out *Output) IsLockedWithKey(publicKeyHash []byte) bool {
	return bytes.Compare(out.PublicKeyHash, publicKeyHash) == 0
}

func NewOutput(amount int, address string) Output {
	output := Output{
		Amount:        amount,
		PublicKeyHash: nil,
	}
	output.Lock([]byte(address))

	return output
}

type Transaction struct {
	ID      []byte
	Inputs  []Input
	Outputs []Output
}

type TransactionContainer []Transaction

func (trans TransactionContainer) Hash() []byte {
	var transactionsHash [][]byte

	for _, tran := range trans {
		transactionsHash = append(transactionsHash, tran.ID)
	}

	mTree := NewMerkleTree(transactionsHash)

	return mTree.RootNode.Data
}

func (trans TransactionContainer) Serialize() []byte {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(trans)
	if err != nil {
		logrus.Panicf("TransactionContainer Serialize err: %v", err)
	}

	return buffer.Bytes()
}

func DeserializeTransactionContainer(data []byte) TransactionContainer {
	trans := TransactionContainer{}
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&trans)
	if err != nil {
		logrus.Panicf("DeserializeTransactionContainer err: %v", err)
	}

	return trans
}

func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TransactionID) == 0 && tx.Inputs[0].OutputIndex == -1
}

func (tx *Transaction) Hash() {
	txCopy := *tx
	txCopy.ID = []byte{}
	hash := sha256.Sum256(txCopy.Serialize())
	tx.ID = hash[:]
}

func (tx *Transaction) Serialize() []byte {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(*tx)
	if err != nil {
		logrus.Panicf("Transaction Serialize err: %v", err)
	}

	return buffer.Bytes()
}

func DeserializeTransaction(data []byte) Transaction {
	tran := Transaction{}
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tran)
	if err != nil {
		logrus.Panicf("DeserializeTransaction err: %v", err)
	}

	return tran
}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTrans map[string]*Transaction) {
	if tx.IsCoinBase() {
		return
	}

	copyTran := tx.TrimmedCopy()

	for inputIndex, input := range copyTran.Inputs {
		prevTran := prevTrans[hex.EncodeToString(input.TransactionID)]
		if prevTran == nil {
			logrus.Panicf("Transaction sign err: cant find previous transaction ")
		}
		copyTran.Inputs[inputIndex].PublicKey = prevTran.Outputs[input.OutputIndex].PublicKeyHash
		copyTran.Hash()
		copyTran.Inputs[inputIndex].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, copyTran.ID)
		if err != nil {
			logrus.Panicf("Transaction ecdsa.Sign err: %v", err)
		}
		signature := bytes.Join([][]byte{r.Bytes(), s.Bytes()}, []byte{})
		tx.Inputs[inputIndex].Signature = signature
	}
}

func (tx *Transaction) Verify(prevTrans map[string]*Transaction) bool {
	copyTran := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inputIndex, input := range tx.Inputs {
		prevTran := prevTrans[hex.EncodeToString(input.TransactionID)]
		if prevTran == nil {
			logrus.Panicf("Transaction sign err: cant find previous transaction ")
		}
		copyTran.Inputs[inputIndex].PublicKey = prevTran.Outputs[input.OutputIndex].PublicKeyHash
		copyTran.Hash()
		copyTran.Inputs[inputIndex].PublicKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(input.Signature)
		r.SetBytes(input.Signature[:sigLen/2])
		s.SetBytes(input.Signature[sigLen/2:])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(input.PublicKey)
		x.SetBytes(input.PublicKey[:(keyLen / 2)])
		y.SetBytes(input.PublicKey[(keyLen / 2):])

		rawPublicKey := &ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}
		if !ecdsa.Verify(rawPublicKey, copyTran.ID, &r, &s) {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() *Transaction {
	inputs := make([]Input, 0)
	outputs := make([]Output, 0)

	for _, input := range tx.Inputs {
		inputs = append(inputs, Input{input.TransactionID, input.OutputIndex, nil, nil, ""})
	}

	for _, output := range tx.Outputs {
		outputs = append(outputs, Output{output.Amount, output.PublicKeyHash, output.ScriptPubKey})
	}

	return &Transaction{tx.ID, inputs, outputs}
}

func newGenesisBlock(coinbase Transaction) *Block {
	trans := make(TransactionContainer, 0)
	trans = append(trans, coinbase)
	return NewBlock(trans.Serialize(), []byte{}, 0)
}

func NewTransaction(from, to string, amount int, c *BlockChain) Transaction {
	inputs := make([]Input, 0)
	outputs := make([]Output, 0)
	chainState := ChainState{c}

	wallets, err := NewWalletManager()
	if err != nil {
		logrus.Panicf("NewTransaction GetWallet err: %v", err)
	}
	wallet := wallets.GetWallet(from)
	if wallet == nil {
		logrus.Panicf("NewTransaction GetWallet not exist: %s", from)
	}
	publicKeyHash := HashPublicKey(wallet.PublicKey)

	unspentAmount, unspentOutputs := chainState.FindSpendableOutputsMapping(publicKeyHash, amount)

	if unspentAmount < amount {
		logrus.Panic("not enough funds")
	}

	for tranIDEncoded, outIndexs := range unspentOutputs {
		tranID, err := hex.DecodeString(tranIDEncoded)
		if err != nil {
			logrus.Panicf("NewTransaction hex.DecodeString err: %v", err)
		}

		for _, outIndex := range outIndexs {
			input := Input{
				TransactionID: tranID,
				OutputIndex:   outIndex,
				PublicKey:     wallet.PublicKey,
			}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, NewOutput(amount, to))
	if unspentAmount > amount {
		outputs = append(outputs, NewOutput(unspentAmount-amount, from))
	}

	tran := Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	tran.Hash()
	SignTransaction(c, &tran, wallet.PrivateKey)

	return tran
}

func NewCoinbaseTransaction(to string, data string) Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			logrus.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}
	tran := Transaction{
		Inputs: []Input{
			{
				TransactionID: nil,
				OutputIndex:   -1,
				Signature:     nil,
				PublicKey:     []byte(data),
			},
		},
		Outputs: []Output{NewOutput(RewardAmount, to)},
	}
	tran.Hash()

	return tran
}

func findTransaction(c *BlockChain, ID []byte) *Transaction {
	it := c.Iterator()

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
			logrus.Panicf("FindUnspentTransactions block data cant be decoded")
		}

		for _, tran := range trans {
			if bytes.Compare(ID, tran.ID) == 0 {
				return &tran
			}
		}

		if len(block.Header.PrevBlockHash) == 0 {
			break
		}
	}

	return nil
}

func SignTransaction(c *BlockChain, tran *Transaction, privateKey ecdsa.PrivateKey) {
	prevTrans := make(map[string]*Transaction)

	for _, input := range tran.Inputs {
		prevTran := findTransaction(c, input.TransactionID)
		if prevTran == nil {
			logrus.Panicf("SignTransaction transaction not found: %s", hex.EncodeToString(input.TransactionID))
		}
		prevTrans[hex.EncodeToString(prevTran.ID)] = prevTran
	}

	tran.Sign(privateKey, prevTrans)
}

func VerifyTransaction(c *BlockChain, tran *Transaction) bool {
	if tran.IsCoinBase() {
		return true
	}
	prevTrans := make(map[string]*Transaction)

	for _, input := range tran.Inputs {
		prevTran := findTransaction(c, input.TransactionID)
		if prevTran == nil {
			logrus.Panicf("SignTransaction transaction not found: %s", hex.EncodeToString(input.TransactionID))
		}
		prevTrans[hex.EncodeToString(prevTran.ID)] = prevTran
	}

	return tran.Verify(prevTrans)
}
