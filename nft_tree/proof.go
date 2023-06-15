package nft_tree

// Proof represents a merkle proof for a MS-SMT.
type Proof struct {
	// Nodes represents the siblings that should be hashed with the leaf and
	// its parents to arrive at the root of the MS-SMT.
	Nodes []Node
}

// NewProof initializes a new merkle proof for the given leaf node.
func NewProof(nodes []Node) *Proof {
	return &Proof{
		Nodes: nodes,
	}
}

// Root returns the root node obtained by walking up the tree.
func (p Proof) Root(key [32]byte, leaf *LeafNode) *BranchNode {
	// Note that we don't need to check the error here since the only point
	// where the error could come from is the passed iterator which is nil.
	node, _ := walkUp(&key, leaf, p.Nodes, nil)
	return node
}
