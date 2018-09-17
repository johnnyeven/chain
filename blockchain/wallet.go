package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"github.com/sirupsen/logrus"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"bytes"
)

const (
	Version            = byte(0x00)
	AddressChecksumLen = 4
)

type Wallet struct {
	Alias      string
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet(alias string) *Wallet {
	private, public := newKeyPair()
	wallet := &Wallet{
		PrivateKey: private,
		PublicKey:  public,
	}

	if alias != "" {
		wallet.Alias = alias
	}

	return wallet
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		logrus.Panicf("NewWallet ecdsa.GenerateKey err: %v", err)
	}

	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

func (w Wallet) GetAddress() []byte {
	publicKeyHash := HashPublicKey(w.PublicKey)

	payload := bytes.Join([][]byte{{Version}, publicKeyHash}, []byte{})
	checkSum := checkSum(payload)

	fullPayload := bytes.Join([][]byte{payload, checkSum}, []byte{})
	address := Base58Encode(fullPayload)

	return address
}

func ValidateAddress(address string) bool {
	publicKeyHash := Base58Decode([]byte(address))
	originCheckSum := publicKeyHash[len(publicKeyHash)-AddressChecksumLen:]
	version := publicKeyHash[0]
	publicKeyHash = publicKeyHash[1 : len(publicKeyHash)-AddressChecksumLen]
	targetCheckSum := checkSum(bytes.Join([][]byte{{version}, publicKeyHash}, []byte{}))

	return bytes.Compare(originCheckSum, targetCheckSum) == 0
}

func HashPublicKey(publicKey []byte) []byte {
	publicSHA256 := sha256.Sum256(publicKey)
	hasher := ripemd160.New()
	_, err := hasher.Write(publicSHA256[:])
	if err != nil {
		logrus.Panicf("HashPublicKey err: %v", err)
	}
	return hasher.Sum(nil)
}

func checkSum(payload []byte) []byte {
	firstSum := sha256.Sum256(payload)
	secondSum := sha256.Sum256(firstSum[:])
	return secondSum[:AddressChecksumLen]
}
