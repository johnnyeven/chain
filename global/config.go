package global

import (
	"github.com/boltdb/bolt"
	"time"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/spf13/viper"
	"math"
	"net"
	"github.com/johnnyeven/terra/dht/util"
)

var ConfigFileGroup = ""

var Config = &struct {
	DB                    *bolt.DB
	SeedNodes             []string
	Version               uint32
	Guid                  []byte
	ReceiveAddress        string
	HeartbeatTaskSpec     string
	RequestHeightTaskSpec string

	// ---------------------------------- DHT Config ------------------------------------
	// the bucket expired duration
	BucketExpiredAfter time.Duration
	// the node expired duration
	NodeExpriedAfter time.Duration
	// how long it checks whether the bucket is expired
	CheckBucketPeriod time.Duration
	// the max transaction id
	MaxTransactionCursor uint64
	// how many nodes routing table can hold
	MaxNodes int
	// in mainline dht, k = 8
	K int
	// for crawling mode, we put all nodes in one bucket, so BucketSize may
	// not be K
	BucketSize int
	// the nodes num to be fresh in a bucket
	RefreshNodeCount int
	// udp, udp4, udp6
	Network string
	// local network address
	LocalAddr *net.UDPAddr
	// -----------------------------------------------------------------------------------

	// -------------------------------BlockChain Config ----------------------------------

	CheckPeerPeriod     time.Duration
	// -----------------------------------------------------------------------------------
}{
	HeartbeatTaskSpec:     "*/10 * * * * *",
	RequestHeightTaskSpec: "0 * * * * *",
}

func InitConfig(fileGroup string) {
	var err error
	ConfigFileGroup = fileGroup
	Config.DB, err = bolt.Open(GetConfigFilePath(), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(fmt.Sprintf("InitConfig err: %v", err))
	}
	initFirstRun()
}

func GetConfigFilePath() string {
	filePath := ConfigFileName
	if ConfigFileGroup != "" {
		filePath += "." + ConfigFileGroup
	}
	filePath += ".db"
	return filePath
}

func GetBlockFilePath() string {
	filePath := BlockFileName
	if ConfigFileGroup != "" {
		filePath += "." + ConfigFileGroup
	}
	filePath += ".db"
	return filePath
}

func GetWalletFilePath() string {
	filePath := WalletFileName
	if ConfigFileGroup != "" {
		filePath += "." + ConfigFileGroup
	}
	filePath += ".dat"
	return filePath
}

func initFirstRun() {
	err := Config.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(ConfigBucketIdentity))
		if err != nil {
			return err
		}
		checkFirstRunGuidConfig(bucket)
		checkFirstRunVersionConfig(bucket)
		checkFirstRunSeedConfig(bucket)
		checkFirstRunReceiveAddress(bucket)
		checkFirstRunDHTConfig(bucket)
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("initFirstRun err: %v", err))
	}
}

func checkFirstRunReceiveAddress(bucket *bolt.Bucket) {
	value := bucket.Get([]byte(ConfigReceiveAddressKey))
	if value != nil {
		Config.ReceiveAddress = string(value)
		return
	}

	coinBaseReceiveAddress := viper.GetString("CONFIG_COINBASE_RECEIVING_ADDRESS")
	if coinBaseReceiveAddress == "" {
		logrus.Panic("must set a coinbase receiving address")
	}
	bucket.Put([]byte(ConfigReceiveAddressKey), []byte(coinBaseReceiveAddress))
}

func checkFirstRunSeedConfig(bucket *bolt.Bucket) {
	seedAddrs := make([]string, 0)
	seeds := viper.GetStringSlice("CONFIG_SEED_NODES")
	if len(seeds) > 0 {
		seedAddrs = append(seedAddrs, seeds...)
	} else {
		value := bucket.Get([]byte(ConfigSeedKey))
		if value != nil {

			decoder := gob.NewDecoder(bytes.NewReader(value))
			err := decoder.Decode(&seedAddrs)
			if err != nil {
				logrus.Fatal(err)
			}
			Config.SeedNodes = seedAddrs
			return
		}
	}
	Config.SeedNodes = seedAddrs

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(Config.SeedNodes)
	if err != nil {
		logrus.Fatal(err)
	}
	bucket.Put([]byte(ConfigSeedKey), buffer.Bytes())
}

