package server

import (
	"bitcoin_nft_v2/db"
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/nft_tree/common"
	"bitcoin_nft_v2/utils"
	"bytes"
	"context"
	"database/sql"
	"fmt"
)

const (
	DefaultNameSpace = "default"
)

func (sv *Server) ImportProof(ctx context.Context, id, url, memo string) error {
	// import nft data and merge tree
	dataByte, key := sv.ComputeNftDataByte(&NftData{
		ID:   id,
		Url:  url,
		Memo: memo,
	})

	// Init Root Hash For Receiver
	leaf := nft_tree.NewLeafNode(dataByte, 0) // CoinsToSend
	leaf.NodeHash()

	txCreator := func(tx *sql.Tx) db.TreeStore {
		return sv.PostgresDB.WithTx(tx)
	}

	treeDB := db.NewTransactionExecutor[db.TreeStore](sv.PostgresDB, txCreator)

	taroTreeStore := db.NewTaroTreeStore(treeDB, DefaultNameSpace)

	tree := nft_tree.NewFullTree(taroTreeStore)

	//We use the default, in-memory store that doesn't actually use the
	//context.
	updatedTree, err := tree.Insert(context.Background(), key, leaf)
	if err != nil {
		fmt.Println(err)
		return err
	}

	updatedRoot, err := updatedTree.Root(context.Background())
	if err != nil {
		fmt.Println(err)
		return err
	}

	rootHash := utils.GetNftRoot(updatedRoot)
	EmbeddedData = rootHash
	err = sv.InsertNewNftData(ctx, id, url, memo)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return err
}

func (sv *Server) ExportProof(ctx context.Context, url string) (*NftData, error) {
	// get nft by url
	if url == "" {
		return nil, utils.WrapperError("[ExportProof] _NFT_URL_REQUIRED_")
	}

	nftDatas, err := sv.GetNftDataByUrl(ctx, []string{url})
	if err != nil {
		fmt.Println("[ExportProof] Get nft data error ", err)
		fmt.Println(err)
		return nil, err
	}

	nftData := nftDatas[0]
	// export data and delete
	nftDataRes := &NftData{
		ID:   nftData.ID,
		Url:  nftData.Url,
		Memo: nftData.Memo,
	}

	// delete data
	err = sv.DeleteNftData(ctx, url)
	if err != nil {
		fmt.Println("[ExportProof] Delete nft data error ", err)
		return nil, err
	}

	return nftDataRes, nil
}

// verify a tree one leaf
func (sv *Server) VerifyDataInTree(ctx context.Context, rootHash []byte, nftData NftData) (bool, error) {
	dataByte, key := sv.ComputeNftDataByte(&nftData)

	// Init Root Hash For Receiver
	leaf := nft_tree.NewLeafNode(dataByte, 0) // CoinsToSend
	leaf.NodeHash()

	// check name space or merge all namespace into one?
	updatedTree, err := common.LoadTreeIntoMemoryByNameSpace(ctx, sv.PostgresDB, "default")
	if err != nil {
		return false, err
	}

	updated2Tree, err := updatedTree.Insert(context.TODO(), key, leaf)
	if err != nil {
		return false, err
	}

	updatedRoot, err := updated2Tree.Root(context.Background())
	if err != nil {
		return false, err
	}

	return bytes.Equal(utils.GetNftRoot(updatedRoot), rootHash), nil
}
