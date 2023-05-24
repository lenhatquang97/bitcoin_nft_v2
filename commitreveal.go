package main

import (
	"bitcoin_nft_v2/config"
	db2 "bitcoin_nft_v2/db"
	"bitcoin_nft_v2/nft_data"
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/offchainnft"
	"bitcoin_nft_v2/utils"
	"context"
	"database/sql"
	"fmt"
)

func DoCommitRevealTransaction(netConfig *config.NetworkConfig) {
	client, err := utils.GetBitcoinWalletRpcClient("btcwallet", netConfig)
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

	//db, err := sqlc.NewDBByConn(sqlc.GetDBConnectionString())
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//postgresDB := sqlc.New(db)

	sqlFixture := db2.NewTestPgFixture()
	store, err := db2.NewPostgresStore(sqlFixture.GetConfig())
	if err != nil {
		fmt.Println(err)
		return
	}

	txCreator := func(tx *sql.Tx) db2.TreeStore {
		return store.WithTx(tx)
	}

	treeDB := db2.NewTransactionExecutor[db2.TreeStore](store, txCreator)

	taroTreeStore := db2.NewTaroTreeStore(treeDB, "")

	tree := nft_tree.NewFullTree(taroTreeStore)
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

	customData, err := offchainnft.FileSha256("./README.md")
	if err != nil {
		fmt.Println(err)
		return
	}

	commitTxHash, wif, err := ExecuteCommitTransaction(client, []byte(customData), netConfig)
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

	revealTxInput := RevealTxInput{
		CommitTxHash: commitTxHash,
		Idx:          0,
		Wif:          wif,
		CommitOutput: retrievedCommitTx.MsgTx().TxOut[0],
		ChainConfig:  netConfig.ParamsObject,
	}

	revealTxHash, err := ExecuteRevealTransaction(client, &revealTxInput, []byte(customData))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("===================================Checkpoint 2====================================")
	fmt.Printf("Your reveal tx hash is: %s\n", revealTxHash.String())
	fmt.Println("===================================Success====================================")

}
