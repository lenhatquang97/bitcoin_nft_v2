package db

import (
	"bitcoin_nft_v2/db/sqlc"
	"bitcoin_nft_v2/nft_tree"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type (
	NewBranch = sqlc.InsertBranchParams

	NewLeaf = sqlc.InsertLeafParams

	NewCompactedLeaf = sqlc.InsertCompactedLeafParams

	StoredNode = sqlc.FetchChildrenRow

	DelNode = []byte

	ChildQuery = []byte

	UpdateRoot = []byte
)

// TreeStore is a sub-set of the main sqlc.Querier interface that contains
// only the methods needed to manipulate and query stored MSSMT trees.
type TreeStore interface {
	InsertBranch(ctx context.Context, newNode NewBranch) error

	InsertLeaf(ctx context.Context, newNode NewLeaf) error

	InsertCompactedLeaf(ctx context.Context, newNode NewCompactedLeaf) error

	FetchChildren(ctx context.Context, c ChildQuery) ([]StoredNode, error)

	DeleteNode(ctx context.Context, n DelNode) (int64, error)

	FetchRootNode(ctx context.Context) (sqlc.NftNode, error)

	UpsertRootNode(ctx context.Context, arg UpdateRoot) error
}

type TreeStoreTxOptions struct {
	// readOnly governs if a read only transaction is needed or not.
	readOnly bool
}

func (t *TreeStoreTxOptions) ReadOnly() bool {
	return t.readOnly
}

type BatchedTreeStore interface {
	TreeStore

	BatchedTx[TreeStore]
}

type TaroTreeStore struct {
	db BatchedTreeStore
}

func NewTaroTreeStore(db BatchedTreeStore) *TaroTreeStore {
	return &TaroTreeStore{
		db: db,
	}
}

var _ nft_tree.TreeStore = (*TaroTreeStore)(nil)

func (t *TaroTreeStore) Update(ctx context.Context,
	update func(tx nft_tree.TreeStoreUpdateTx) error) error {

	txBody := func(dbTx TreeStore) error {
		updateTx := &taroTreeStoreTx{
			ctx:  ctx,
			dbTx: dbTx,
		}

		return update(updateTx)
	}

	var writeTxOpts TreeStoreTxOptions
	return t.db.ExecTx(ctx, &writeTxOpts, txBody)
}

func (t *TaroTreeStore) View(ctx context.Context,
	update func(tx nft_tree.TreeStoreViewTx) error) error {

	txBody := func(dbTx TreeStore) error {
		viewTx := &taroTreeStoreTx{
			ctx:  ctx,
			dbTx: dbTx,
		}

		return update(viewTx)
	}

	readTxOpts := TreeStoreTxOptions{
		readOnly: true,
	}

	return t.db.ExecTx(ctx, &readTxOpts, txBody)
}

type taroTreeStoreTx struct {
	ctx  context.Context
	dbTx TreeStore
}

func (t *taroTreeStoreTx) InsertBranch(branch *nft_tree.BranchNode) error {
	hashKey := branch.NodeHash()
	lHashKey := branch.Left.NodeHash()
	rHashKey := branch.Right.NodeHash()

	if err := t.dbTx.InsertBranch(t.ctx, NewBranch{
		HashKey:  hashKey[:],
		LHashKey: lHashKey[:],
		RHashKey: rHashKey[:],
	}); err != nil {
		return fmt.Errorf("unable to insert branch: %w", err)
	}

	return nil
}

func (t *taroTreeStoreTx) InsertLeaf(leaf *nft_tree.LeafNode) error {
	hashKey := leaf.NodeHash()

	if err := t.dbTx.InsertLeaf(t.ctx, NewLeaf{
		HashKey: hashKey[:],
		Value:   leaf.Value,
	}); err != nil {
		return fmt.Errorf("unable to insert leaf: %w", err)
	}

	return nil
}

func (t *taroTreeStoreTx) InsertCompactedLeaf(
	leaf *nft_tree.CompactedLeafNode) error {

	hashKey := leaf.NodeHash()
	key := leaf.Key()

	if err := t.dbTx.InsertCompactedLeaf(t.ctx, NewCompactedLeaf{
		HashKey: hashKey[:],
		Key:     key[:],
		Value:   leaf.Value,
	}); err != nil {
		return fmt.Errorf("unable to insert compacted leaf: %w", err)
	}

	return nil
}

func (t *taroTreeStoreTx) DeleteBranch(hashKey nft_tree.NodeHash) error {
	_, err := t.dbTx.DeleteNode(t.ctx,
		hashKey[:],
	)
	return err
}

func (t *taroTreeStoreTx) DeleteLeaf(hashKey nft_tree.NodeHash) error {
	_, err := t.dbTx.DeleteNode(t.ctx,
		hashKey[:],
	)
	return err
}

func (t *taroTreeStoreTx) DeleteCompactedLeaf(hashKey nft_tree.NodeHash) error {
	_, err := t.dbTx.DeleteNode(t.ctx,
		hashKey[:],
	)
	return err
}

func newKey(data []byte) ([32]byte, error) {
	var key [32]byte

	if len(data) != 32 {
		return key, fmt.Errorf("invalid key size")
	}

	copy(key[:], data)
	return key, nil
}

func (t *taroTreeStoreTx) GetChildren(height int, hashKey nft_tree.NodeHash) (
	nft_tree.Node, nft_tree.Node, error) {

	dbRows, err := t.dbTx.FetchChildren(t.ctx,
		hashKey[:],
	)
	if err != nil {
		return nil, nil, err
	}

	var (
		left  nft_tree.Node = nft_tree.EmptyTree[height+1]
		right nft_tree.Node = nft_tree.EmptyTree[height+1]
	)

	var lHashKey, rHashKey []byte

	for i, row := range dbRows {
		if i == 0 {
			// The root of the subtree, we're looking for the
			// children, so we skip this node.
			lHashKey = row.LHashKey
			rHashKey = row.RHashKey
			continue
		}

		isLeft := bytes.Equal(row.HashKey, lHashKey)
		isRight := bytes.Equal(row.HashKey, rHashKey)

		if !isLeft && !isRight {
			// Some child node further down the tree.
			continue
		}

		var node nft_tree.Node

		// Since both children are nil, we can assume this is a leaf.
		if row.LHashKey == nil && row.RHashKey == nil {
			leaf := nft_tree.NewLeafNode(row.Value)

			// Precompute the node hash key.
			leaf.NodeHash()

			// We store the key for compacted leafs.
			if row.Key != nil {
				key, err := newKey(row.Key)
				if err != nil {
					return nil, nil, err
				}

				node = nft_tree.NewCompactedLeafNode(
					height+1, &key, leaf,
				)
			} else {
				node = leaf
			}
		} else {
			hashKey, err := newKey(row.HashKey)
			if err != nil {
				return nil, nil, err
			}

			node = nft_tree.NewComputedBranch(hashKey)
		}

		if isLeft {
			left = node
		} else {
			right = node
		}
	}

	return left, right, nil
}

func (t *taroTreeStoreTx) RootNode() (nft_tree.Node, error) {
	var root nft_tree.Node

	rootNode, err := t.dbTx.FetchRootNode(t.ctx)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nft_tree.EmptyTree[0], nil

	case err != nil:
		return nil, err
	}

	nodeHash, err := newKey(rootNode.HashKey)
	if err != nil {
		return nil, err
	}

	root = nft_tree.NewComputedBranch(nodeHash)

	return root, nil
}

func (t *taroTreeStoreTx) UpdateRoot(rootNode *nft_tree.BranchNode) error {
	rootHash := rootNode.NodeHash()

	if rootHash == nft_tree.EmptyTree[0].NodeHash() {
		return nil
	}

	return t.dbTx.UpsertRootNode(t.ctx,
		rootHash[:],
	)
}
