package main

import (
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/server"
	"bitcoin_nft_v2/utils"
	"context"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
)

const (
	namespace = "default"
)

// func test to export nft
func ImportNewNftData(url string, memo string) {
	sv, err := server.InitServer()
	if err != nil {
		fmt.Println("[ImportNewNftData] init server error", err)
		return
	}

	id := uuid.New().String()
	err = sv.InsertNewNftData(context.Background(), id, url, memo)
	if err != nil {
		fmt.Println("[ImportNewNftData] insert nft data error ", err)
		return
	}

	fmt.Println("[ImportNewNftData] import nft data success")
}

func DeleteNftData(url string) {
	sv, err := server.InitServer()
	if err != nil {
		fmt.Println("[DeleteNftData] init server error", err)
		return
	}

	err = sv.DeleteNftData(context.Background(), url)
	if err != nil {
		fmt.Println("[DeleteNftData] delete nft data error ", err)
		return
	}

	fmt.Println("[DeleteNftData] delete nft data success")
}

func ImportNftData(id, url, memo string) {
	sv, err := server.InitServer()
	if err != nil {
		fmt.Println("[ImportNftData] init server error", err)
		return
	}

	err = sv.ImportProof(context.Background(), id, url, memo)
	if err != nil {
		fmt.Println("[ImportNftData] import proof failed", err)
		fmt.Println(err)
		return
	}

	fmt.Println("[ImportNftData] success")
}

func ExportNftData(url string) {
	sv, err := server.InitServer()
	if err != nil {
		fmt.Println("[ExportNftData] init server error", err)
		return
	}

	data, err := sv.ExportProof(context.Background(), url)
	if err != nil {

		fmt.Println(err)
		return
	}

	fmt.Println("[ExportNftData] data export is ", data)
	fmt.Println("[ExportNftData] success")
}

func VerifyEqual(id, url, memo string) {
	sv, err := server.InitServer()
	if err != nil {
		fmt.Println("[ExportNftData] init server error", err)
		return
	}

	nftData := server.NftData{
		ID:   id,
		Url:  url,
		Memo: memo,
	}

	dataByte, key := sv.ComputeNftDataByte(&nftData)

	// Init Root Hash For Receiver
	leaf := nft_tree.NewLeafNode(dataByte, 0) // CoinsToSend
	leaf.NodeHash()

	tree := nft_tree.NewCompactedTree(nft_tree.NewDefaultStore())

	updatedTree, err := tree.Insert(context.TODO(), key, leaf)
	if err != nil {
		fmt.Println("[VerifyEqual] update tree error ", err)
		return
	}

	updatedRoot, err := updatedTree.Root(context.Background())
	if err != nil {
		fmt.Println("[VerifyEqual] update root error ", err)
		return
	}

	rootHash := utils.GetNftRoot(updatedRoot)

	isEqual, err := sv.VerifyDataInTree(context.Background(), rootHash, nftData)
	if err != nil {
		fmt.Println("[VerifyEqual] verify data in tree ", err)
		return
	}

	if !isEqual {
		fmt.Println("[VerifyEqual] data is diff ", rootHash)
		return
	}

	fmt.Println("[VerifyEqual] success")
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// func test to create tx
func main() {
	url := "https://amnhacvietthanh.vn/wp-content/uploads/2020/10/Yamaha-C40.jpg"
	memo := RandStringBytes(5)
	id := uuid.New().String()

	ImportNftData(id, url, memo)
	//DeleteNftData(url)
	//ExportNftData(url)
	//ImportNftData(id, url, memo)
	//VerifyEqual(id, url, memo)
}
