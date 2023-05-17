package utils

import (
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func SignTx(wif *btcutil.WIF, pkScript []byte, redeemTx *wire.MsgTx) (*wire.MsgTx, error) {
	for i := range redeemTx.TxIn {
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
	chunks := ChunkSlice(embeddedData, 520)
	for _, chunk := range chunks {
		builder.AddFullData([]byte("m25"))
		builder.AddFullData(chunk)
	}
	hashLockScript, err := builder.Script()
	if err != nil {
		return nil, err
	}
	hashLockScript = append(hashLockScript, txscript.OP_ENDIF)
	return hashLockScript, nil
}

func CreateOutputKeyBasedOnScript(pubKey *secp256k1.PublicKey, script []byte) (*secp256k1.PublicKey, *txscript.IndexedTapScriptTree, *txscript.TapLeaf) {
	tapLeaf := txscript.NewBaseTapLeaf(script)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	tapScriptRootHash := tapScriptTree.LeafMerkleProofs[0].RootNode.TapHash()
	return txscript.ComputeTaprootOutputKey(pubKey, tapScriptRootHash[:]), tapScriptTree, &tapLeaf
}

func GetDefaultAddress(client *rpcclient.Client, senderAddress string, config *chaincfg.Params) (btcutil.Address, error) {
	if len(senderAddress) == 0 {
		testNetAddress, err := client.GetAccountAddress("default")
		if err != nil {
			return nil, err
		}
		return testNetAddress, nil
	}
	simNetAddress, err := btcutil.DecodeAddress(senderAddress, config)
	if err != nil {
		return nil, err
	}
	return simNetAddress, nil
}
