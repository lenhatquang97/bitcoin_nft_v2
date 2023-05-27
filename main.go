package main

import (
	"bitcoin_nft_v2/server"
)

func main() {
	var GlobalNetCfg = server.TestNetConfig
	sv, err := server.InitServer()
	if err != nil {
		panic(err)
	}

	sv.DoCommitRevealTransaction(&GlobalNetCfg)
}
