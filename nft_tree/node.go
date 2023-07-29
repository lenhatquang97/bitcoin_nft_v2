package nft_tree

import (
	"crypto/sha256"
	"encoding/hex"
)

const (
	// hashSize is the size of hashes used in the MS-SMT.
	hashSize = sha256.Size
)

var (
	// EmptyLeafNode represents an empty leaf in a MS-SMT, one with a nil
	// value and 0 sum.
	EmptyLeafNode = NewLeafNode(nil)
)

// NodeHash represents the key of a MS-SMT node.
type NodeHash [hashSize]byte

// String returns a NodeHash as a hex-encoded string.
func (k NodeHash) String() string {
	return hex.EncodeToString(k[:])
}

// Node represents a MS-SMT node. A node can either be a leaf or a branch.
type Node interface {
	// NodeHash returns the unique identifier for a MS-SMT node. It
	// represents the hash of the node committing to its internal data.
	NodeHash() NodeHash

	// Copy returns a deep copy of the node.
	Copy() Node
}

// IsEqualNode determines whether a and b are equal based on their NodeHash and
// NodeSum.
func IsEqualNode(a, b Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.NodeHash() == b.NodeHash()
}

// LeafNode represents a leaf node within a MS-SMT. Leaf nodes commit to a value
// and some integer value (the sum) associated with the value.
type LeafNode struct {
	// Cached nodeHash instance to prevent redundant computations.
	nodeHash *NodeHash

	Value []byte
}

// NewLeafNode constructs a new leaf node.
func NewLeafNode(value []byte) *LeafNode {
	return &LeafNode{
		Value: value,
	}
}

// NodeHash returns the unique identifier for a MS-SMT node. It represents the
// hash of the leaf committing to its internal data.
func (n *LeafNode) NodeHash() NodeHash {
	if n.nodeHash != nil {
		return *n.nodeHash
	}

	h := sha256.New()
	h.Write(n.Value)
	//_ = binary.Write(h, binary.BigEndian, n.sum)
	n.nodeHash = (*NodeHash)(h.Sum(nil))
	return *n.nodeHash
}

// IsEmpty returns whether this is an empty leaf.
func (n *LeafNode) IsEmpty() bool {
	return len(n.Value) == 0
}

// Copy returns a deep copy of the leaf node.
func (n *LeafNode) Copy() Node {
	var nodeHashCopy *NodeHash
	if n.nodeHash != nil {
		nodeHashCopy = new(NodeHash)
		*nodeHashCopy = *n.nodeHash
	}

	valueCopy := make([]byte, 0, len(n.Value))
	copy(valueCopy, n.Value)

	return &LeafNode{
		nodeHash: nodeHashCopy,
		Value:    valueCopy,
	}
}

// CompactedLeafNode holds a leafnode that represents a whole "compacted"
// subtree omitting all default branches and leafs in the represented subtree.
type CompactedLeafNode struct {
	*LeafNode

	// key holds the leaf's key.
	key [32]byte

	// compactedNodeHash holds the topmost (omitted) node's node hash in the
	// subtree.
	compactedNodeHash NodeHash
}

// newCompactedLeafNode creates a new compacted leaf at the passed height with
// the passed leaf key.
func NewCompactedLeafNode(height int, key *[32]byte,
	leaf *LeafNode) *CompactedLeafNode {

	var current Node = leaf
	for i := lastBitIndex; i >= height; i-- {
		if bitIndex(uint8(i), key) == 0 {
			current = NewBranch(current, EmptyTree[i+1])
		} else {
			current = NewBranch(EmptyTree[i+1], current)
		}
	}
	nodeHash := current.NodeHash()

	node := &CompactedLeafNode{
		LeafNode:          leaf,
		key:               *key,
		compactedNodeHash: nodeHash,
	}

	return node
}

// NodeHash returns the compacted subtree's node hash.
func (c *CompactedLeafNode) NodeHash() NodeHash {
	return c.compactedNodeHash
}

// Key returns the leaf key.
func (c *CompactedLeafNode) Key() [32]byte {
	return c.key
}

// Extract extracts the subtree represented by this compacted leaf and returns
// the topmost node in the tree.
func (c *CompactedLeafNode) Extract(height int) Node {
	var current Node = c.LeafNode

	// Walk up and recreate the missing branches.
	for j := MaxTreeLevels; j > height+1; j-- {
		var left, right Node
		if bitIndex(uint8(j-1), &c.key) == 0 {
			left, right = current, EmptyTree[j]
		} else {
			left, right = EmptyTree[j], current
		}

		current = NewBranch(left, right)
	}

	return current
}

// BranchNode represents an intermediate or root node within a MS-SMT. It
// commits to its left and right children, along with their respective sum
// values.
type BranchNode struct {
	// Cached instances to prevent redundant computations.
	nodeHash *NodeHash

	Left  Node
	Right Node
}

// NewComputedBranch creates a new branch without any reference it its
// children. This method of construction allows as to walk the tree down by
// only fetching minimal subtrees.
func NewComputedBranch(nodeHash NodeHash) *BranchNode {
	return &BranchNode{
		nodeHash: &nodeHash,
	}
}

// NewBranch constructs a new branch backed by its left and right children.
func NewBranch(left, right Node) *BranchNode {
	return &BranchNode{
		Left:  left,
		Right: right,
	}
}

// NodeHash returns the unique identifier for a MS-SMT node. It represents the
// hash of the branch committing to its internal data.
func (n *BranchNode) NodeHash() NodeHash {
	if n.nodeHash != nil {
		return *n.nodeHash
	}

	left := n.Left.NodeHash()
	right := n.Right.NodeHash()

	h := sha256.New()
	h.Write(left[:])
	h.Write(right[:])
	//_ = binary.Write(h, binary.BigEndian, n.NodeSum())
	n.nodeHash = (*NodeHash)(h.Sum(nil))
	return *n.nodeHash
}

// Copy returns a deep copy of the branch node, with its children returned as
// `ComputedNode`.
func (n *BranchNode) Copy() Node {
	var nodeHashCopy *NodeHash
	if n.nodeHash != nil {
		nodeHashCopy = new(NodeHash)
		*nodeHashCopy = *n.nodeHash
	}

	return &BranchNode{
		nodeHash: nodeHashCopy,
		Left:     NewComputedNode(n.Left.NodeHash()),
		Right:    NewComputedNode(n.Right.NodeHash()),
	}
}

// ComputedNode is a node within a MS-SMT that has already had its NodeHash and
// NodeSum computed, i.e., its preimage is not available.
type ComputedNode struct {
	hash NodeHash
}

// NewComputedNode instantiates a new computed node.
func NewComputedNode(hash NodeHash) ComputedNode {
	return ComputedNode{hash: hash}
}

// NodeHash returns the unique identifier for a MS-SMT node. It represents the
// hash of the node committing to its internal data.
func (n ComputedNode) NodeHash() NodeHash {
	return n.hash
}

// Copy returns a deep copy of the branch node.
func (n ComputedNode) Copy() Node {
	return ComputedNode{
		hash: n.hash,
	}
}
