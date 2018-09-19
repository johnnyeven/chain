package blockchain

import (
	"os"
	"io/ioutil"
	"github.com/sirupsen/logrus"
	"bytes"
	"encoding/gob"
	"crypto/elliptic"
	"github.com/profzone/chain/global"
)

type WalletManager struct {
	W map[string]*Wallet
}

func init() {
	gob.Register(elliptic.P256())
}

func (w *WalletManager) LoadFromFile() error {
	walletFileName := global.GetWalletFilePath()
	if _, err := os.Stat(walletFileName); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFileName)
	if err != nil {
		logrus.Panicf("WalletManager LoadFromFile err: %v", err)
	}

	reader := bytes.NewReader(fileContent)
	decoder := gob.NewDecoder(reader)
	err = decoder.Decode(w)
	if err != nil {
		logrus.Panicf("WalletManager LoadFromFile err: %v", err)
	}

	return nil
}

func (w *WalletManager) SaveToFile() {
	var buffer bytes.Buffer
	walletFileName := global.GetWalletFilePath()

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(w)
	if err != nil {
		logrus.Panicf("WalletManager SaveToFile err: %v", err)
	}

	err = ioutil.WriteFile(walletFileName, buffer.Bytes(), 0644)
	if err != nil {
		logrus.Panicf("WalletManager SaveToFile err: %v", err)
	}
}

func (w *WalletManager) GetWallet(address string) *Wallet {
	return w.W[address]
}

func (w *WalletManager) GetAddress() []string {
	var addresses []string

	for address := range w.W {
		addresses = append(addresses, address)
	}

	return addresses
}

func (w *WalletManager) CreateWallet(alias string) string {
	wallet := NewWallet(alias)
	address := wallet.GetAddress()

	w.W[string(address)] = wallet
	return string(address)
}

func NewWalletManager() (*WalletManager, error) {
	wallets := &WalletManager{
		W: make(map[string]*Wallet),
	}
	err := wallets.LoadFromFile()
	return wallets, err
}
