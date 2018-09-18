package services

var presetStack = map[string][]serviceConstructor{
	"blockchain": {
		NewHandshakeService,
		NewDiscoveryService,
		NewBlockChainService,
	},
}
