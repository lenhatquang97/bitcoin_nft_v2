package nft_tree

type NftData struct {
	ID   string
	Url  string
	Memo string
}

type VirtualTree struct {
	Left  *VirtualTree
	Right *VirtualTree
	Hash  *NodeHash
	Value []byte
	Sum   *uint64
	Data  *NftData
}

func NewVirtualTree() *VirtualTree {
	return &VirtualTree{
		Left:  nil,
		Right: nil,
		Hash:  nil,
		Value: []byte{},
		Data:  nil,
	}
}
