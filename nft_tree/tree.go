package nft_tree

import (
	"context"
)

const (
	// MaxTreeLevels represents the depth of the MS-SMT.
	MaxTreeLevels = hashSize * 8

	// lastBitIndex represents the index of the last bit for MS-SMT keys.
	lastBitIndex = MaxTreeLevels - 1
)

var (
	// EmptyTree stores a copy of all nodes up to the root in a MS-SMT in
	// which all the leaves are empty.
	EmptyTree []Node
)

func init() {
	// Force the calculation of the node key for the empty node. This will
	// ensure the value is fully cached for the loop below.
	EmptyLeafNode.NodeHash()

	// Initialize the empty MS-SMT by starting from an empty leaf and
	// hashing all the way up to the root.
	EmptyTree = make([]Node, MaxTreeLevels+1)
	EmptyTree[MaxTreeLevels] = EmptyLeafNode
	for i := lastBitIndex; i >= 0; i-- {
		// Create the branch and force the calculation of the node key.
		// At this point we already have computed the keys of each of
		// the siblings, so those cached values can be used here. If we
		// don't do this, then concurrent callers will attempt to
		// read/populate this value causing a race condition.
		branch := NewBranch(EmptyTree[i+1], EmptyTree[i+1])
		branch.NodeHash()
		branch.NodeSum()

		EmptyTree[i] = branch
	}
}

// FullTree represents a Merkle-Sum Sparse Merkle Tree (MS-SMT). A MS-SMT is an
// augmented version of a sparse merkle tree that includes a sum value, which is
// combined during the internal branch hashing operation. Such trees permit
// efficient proofs of non-inclusion, while also supporting efficient fault
// proofs of invalid merkle sum commitments.
type FullTree struct {
	store TreeStore
}

var _ Tree = (*FullTree)(nil)

// NewFullTree initializes an empty MS-SMT backed by `store`. As a result,
// `store` will only maintain non-empty relevant nodes, i.e., stale parents are
// deleted and empty nodes are never stored.
func NewFullTree(store TreeStore) *FullTree {
	return &FullTree{
		store: store,
	}
}

// Root returns the root node of the MS-SMT.
func (t *FullTree) Root(ctx context.Context) (*BranchNode, error) {
	var root Node
	err := t.store.View(ctx, func(tx TreeStoreViewTx) error {
		var err error

		root, err = tx.RootNode()
		return err
	})
	if err != nil {
		return nil, err
	}

	return root.(*BranchNode), nil
}

// bitIndex returns the bit found at `idx` for a NodeHash.
func bitIndex(idx uint8, key *[hashSize]byte) byte {
	byteVal := key[idx/8]
	return (byteVal >> (idx % 8)) & 1
}

// iterFunc is a type alias for closures to be invoked at every iteration of
// walking through a tree.
type iterFunc = func(height int, current, sibling, parent Node) error

// walkDown walks down the tree from the root node to the leaf indexed by `key`.
// The leaf node found is returned.
func (t *FullTree) walkDown(tx TreeStoreViewTx, key *[hashSize]byte,
	iter iterFunc) (*LeafNode, error) {

	current, err := tx.RootNode()
	if err != nil {
		return nil, err
	}

	for i := 0; i <= lastBitIndex; i++ {
		left, right, err := tx.GetChildren(i, current.NodeHash())
		if err != nil {
			return nil, err
		}

		var next, sibling Node
		if bitIndex(uint8(i), key) == 0 {
			next, sibling = left, right
		} else {
			next, sibling = right, left
		}
		if iter != nil {
			err := iter(i, next, sibling, current)
			if err != nil {
				return nil, err
			}
		}
		current = next
	}

	return current.(*LeafNode), nil
}

// walkUp walks up from the `start` leaf node up to the root with the help of
// `siblings`. The root branch node computed is returned.
func walkUp(key *[hashSize]byte, start *LeafNode, siblings []Node,
	iter iterFunc) (*BranchNode, error) {

	var current Node = start
	for i := lastBitIndex; i >= 0; i-- {
		sibling := siblings[lastBitIndex-i]
		var parent Node
		if bitIndex(uint8(i), key) == 0 {
			parent = NewBranch(current, sibling)
		} else {
			parent = NewBranch(sibling, current)
		}
		if iter != nil {
			err := iter(i, current, sibling, parent)
			if err != nil {
				return nil, err
			}
		}
		current = parent
	}

	return current.(*BranchNode), nil
}