func checkFirstRunGuidConfig(bucket *bolt.Bucket) {
	value := bucket.Get([]byte(ConfigGuidKey))
	if value != nil {
		Config.Guid = value
		return
	}

	Config.Guid = []byte(util.RandomString(20))
	bucket.Put([]byte(ConfigGuidKey), []byte(Config.Guid))
}

func checkFirstRunVersionConfig(bucket *bolt.Bucket) {
	version := viper.GetInt("CONFIG_VERSION")
	if version != 0 {
		Config.Version = uint32(version)
	} else {
		value := bucket.Get([]byte(ConfigVersionKey))
		if value != nil {
			Config.Version = binary.BigEndian.Uint32(value)
			return
		}
		Config.Version = DefaultVersion
	}
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, Config.Version)
	bucket.Put([]byte(ConfigVersionKey), buffer.Bytes())
}

func checkFirstRunDHTConfig(bucket *bolt.Bucket) {
	nodeExpriedAfter := viper.GetString("DHT_NODE_EXPIRED_AFTER")
	if nodeExpriedAfter != "" {
		d, err := time.ParseDuration(nodeExpriedAfter)
		if err != nil {
			logrus.Panicf("DHT_NODE_EXPIRED_AFTER parse error: %v", err)
		}
		Config.NodeExpriedAfter = d
	} else {
		Config.NodeExpriedAfter = 0
	}

	bucketExpiredAfter := viper.GetString("DHT_BUCKET_EXPIRED_AFTER")
	if bucketExpiredAfter != "" {
		d, err := time.ParseDuration(bucketExpiredAfter)
		if err != nil {
			logrus.Panicf("DHT_BUCKET_EXPIRED_AFTER parse error: %v", err)
		}
		Config.BucketExpiredAfter = d
	} else {
		Config.BucketExpiredAfter = 0
	}

	checkBucketPeriod := viper.GetString("DHT_CHECK_BUCKET_PERIOD")
	if checkBucketPeriod != "" {
		d, err := time.ParseDuration(checkBucketPeriod)
		if err != nil {
			logrus.Panicf("DHT_CHECK_BUCKET_PERIOD parse error: %v", err)
		}
		Config.CheckBucketPeriod = d
	} else {
		Config.CheckBucketPeriod = 0
	}

	checkPeerPeriod := viper.GetString("PEER_CHECK_PEER_PERIOD")
	if checkPeerPeriod != "" {
		d, err := time.ParseDuration(checkPeerPeriod)
		if err != nil {
			logrus.Panicf("PEER_CHECK_PEER_PERIOD parse error: %v", err)
		}
		Config.CheckPeerPeriod = d
	} else {
		Config.CheckPeerPeriod = 0
	}

	maxTransactionCursor := viper.GetInt64("DHT_MAX_TRANSACTION_CURSOR")
	if maxTransactionCursor != 0 {
		Config.MaxTransactionCursor = uint64(maxTransactionCursor)
	} else {
		Config.MaxTransactionCursor = math.MaxUint32
	}

	maxNode := viper.GetInt("DHT_MAX_NODE")
	if maxNode != 0 {
		Config.MaxNodes = maxNode
	} else {
		Config.MaxNodes = 5000
	}

	k := viper.GetInt("DHT_K")
	if k != 0 {
		Config.K = k
	} else {
		Config.K = 8
	}

	bucketSize := viper.GetInt("DHT_BUCKET_SIZE")
	if bucketSize != 0 {
		Config.BucketSize = bucketSize
	} else {
		Config.BucketSize = math.MaxInt32
	}

	refreshNodeCount := viper.GetInt("DHT_REFRESH_NODE_COUNT")
	if refreshNodeCount != 0 {
		Config.RefreshNodeCount = refreshNodeCount
	} else {
		Config.RefreshNodeCount = Config.K
	}

	network := viper.GetString("DHT_NETWORK")
	if network != "" {
		Config.Network = network
	} else {
		Config.Network = "udp"
	}

	localAddr := viper.GetString("DHT_LOCAL_ADDR")
	if localAddr == "" {
		localAddr = ":34212"
	}
	addr, err := net.ResolveUDPAddr(Config.Network, localAddr)
	if err != nil {
		logrus.Panicf("DHT_LOCAL_ADDR ResolveUDPAddr error: %v", err)
	}
	Config.LocalAddr = addr
}
