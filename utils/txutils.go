package utils

import (
	"bitcoin_nft_v2/nft_tree"
	"crypto/sha256"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func CreateInscriptionScriptV2(pubKey *secp256k1.PublicKey, embeddedData []byte, isRef bool, mode string) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(pubKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddOp(txscript.OP_0)
	builder.AddOp(txscript.OP_IF)
	chunks := ChunkSlice(embeddedData, 500)
	flagStart := "m25start"
	flagEnd := "m25end"
	if mode == "off_chain" {
		flagStart = "m25off-chain-start"
		flagEnd = "m25off-chain-end"
	}
	flagRef := "-ref"
	flagData := "-data"
	if isRef {
		flagStart += flagRef
		flagEnd += flagRef
	} else {
		flagStart += flagData
		flagEnd += flagData
	}

	for i, chunk := range chunks {
		if i == 0 {
			var tmp []byte
			tmp = append(tmp, []byte(flagStart)...)
			tmp = append(tmp, chunk...)
			builder.AddFullData(tmp)
		} else if i == len(chunks)-1 {
			var tmp []byte
			tmp = append(tmp, chunk...)
			tmp = append(tmp, []byte(flagEnd)...)
			builder.AddFullData(tmp)
		} else {
			builder.AddFullData(chunk)
		}
	}

	if len(chunks) == 1 {
		var tmp []byte
		tmp = append(tmp, []byte(flagEnd)...)
		builder.AddFullData(tmp)
	}

	hashLockScript, err := builder.Script()
	if err != nil {
		return nil, err
	}
	hashLockScript = append(hashLockScript, txscript.OP_ENDIF)
	return hashLockScript, nil
}

func GetNftRoot(c *nft_tree.BranchNode) []byte {
	if c.NodeHash().String() != "" {
		nodeHash := c.NodeHash()
		h := sha256.New()
		_, _ = h.Write(nodeHash[:])
		return h.Sum(nil)
	}

	left := c.Left.NodeHash()
	right := c.Right.NodeHash()

	h := sha256.New()
	//_, _ = h.Write(c.AssetID[:])
	_, _ = h.Write(left[:])
	_, _ = h.Write(right[:])
	return h.Sum(nil)
}

func CreateOutputKeyBasedOnScript(pubKey *secp256k1.PublicKey, script []byte) (*secp256k1.PublicKey, *txscript.IndexedTapScriptTree, *txscript.TapLeaf) {
	tapLeaf := txscript.NewBaseTapLeaf(script)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	tapScriptRootHash := tapScriptTree.LeafMerkleProofs[0].RootNode.TapHash()
	return txscript.ComputeTaprootOutputKey(pubKey, tapScriptRootHash[:]), tapScriptTree, &tapLeaf
}
