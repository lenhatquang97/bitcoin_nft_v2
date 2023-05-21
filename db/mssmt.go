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
	// NewBranch is a type alias for the params to create a new mssmt
	// branch node.
	NewBranch = sqlc.InsertBranchParams

	// NewLeaf is a type alias for the params to create a new mssmt leaf
	// node.
	NewLeaf = sqlc.InsertLeafParams

	// NewCompactedLeaf is a type alias for the params to create a new
	// mssmt compacted leaf node.
	NewCompactedLeaf = sqlc.InsertCompactedLeafParams

	// StoredNode is a type alias for an arbitrary child of an mssmt branch.
	StoredNode = sqlc.FetchChildrenRow

	// DelNode wraps the args we need to delete a node.
	DelNode = sqlc.DeleteNodeParams

	// ChildQuery wraps the args we need to fetch the children of a node.
	ChildQuery = sqlc.FetchChildrenParams

	// UpdateRoot wraps the args we need to update a root node.
	UpdateRoot = sqlc.UpsertRootNodeParams
)

// TreeStore is a sub-set of the main sqlc.Querier interface that contains
// only the methods needed to manipulate and query stored MSSMT trees.
type TreeStore interface {
	// InsertBranch inserts a new branch to the store.
	InsertBranch(ctx context.Context, newNode NewBranch) error

	// InsertLeaf inserts a new leaf to the store.
	InsertLeaf(ctx context.Context, newNode NewLeaf) error

	// InsertCompactedLeaf inserts a new compacted leaf to the store.
	InsertCompactedLeaf(ctx context.Context, newNode NewCompactedLeaf) error

	// FetchChildren fetches the children (at most two currently) of the
	// passed branch hash key.
	FetchChildren(ctx context.Context, c ChildQuery) ([]StoredNode, error)

	// DeleteNode deletes a node (can be either branch, leaf of compacted
	// leaf) from the store.
	DeleteNode(ctx context.Context, n DelNode) (int64, error)

	// FetchRootNode fetches the root node for the specified namespace.
	FetchRootNode(ctx context.Context,
		namespace string) (sqlc.MssmtNode, error)

	// UpsertRootNode allows us to update the root node in place for a
	// given namespace.
	UpsertRootNode(ctx context.Context, arg UpdateRoot) error
}

type TreeStoreTxOptions struct {
	// readOnly governs if a read only transaction is needed or not.
	readOnly bool
}

// ReadOnly returns true if the transaction should be read only.
//
// NOTE: This implements the TxOptions
func (t *TreeStoreTxOptions) ReadOnly() bool {
	return t.readOnly
}

// NewTreeStoreReadTx creates a new read transaction option set.
func NewTreeStoreReadTx() TreeStoreTxOptions {
	return TreeStoreTxOptions{
		readOnly: true,
	}
}

// BatchedTreeStore is a version of the AddrBook that's capable of batched
// database operations.
type BatchedTreeStore interface {
	TreeStore

	BatchedTx[TreeStore]
}

// TaroTreeStore is an persistent MS-SMT implementation backed by a live SQL
// database.
type TaroTreeStore struct {
	db        BatchedTreeStore
	namespace string
}

// NewTaroTreeStore creates a new TaroTreeStore instance given an open
// BatchedTreeStore storage backend. The namespace argument is required, as it
// allow us to store several distinct trees on disk in the same table.
func NewTaroTreeStore(db BatchedTreeStore, namespace string) *TaroTreeStore {
	return &TaroTreeStore{
		db:        db,
		namespace: namespace,
	}
}

var _ nft_tree.TreeStore = (*TaroTreeStore)(nil)

// Update updates the persistent tree in the passed update closure using the
// update transaction.
func (t *TaroTreeStore) Update(ctx context.Context,
	update func(tx nft_tree.TreeStoreUpdateTx) error) error {

	txBody := func(dbTx TreeStore) error {
		updateTx := &taroTreeStoreTx{
			ctx:       ctx,
			dbTx:      dbTx,
			namespace: t.namespace,
		}

		return update(updateTx)
	}

	var writeTxOpts TreeStoreTxOptions
	return t.db.ExecTx(ctx, &writeTxOpts, txBody)
}

// View gives a view of the persistent tree in the passed view closure using
// the view transaction.
func (t *TaroTreeStore) View(ctx context.Context,
	update func(tx nft_tree.TreeStoreViewTx) error) error {

	txBody := func(dbTx TreeStore) error {
		viewTx := &taroTreeStoreTx{
			ctx:       ctx,
			dbTx:      dbTx,
			namespace: t.namespace,
		}

		return update(viewTx)
	}

	readTxOpts := TreeStoreTxOptions{
		readOnly: true,
	}

	return t.db.ExecTx(ctx, &readTxOpts, txBody)
}

type taroTreeStoreTx struct {
	ctx       context.Context
	dbTx      TreeStore
	namespace string
}

// InsertBranch stores a new branch keyed by its NodeHash.
func (t *taroTreeStoreTx) InsertBranch(branch *nft_tree.BranchNode) error {
	hashKey := branch.NodeHash()
	lHashKey := branch.Left.NodeHash()
	rHashKey := branch.Right.NodeHash()

	if err := t.dbTx.InsertBranch(t.ctx, NewBranch{
		HashKey:   hashKey[:],
		LHashKey:  lHashKey[:],
		RHashKey:  rHashKey[:],
		Sum:       int64(branch.NodeSum()),
		Namespace: t.namespace,
	}); err != nil {
		return fmt.Errorf("unable to insert branch: %w", err)
	}

	return nil
}

