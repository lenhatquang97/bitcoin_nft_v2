package server

import (
	"bitcoin_nft_v2/config"
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/utils"
	"context"
	"fmt"
)

func (sv *Server) DoCommitRevealTransaction(netConfig *config.NetworkConfig) {
	nftUrl := ""
	nameSpace := "kiet"

	// Get Nft Data
	nftData, err := sv.GetNftDataByUrl(context.Background(), nftUrl)
	if err != nil {
		print("Get Nft Data Failed")
		fmt.Println(err)
		return
	}

	// Compute Nft Data Info
	dataByte, key := sv.ComputeNftDataByte(nftData)

	// Init Root Hash For Receiver
	leaf := nft_tree.NewLeafNode(dataByte, 0) // CoinsToSend
	leaf.NodeHash()

	rootHashForReceiver, err := sv.NewRootHashForReceiver(key, leaf)
	if err != nil {
		fmt.Println("Compute root hash for receiver error")
		fmt.Println(err)
		return
	}

	// root hash for sender
	rootHashForSender, err := sv.PreComputeRootHashForSender(context.Background(), key, leaf, nameSpace)
	if err != nil {
		fmt.Println("Compute root hash for sender error")
		fmt.Println(err)
		return
	}

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

	//txCreator := func(tx *sql.Tx) db2.TreeStore {
	//	return sv.PostgresDB.WithTx(tx)
	//}
	//
	//treeDB := db2.NewTransactionExecutor[db2.TreeStore](sv.PostgresDB, txCreator)
	//
	//taroTreeStore := db2.NewTaroTreeStore(treeDB, "quang4")
	//
	//tree := nft_tree.NewFullTree(taroTreeStore)

	// We use the default, in-memory store that doesn't actually use the
	// context.
	//updatedTree, err := tree.Insert(context.Background(), key, leaf)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//updatedRoot, err := updatedTree.Root(context.Background())
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//rootHash := utils.GetNftRoot(updatedRoot)
	//EmbeddedData = rootHash

	//customData, err := offchainnft.FileSha256("./README.md")
	if err != nil {
		fmt.Println(err)
		return
	}

	commitTxHash, wif, err := ExecuteCommitTransaction(client, rootHashForReceiver, netConfig)
	if err != nil {
		fmt.Println("commitLog")
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

	revealTxHash, err := ExecuteRevealTransaction(client, &revealTxInput, rootHashForReceiver)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("===================================Checkpoint 2====================================")
	fmt.Printf("Your reveal tx hash is: %s\n", revealTxHash.String())
	fmt.Println("===================================Success====================================")

}