// insert inserts a leaf node at the given key within the MS-SMT.
func (t *FullTree) insert(tx TreeStoreUpdateTx, key *[hashSize]byte,
	leaf *LeafNode) (*BranchNode, error) {

	// As we walk down to the leaf node, we'll keep track of the sibling
	// and parent for each node we visit.
	prevParents := make([]NodeHash, MaxTreeLevels)
	siblings := make([]Node, MaxTreeLevels)
	_, err := t.walkDown(
		tx, key, func(i int, _, sibling, parent Node) error {
			prevParents[MaxTreeLevels-1-i] = parent.NodeHash()
			siblings[MaxTreeLevels-1-i] = sibling
			return nil
		})
	if err != nil {
		return nil, err
	}

	// Now that we've arrived at the leaf node, we'll need to work our way
	// back up to the root, updating any stale and new intermediate branch
	// nodes.
	root, err := walkUp(
		key, leaf, siblings, func(i int, _, _, parent Node) error {
			// Replace the old parent with the new one. Our store
			// should never track empty branches.
			prevParent := prevParents[MaxTreeLevels-1-i]
			if prevParent != EmptyTree[i].NodeHash() {
				err := tx.DeleteBranch(prevParent)
				if err != nil {
					return err
				}
			}

			if parent.NodeHash() != EmptyTree[i].NodeHash() {
				err := tx.InsertBranch(parent.(*BranchNode))
				if err != nil {
					return err
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	// With our new root updated, we can update the leaf node within the
	// store. If we've inserted an empty leaf, then the leaf node found at
	// the given key is being deleted, otherise it's being inserted.
	if leaf.IsEmpty() {
		if err := tx.DeleteLeaf(*key); err != nil {
			return nil, err
		}
	} else {
		if err := tx.InsertLeaf(leaf); err != nil {
			return nil, err
		}
	}

	return root, nil
}

// Insert inserts a leaf node at the given key within the MS-SMT.
func (t *FullTree) Insert(ctx context.Context, key [hashSize]byte,
	leaf *LeafNode) (Tree, error) {

	err := t.store.Update(ctx, func(tx TreeStoreUpdateTx) error {
		root, err := t.insert(tx, &key, leaf)
		if err != nil {
			return err
		}

		return tx.UpdateRoot(root)
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Delete deletes the leaf node found at the given key within the MS-SMT.
func (t *FullTree) Delete(ctx context.Context, key [hashSize]byte) (
	Tree, error) {

	err := t.store.Update(ctx, func(tx TreeStoreUpdateTx) error {
		root, err := t.insert(tx, &key, EmptyLeafNode)
		if err != nil {
			return err
		}

		return tx.UpdateRoot(root)
	})
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Get returns the leaf node found at the given key within the MS-SMT.
func (t *FullTree) Get(ctx context.Context, key [hashSize]byte) (
	*LeafNode, error) {

	var leaf *LeafNode
	err := t.store.View(ctx, func(tx TreeStoreViewTx) error {
		var err error
		leaf, err = t.walkDown(tx, &key, nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	return leaf, nil
}

// MerkleProof generates a merkle proof for the leaf node found at the given key
// within the MS-SMT. If a leaf node does not exist at the given key, then the
// proof should be considered a non-inclusion proof. This is noted by the
// returned `Proof` containing an empty leaf.
func (t *FullTree) MerkleProof(ctx context.Context, key [hashSize]byte) (
	*Proof, error) {

	proof := make([]Node, MaxTreeLevels)
	err := t.store.View(ctx, func(tx TreeStoreViewTx) error {
		_, err := t.walkDown(
			tx, &key, func(i int, _, sibling, _ Node) error {
				proof[MaxTreeLevels-1-i] = sibling
				return nil
			},
		)
		return err
	})
	if err != nil {
		return nil, err
	}

	return NewProof(proof), nil
}

// VerifyMerkleProof determines whether a merkle proof for the leaf found at the
// given key is valid.
func VerifyMerkleProof(key [hashSize]byte, leaf *LeafNode, proof *Proof,
	root Node) bool {

	return IsEqualNode(proof.Root(key, leaf), root)
}

func (t *FullTree) RenderTree(ctx context.Context, data map[[hashSize]byte]NftData) (*VirtualTree, error) {
	// get all data from db: ref to walk down func
	virtualTree := NewVirtualTree()
	currentTree := virtualTree

	err := t.store.View(ctx, func(tx TreeStoreViewTx) error {
		for key := range data {
			current, err := tx.RootNode()
			if err != nil {
				return err
			}

			var leaf *LeafNode
			for i := 0; i <= lastBitIndex; i++ {
				left, right, err := tx.GetChildren(i, current.NodeHash())
				if err != nil {
					return err
				}

				//var next, sibling Node
				var next Node
				if bitIndex(uint8(i), &key) == 0 {
					if currentTree.Left == nil {
						currentTree.Left = NewVirtualTree()
					}

					if i == lastBitIndex {
						currentTree.Left = NewVirtualTree()
						currentTree = currentTree.Left
						leaf = left.(*LeafNode)
						continue
					}
					currentTree = currentTree.Left
					branchNode := left.(*BranchNode)
					currentTree.Hash = branchNode.nodeHash
					currentTree.Sum = branchNode.sum
					//next, sibling = left, right
					next = left
				} else {
					if currentTree.Right == nil {
						currentTree.Right = NewVirtualTree()
					}

					if i == lastBitIndex {
						currentTree.Right = NewVirtualTree()
						currentTree = currentTree.Right
						leaf = right.(*LeafNode)
						continue
					}
					currentTree = currentTree.Right
					branchNode := right.(*BranchNode)
					currentTree.Hash = branchNode.nodeHash
					currentTree.Sum = branchNode.sum
					//next, sibling = right, left
					next = right
				}

				current = next
			}

			if leaf == nil {
				continue
			}

			currentTree.Hash = leaf.nodeHash
			currentTree.Value = leaf.Value
			currentTree.Sum = &leaf.sum

			value := data[key]
			currentTree.Data = &value
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// get all data and compute key

	// mapping data into map

	// call recursion to fetch this map into tree

	// demo send tree and fetch this tree again

	// get the leaf by key in this tree & data

	// show the data inactive?

	return virtualTree, nil
}