// InsertLeaf stores a new leaf keyed by its NodeHash (not the insertion key).
func (t *taroTreeStoreTx) InsertLeaf(leaf *nft_tree.LeafNode) error {
	hashKey := leaf.NodeHash()

	if err := t.dbTx.InsertLeaf(t.ctx, NewLeaf{
		HashKey:   hashKey[:],
		Value:     leaf.Value,
		Sum:       int64(leaf.NodeSum()),
		Namespace: t.namespace,
	}); err != nil {
		return fmt.Errorf("unable to insert leaf: %w", err)
	}

	return nil
}

// InsertCompactedLeaf stores a new compacted leaf keyed by its
// NodeHash (not the insertion key).
func (t *taroTreeStoreTx) InsertCompactedLeaf(
	leaf *nft_tree.CompactedLeafNode) error {

	hashKey := leaf.NodeHash()
	key := leaf.Key()

	if err := t.dbTx.InsertCompactedLeaf(t.ctx, NewCompactedLeaf{
		HashKey:   hashKey[:],
		Key:       key[:],
		Value:     leaf.Value,
		Sum:       int64(leaf.NodeSum()),
		Namespace: t.namespace,
	}); err != nil {
		return fmt.Errorf("unable to insert compacted leaf: %w", err)
	}

	return nil
}

// DeleteBranch deletes the branch node keyed by the given NodeHash.
func (t *taroTreeStoreTx) DeleteBranch(hashKey nft_tree.NodeHash) error {
	_, err := t.dbTx.DeleteNode(t.ctx, DelNode{
		HashKey:   hashKey[:],
		Namespace: t.namespace,
	})
	return err
}

// DeleteLeaf deletes the leaf node keyed by the given NodeHash.
func (t *taroTreeStoreTx) DeleteLeaf(hashKey nft_tree.NodeHash) error {
	_, err := t.dbTx.DeleteNode(t.ctx, DelNode{
		HashKey:   hashKey[:],
		Namespace: t.namespace,
	})
	return err
}

// DeleteCompactedLeaf deletes a compacted leaf keyed by the given NodeHash.
func (t *taroTreeStoreTx) DeleteCompactedLeaf(hashKey nft_tree.NodeHash) error {
	_, err := t.dbTx.DeleteNode(t.ctx, DelNode{
		HashKey:   hashKey[:],
		Namespace: t.namespace,
	})
	return err
}

// newKey is a helper to convert a byte slice of the correct size to a 32 byte
// array.
func newKey(data []byte) ([32]byte, error) {
	var key [32]byte

	if len(data) != 32 {
		return key, fmt.Errorf("invalid key size")
	}

	copy(key[:], data)
	return key, nil
}

// GetChildren returns the left and right child of the node keyed by the given
// NodeHash.
func (t *taroTreeStoreTx) GetChildren(height int, hashKey nft_tree.NodeHash) (
	nft_tree.Node, nft_tree.Node, error) {

	dbRows, err := t.dbTx.FetchChildren(t.ctx, ChildQuery{
		HashKey:   hashKey[:],
		Namespace: t.namespace,
	})
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
			leaf := nft_tree.NewLeafNode(
				row.Value, uint64(row.Sum),
			)

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

			node = nft_tree.NewComputedBranch(hashKey, uint64(row.Sum))
		}

		if isLeft {
			left = node
		} else {
			right = node
		}
	}

	return left, right, nil
}

// RootNode returns the root nodes of the MS-SMT. If the tree has no elements,
// then a nil node is returned.
func (t *taroTreeStoreTx) RootNode() (nft_tree.Node, error) {
	var root nft_tree.Node

	rootNode, err := t.dbTx.FetchRootNode(t.ctx, t.namespace)
	switch {
	// If there're no rows, then this means it's an empty tree, so we
	// return the root empty node.
	case errors.Is(err, sql.ErrNoRows):
		return nft_tree.EmptyTree[0], nil

	case err != nil:
		return nil, err
	}

	nodeHash, err := newKey(rootNode.HashKey)
	if err != nil {
		return nil, err
	}

	root = nft_tree.NewComputedBranch(nodeHash, uint64(rootNode.Sum))

	return root, nil
}

// UpdateRoot updates the index that points to the root node for the persistent
// tree.
func (t *taroTreeStoreTx) UpdateRoot(rootNode *nft_tree.BranchNode) error {
	rootHash := rootNode.NodeHash()

	// We'll do a sanity check here to ensure that we're not trying to
	// insert a root hash. This might happen when we delete all the items
	// in a tree.
	//
	// If we try to insert this, then the foreign key constraint will fail,
	// as empty hashes are never stored (root would point to a node not in
	// the DB).
	if rootHash == nft_tree.EmptyTree[0].NodeHash() {
		return nil
	}

	return t.dbTx.UpsertRootNode(t.ctx, UpdateRoot{
		RootHash:  rootHash[:],
		Namespace: t.namespace,
	})
}
