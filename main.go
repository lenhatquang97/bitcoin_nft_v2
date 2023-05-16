package main

import (
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg"
)

func main() {
	client, err := utils.GetBitcoinWalletRpcClient("btcwallet", TestNetConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.WalletPassphrase(PassphraseInWallet, PassphraseTimeout)
	if err != nil {
		fmt.Println(err)
		return
	}

	commitTx, wif, err := CreateCommitTx(CoinsToSend, client, EmbeddedData, &TestNetConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	commitTxHash, err := client.SendRawTransaction(commitTx, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	retrievedCommitTx, err := client.GetRawTransaction(commitTxHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("===================================Checkpoint 1====================================")

	revealTx, _, err := RevealTx(EmbeddedData, *commitTxHash, *retrievedCommitTx.MsgTx().TxOut[0], 0, wif.PrivKey, &chaincfg.SimNetParams)
	if err != nil {
		fmt.Println(err)
		return
	}

	revealTxHash, err := client.SendRawTransaction(revealTx, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(revealTxHash)

	fmt.Println("===================================Checkpoint 2====================================")

	retrievedTx, err := client.GetRawTransaction(revealTxHash)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range retrievedTx.MsgTx().TxIn {
		for _, witnessItem := range item.Witness {
			fmt.Println(witnessItem)
		}
	}

	fmt.Println("===================================Success====================================")

}
