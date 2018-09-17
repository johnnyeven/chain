package global

const (
	ConfigFileName           = "config"
	BlockFileName            = "block"
	WalletFileName           = "wallet"
	ConfigBucketIdentity     = "self"
	BlockBucketIdentity      = "block"
	ChainStateBucketIdentity = "chain_state"
	ChainLastIndexKey        = "last"
	ConfigGuidKey            = "guid"
	ConfigVersionKey         = "version"
	ConfigPortKey            = "port"
	ConfigSeedKey            = "seeds"
	ConfigReceiveAddressKey  = "receive_address"

	DefaultVersion  uint32 = 32
	DefaultPort     uint32 = 34212
	DefaultSeedNode        = "127.0.0.1:34212"
)
