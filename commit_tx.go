package main

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func CreateTaprootCommitTransaction(embeddedData []byte, previousHash *chainhash.Hash, idx uint32, commitOutput wire.TxOut, randPriv *btcec.PrivateKey, amount int64) (*wire.MsgTx, error) {
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
	builder.AddOp(txscript.OP_ENDIF)
	pkScript, err := builder.Script()

	if err != nil {
		return nil, fmt.Errorf("error building script: %v", err)
	}

	tapLeaf := txscript.NewBaseTapLeaf(pkScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	tapScriptRootHash := tapScriptTree.LeafMerkleProofs[0].RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(
		pubKey, tapScriptRootHash[:],
	)
	outputScriptBuilder := txscript.NewScriptBuilder()
	outputScriptBuilder.AddOp(txscript.OP_1)
	outputScriptBuilder.AddData(schnorr.SerializePubKey(outputKey))
	outputScript, _ := outputScriptBuilder.Script()
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  *previousHash,
			Index: idx,
		},
	})
	tx.AddTxOut(&wire.TxOut{
		Value:    amount,
		PkScript: outputScript,
	})

	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		commitOutput.PkScript,
		commitOutput.Value,
	)
	sigHashes := txscript.NewTxSigHashes(tx, inputFetcher)

	sig, err := txscript.RawTxInTapscriptSignature(
		tx, sigHashes, 0, amount,
		outputScript, tapLeaf, txscript.SigHashDefault,
		randPriv,
	)
	if err != nil {
		return nil, err
	}
	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(pubKey)
	ctrlBlockBytes, err := ctrlBlock.ToBytes()
	if err != nil {
		return nil, fmt.Errorf("error including control block: %v", err)
	}
	tx.TxIn[0].Witness = wire.TxWitness{sig, pkScript, ctrlBlockBytes}

	engine, err := txscript.NewEngine(outputScript, tx, 0, txscript.StandardVerifyFlags, nil, sigHashes, amount, inputFetcher)
	if err != nil {
		return nil, err
	}
	err = engine.Execute()
	if err != nil {
		return nil, err
	}
	return tx, nil
}
