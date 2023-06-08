package main

import (
	"bitcoin_nft_v2/handler"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

const (
	NUM_OF_PARAM     = 7
	CHAIN_MODE       = 1
	NETWORK          = 2
	HOST             = 3
	USER             = 4
	PASSWORD         = 5
	SENDER_ADDRESS   = 6
	DEFAULT_PORT_API = 3000
	// input info: mode, network, host, user, pass, sender-address, port-api
)

func main() {
	//var GlobalNetCfg = server.TestNetConfig
	//sv, err := server.InitServer()
	//if err != nil {
	//	panic(err)
	//}
	//
	//sv.DoCommitRevealTransaction(&GlobalNetCfg)
	if len(os.Args) != NUM_OF_PARAM {
		log.Fatal("Num of args is invalid")
	}

	handler.Run(&handler.Config{
		Mode:          os.Args[CHAIN_MODE],
		Network:       os.Args[NETWORK],
		Host:          os.Args[HOST],
		User:          os.Args[USER],
		Password:      os.Args[PASSWORD],
		SenderAddress: os.Args[SENDER_ADDRESS],
	})
}

func SampleFile() {
	file, err := os.ReadFile("./sample.jpg")
	if err != nil {
		fmt.Println("Error 1")
		return
	}
	//Convert bytes to string
	str := hex.EncodeToString(file)
	fmt.Println("Done")
	fmt.Println(str)
}
