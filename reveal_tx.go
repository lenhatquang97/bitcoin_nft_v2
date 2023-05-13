package main

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// revealTx spends the output from the commit transaction and as part of the
// script satisfying the tapscript spend path, posts the embedded data on
// chain. It returns the hash of the reveal transaction and error, if any.
func RevealTx(embeddedData []byte, commitTxHash chainhash.Hash, commitOutput wire.TxOut, txOutIndex uint32, randPriv *btcec.PrivateKey) (*wire.MsgTx, *btcutil.AddressTaproot, error) {
	pubKey := randPriv.PubKey()

	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(pubKey))
	builder.AddOp(txscript.OP_CHECKSIG)
	builder.AddOp(txscript.OP_0)
	builder.AddOp(txscript.OP_IF)
	chunks := ChunkSlice(embeddedData, 520)
	for _, chunk := range chunks {
		builder.AddFullData(chunk)
	}

	pkScript, err := builder.Script()
	//append op endif to prevent checking 10k size limit
	pkScript = append(pkScript, txscript.OP_ENDIF)

	if err != nil {
		return nil, nil, fmt.Errorf("error building script: %v", err)
	}

	tapLeaf := txscript.NewBaseTapLeaf(pkScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	tapScriptRootHash := tapScriptTree.RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(
		pubKey, tapScriptRootHash[:],
	)
	address, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), &chaincfg.SimNetParams)
	if err != nil {
		return nil, nil, fmt.Errorf("error building script: %v", err)
	}

	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(
		pubKey,
	)

	tx := wire.NewMsgTx(2)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  commitTxHash,
			Index: txOutIndex,
		},
	})

	opReturnScript, err := txscript.NullDataScript([]byte("https://example.com"))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating op return script: %v", err)
	}

	txOut := &wire.TxOut{
		Value: 0, PkScript: opReturnScript,
	}
	tx.AddTxOut(txOut)

	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		commitOutput.PkScript,
		commitOutput.Value,
	)
	sigHashes := txscript.NewTxSigHashes(tx, inputFetcher)

	sig, err := txscript.RawTxInTapscriptSignature(
		tx, sigHashes, 0, txOut.Value,
		txOut.PkScript, tapLeaf, txscript.SigHashDefault,
		randPriv,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("error signing tapscript: %v", err)
	}

	// Now that we have the sig, we'll make a valid witness
	// including the control block.
	ctrlBlockBytes, err := ctrlBlock.ToBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("error including control block: %v", err)
	}
	tx.TxIn[0].Witness = wire.TxWitness{
		sig, pkScript, ctrlBlockBytes,
	}

	return tx, address, nil
}
