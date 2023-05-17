package main

import (
	"bitcoin_nft_v2/utils"
	"fmt"
)

func DoCommitRevealTransaction() {
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
	fmt.Println("===================================Checkpoint 0====================================")

	commitTxHash, wif, err := ExecuteCommitTransaction(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Your commit tx hash is: %s\n", commitTxHash.String())

	retrievedCommitTx, err := client.GetRawTransaction(commitTxHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("===================================Checkpoint 1====================================")

	revealTxHash, err := ExecuteRevealTransaction(client, commitTxHash, 0, wif, retrievedCommitTx.MsgTx().TxOut[0], TestNetConfig.ParamsObject)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("===================================Checkpoint 2====================================")
	fmt.Printf("Your reveal tx hash is: %s\n", revealTxHash.String())
	fmt.Println("===================================Success====================================")

}
