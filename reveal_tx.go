package main

import (
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func RevealTx(embeddedData []byte, commitTxHash chainhash.Hash, commitOutput wire.TxOut, txOutIndex uint32, randPriv *btcec.PrivateKey, params *chaincfg.Params) (*wire.MsgTx, *btcutil.AddressTaproot, error) {
	pubKey := randPriv.PubKey()
	pkScript, err := utils.CreateInscriptionScript(pubKey, embeddedData)

	if err != nil {
		return nil, nil, fmt.Errorf("error building script: %v", err)
	}
	outputKey, tapScriptTree, tapLeaf := utils.CreateOutputKeyBasedOnScript(pubKey, pkScript)

	outputScriptBuilder := txscript.NewScriptBuilder()
	outputScriptBuilder.AddOp(txscript.OP_1)
	outputScriptBuilder.AddData(schnorr.SerializePubKey(outputKey))
	outputScript, _ := outputScriptBuilder.Script()

	address, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), params)
	if err != nil {
		return nil, nil, err
	}

	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(pubKey)

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
		txOut.PkScript, *tapLeaf, txscript.SigHashDefault,
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

	fmt.Println("==================================Validation=============================================")
	engine, err := txscript.NewEngine(outputScript, tx, 0, txscript.StandardVerifyFlags, nil, sigHashes, 8000, inputFetcher)
	if err != nil {
		return nil, nil, err
	}
	err = engine.Execute()
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("===================================Success validation====================================")
	return tx, address, nil
}
