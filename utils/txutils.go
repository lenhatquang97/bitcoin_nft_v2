package utils

import (
	"bitcoin_nft_v2/nft_tree"
	"crypto/sha256"
	"encoding/binary"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func SignTx(wif *btcutil.WIF, sendUtxos []*MyUtxo, redeemTx *wire.MsgTx) (*wire.MsgTx, error) {
	for i := range redeemTx.TxIn {
		addressObj, err := btcutil.DecodeAddress(sendUtxos[i].Address, &chaincfg.TestNet3Params)
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addressObj)
		if err != nil {
			return nil, err
		}

		signature, err := txscript.SignatureScript(redeemTx, i, pkScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return nil, err
		}

		redeemTx.TxIn[i].SignatureScript = signature
	}
	return redeemTx, nil
}

func CreateInscriptionScript(pubKey *secp256k1.PublicKey, embeddedData []byte) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(pubKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddOp(txscript.OP_0)
	builder.AddOp(txscript.OP_IF)
	chunks := ChunkSlice(embeddedData, 500)
	for i, chunk := range chunks {
		if i == 0 {
			var tmp []byte
			tmp = append(tmp, []byte("m25start")...)
			tmp = append(tmp, chunk...)
			builder.AddFullData(tmp)
		} else if i == len(chunks)-1 {
			var tmp []byte
			tmp = append(tmp, chunk...)
			tmp = append(tmp, []byte("m25end")...)
			builder.AddFullData(tmp)
		} else {
			builder.AddFullData(chunk)
		}
	}
	hashLockScript, err := builder.Script()
	if err != nil {
		return nil, err
	}
	hashLockScript = append(hashLockScript, txscript.OP_ENDIF)
	return hashLockScript, nil
}

func CreateInscriptionScriptV2(pubKey *secp256k1.PublicKey, embeddedData []byte, isRef bool) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(pubKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddOp(txscript.OP_0)
	builder.AddOp(txscript.OP_IF)
	chunks := ChunkSlice(embeddedData, 500)
	flagStart := "m25start"
	flagEnd := "m25end"
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
		_ = binary.Write(h, binary.BigEndian, c.NodeSum())
		return h.Sum(nil)
	}

	left := c.Left.NodeHash()
	right := c.Right.NodeHash()

	h := sha256.New()
	//_, _ = h.Write(c.AssetID[:])
	_, _ = h.Write(left[:])
	_, _ = h.Write(right[:])
	_ = binary.Write(h, binary.BigEndian, c.NodeSum())
	return h.Sum(nil)
}

func CreateOutputKeyBasedOnScript(pubKey *secp256k1.PublicKey, script []byte) (*secp256k1.PublicKey, *txscript.IndexedTapScriptTree, *txscript.TapLeaf) {
	tapLeaf := txscript.NewBaseTapLeaf(script)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	tapScriptRootHash := tapScriptTree.LeafMerkleProofs[0].RootNode.TapHash()
	return txscript.ComputeTaprootOutputKey(pubKey, tapScriptRootHash[:]), tapScriptTree, &tapLeaf
}
