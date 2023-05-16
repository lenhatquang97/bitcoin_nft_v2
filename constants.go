package main

import (
	"bitcoin_nft_v2/config"

	"github.com/btcsuite/btcd/chaincfg"
)

const (
	PassphraseInWallet = "12345"
	PassphraseTimeout  = 3
	CoinsToSend        = 10000
	TESTNET_1_BTC      = 100000000
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
	Host:          "localhost:18332",
	Endpoint:      "ws",
	User:          "4bmeiF7E3ny8cGf8Ok6QJZy/0pk=",
	Pass:          "2oljjSoRFzC5Go7hCGDID6xWi+c=",
	Params:        "testnet3",
	ParamsObject:  &chaincfg.TestNet3Params,
	SenderAddress: "mzcH9PSuCUaB4JShJNUGEqJrtAGV1wNQiB",
}
