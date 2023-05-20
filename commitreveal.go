package main

import (
	"bitcoin_nft_v2/nft_data"
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/utils"
	"context"
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

	tree := nft_tree.NewCompactedTree(nft_tree.NewDefaultStore())
	sampleDataByte, key := nft_data.GetSampleDataByte()
	leaf := nft_tree.NewLeafNode(sampleDataByte, 0) // CoinsToSend

	// We use the default, in-memory store that doesn't actually use the
	// context.
	updatedTree, err := tree.Insert(context.Background(), key, leaf)

	updatedRoot, err := updatedTree.Root(context.Background())
	if err != nil {
		fmt.Println(err)
		// maybe panic
		return
	}

	rootHash := utils.GetNftRoot(updatedRoot)
	EmbeddedData = rootHash

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
