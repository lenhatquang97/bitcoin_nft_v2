package server

import (
	"bitcoin_nft_v2/config"
	"bitcoin_nft_v2/utils"
	"context"
	"fmt"
)

func (sv *Server) DoCommitRevealTransaction(netConfig *config.NetworkConfig) {
	nftUrls := []string{
		"https://genk.mediacdn.vn/k:thumb_w/640/2016/photo-1-1473821552147/top6suthatcucsocvepikachu.jpg",
		"https://pianofingers.vn/wp-content/uploads/2020/12/organ-casio-ct-s100-1.jpg",
		"https://amnhacvietthanh.vn/wp-content/uploads/2020/10/Yamaha-C40.jpg",
	}
	//nameSpace := DefaultNameSpace

	// Get Nft Data
	nftData, err := sv.GetNftDataByUrl(context.Background(), nftUrls)
	if err != nil {
		print("Get Nft Data Failed")
		fmt.Println(err)
		return
	}

	rootHashForReceiver, err := sv.NewRootHashForReceiver(nftData)
	if err != nil {
		fmt.Println("Compute root hash for receiver error")
		fmt.Println(err)
		return
	}

	// root hash for sender
	//rootHashForSender, err := sv.PreComputeRootHashForSender(context.Background(), key, leaf, nameSpace)
	//if err != nil {
	//	fmt.Println("Compute root hash for sender error")
	//	fmt.Println(err)
	//	return
	//}
	//
	//fmt.Println("Sender root hash update is: ", rootHashForSender)

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

	//txCreator := func(tx *sql.Tx) db.TreeStore {
	//	return sv.PostgresDB.WithTx(tx)
	//}
	//
	//treeDB := db.NewTransactionExecutor[db.TreeStore](sv.PostgresDB, txCreator)
	//
	//taroTreeStore := db.NewTaroTreeStore(treeDB, DefaultNameSpace)
	//
	//tree := nft_tree.NewFullTree(taroTreeStore)
	//
	//_, err = tree.Delete(context.Background(), key)
	//if err != nil {
	//	fmt.Println("Delete leaf after reveal tx failed", err)
	//}
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
	fmt.Println("===================================Checkpoint 2====================================")
	fmt.Printf("Your reveal tx hash is: %s\n", revealTxHash.String())
	fmt.Println("===================================Success====================================")

}
