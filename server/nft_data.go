package server

import (
	"bitcoin_nft_v2/db/sqlc"
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/nft_tree/common"
	"bitcoin_nft_v2/utils"
	"context"
	"crypto/sha256"
	"encoding/json"
)

type NftData struct {
	ID   string
	Url  string
	Memo string
}

// InsertNewNftData need to verify that nft is own by user
func (sv *Server) InsertNewNftData(ctx context.Context, id string, url string, memo string) error {
	if url == "" {
		return utils.WrapperError("_NFT_URL_REQUIRED_")
	}

	if memo == "" {
		return utils.WrapperError("_NFT_MEMO_REQUIRED_")
	}
	err := sv.PostgresDB.InsertNftData(ctx, sqlc.InsertNftDataParams{
		ID:   id,
		Url:  url,
		Memo: memo,
	})

	if err != nil {
		return err
	}

	return nil
}

// DeleteNftData nneed to verify
func (sv *Server) DeleteNftData(ctx context.Context, url string) error {
	err := sv.PostgresDB.DeleteNftDataByUrl(ctx, url)
	if err != nil {
		return err
	}

	return nil
}

// GetNftDataByUrl need to unit test
func (sv *Server) GetNftDataByUrl(ctx context.Context, url string) (*NftData, error) {
	nftData, err := sv.PostgresDB.GetNFtDataByUrl(ctx, url)
	if err != nil {
		return nil, err
	}

	return &NftData{
		ID:   nftData.ID,
		Url:  nftData.Url,
		Memo: nftData.Memo,
	}, nil
}

// ComputeNftDataByte return data byte, key data
func (sv *Server) ComputeNftDataByte(data *NftData) ([]byte, [32]byte) {
	h := sha256.New()
	_, _ = h.Write([]byte(data.ID))
	_, _ = h.Write([]byte(data.Url))
	_, _ = h.Write([]byte(data.Memo))

	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, [32]byte{}
	}

	return rawData, *(*[32]byte)(h.Sum(nil))
}

func (sv *Server) NewRootHashForReceiver(key [32]byte, leaf *nft_tree.LeafNode) ([]byte, error) {
	tree := nft_tree.NewCompactedTree(nft_tree.NewDefaultStore())

	updatedTree, err := tree.Insert(context.TODO(), key, leaf)
	if err != nil {
		return nil, err
	}

	updatedRoot, err := updatedTree.Root(context.Background())
	if err != nil {
		return nil, err
	}

	return utils.GetNftRoot(updatedRoot), nil
}

func (sv *Server) PreComputeRootHashForSender(ctx context.Context, key [32]byte, leaf *nft_tree.LeafNode, nameSpace string) ([]byte, error) {
	updatedTree, err := common.LoadTreeIntoMemoryByNameSpace(ctx, sv.PostgresDB, nameSpace)
	if err != nil {
		return nil, err
	}

	updated2Tree, err := updatedTree.Insert(context.TODO(), key, leaf)
	if err != nil {
		return nil, err
	}

	updatedRoot, err := updated2Tree.Root(context.Background())
	if err != nil {
		return nil, err
	}

	return utils.GetNftRoot(updatedRoot), nil
}
