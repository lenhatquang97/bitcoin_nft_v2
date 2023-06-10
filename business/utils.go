package business

import (
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/utils"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
)

func ComputeNftDataByte(data *NftData) ([]byte, [32]byte) {
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

func WrapperError(errStr string) error {
	return errors.New(errStr)
}

func NewRootHashForReceiver(nftData []*NftData) ([]byte, error) {
	tree := nft_tree.NewCompactedTree(nft_tree.NewDefaultStore())

	var updatedRoot *nft_tree.BranchNode
	for _, item := range nftData {
		// Compute Nft Data Info
		dataByte, key := ComputeNftDataByte(item)

		// Init Root Hash For Receiver
		leaf := nft_tree.NewLeafNode(dataByte, 0) // CoinsToSend
		leaf.NodeHash()

		updatedTree, err := tree.Insert(context.TODO(), key, leaf)
		if err != nil {
			return nil, err
		}

		updatedRoot, err = updatedTree.Root(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return utils.GetNftRoot(updatedRoot), nil
}

func FileSha256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new SHA256 hash
	hash := sha256.New()

	// Copy the file contents to the hash calculator
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Get the hash sum as a byte slice
	hashSum := hash.Sum(nil)
	return hex.EncodeToString(hashSum), nil
}

func RawDataEncode(data string) (string, error) {
	// Create a new SHA256 hash
	hash := sha256.New()
	hash.Write([]byte(data))

	// Get the hash sum as a byte slice
	hashSum := hash.Sum(nil)
	return hex.EncodeToString(hashSum), nil
}
