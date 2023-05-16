package main

import (
	"fmt"
)

const (
	senderAddress = "SeZdpbs8WBuPHMZETPWajMeXZt1xzCJNAJ"
)

func main() {
	embeddedData := []byte("Hello World")
	client, err := GetBitcoinWalletRpcClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.WalletPassphrase("12345", 5)
	if err != nil {
		fmt.Println(err)
		return
	}

	tx, wif, err := CreateFundTx(10000, client, embeddedData)
	if err != nil {
		fmt.Println(err)
		return
	}

	commitTxHash, err := client.SendRawTransaction(tx, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	commitTx, err := client.GetRawTransaction(commitTxHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("===================================Checkpoint 1====================================")

	finalTx, _, err := RevealTx(embeddedData, *commitTxHash, *commitTx.MsgTx().TxOut[0], 0, wif.PrivKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	finalHash, err := client.SendRawTransaction(finalTx, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(finalHash)

	fmt.Println("===================================Checkpoint 2====================================")
}
