package main

import (
	"bitcoin_nft_v2/config"

	"github.com/btcsuite/btcd/chaincfg"
)

const (
	PassphraseInWallet = "12345"
	PassphraseTimeout  = 3
	CoinsToSend        = 15000
	//100k means sending file, so I need to create a mechanism of sending file
	DefaultFee    = 3000
	TESTNET_1_BTC = 100000000
)

var EmbeddedData = []byte("Hello World")

var SimNetConfig = config.NetworkConfig{
	Host:          "localhost:18554",
	Endpoint:      "ws",
	User:          "youruser",
	Pass:          "SomeDecentp4ssw0rd",
	Params:        "simnet",
	ParamsObject:  &chaincfg.SimNetParams,
	SenderAddress: "SeZdpbs8WBuPHMZETPWajMeXZt1xzCJNAJ",
}

var TestNetConfig = config.NetworkConfig{
	Host:         "localhost:18332",
	Endpoint:     "ws",
	User:         "DeW+bgKg011pJHZnaBvgv/lMRks=",
	Pass:         "wD9aohGo2f5LwVg7fdj1ntHQcfY=",
	Params:       "testnet3",
	ParamsObject: &chaincfg.TestNet3Params,
	//Note: in testnet, address is not reused so you need to use default address
	//Another note: Default address has changed everytime you init the server => In UI, you need a mechanism to
	//choose address anyway.
	SenderAddress: "mntb2RxQhyXqXRZV5GE1bDkP6615EPXLHF",
}
