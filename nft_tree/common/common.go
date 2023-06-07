package common

import (
	"bitcoin_nft_v2/db"
	"bitcoin_nft_v2/nft_tree"
	"context"
)

func LoadTreeIntoMemoryByNameSpace(ctx context.Context, postgresDB *db.PostgresStore) (*nft_tree.CompactedTree, error) {
	res, err := postgresDB.GetAllNodeByNameSpace(ctx)
	if err != nil {
		return nil, err
	}

	defaultStore, err := nft_tree.NewStoreWithDB(res)
	if err != nil {
		return nil, err
	}
	tree := nft_tree.NewCompactedTree(defaultStore)

	return tree, nil
}
