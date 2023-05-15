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

func RandomWIF() (*btcutil.WIF, error) {
	priv, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}

	wif, err := btcutil.NewWIF(priv, &chaincfg.SimNetParams, true)
	if err != nil {
		return nil, err
	}
	return wif, nil
}

func CommitTx(wif *btcutil.WIF, fundTx *btcutil.Tx, hash *chainhash.Hash, idx int32) (*wire.MsgTx, error) {
	tapKey := txscript.ComputeTaprootKeyNoScript(wif.PrivKey.PubKey())
	p2pktr, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(tapKey), &chaincfg.SimNetParams)
	if err != nil {
		return nil, err
	}
	p2pktrAddr := p2pktr.EncodeAddress()
	hashLockKeypair, err := RandomWIF()
	if err != nil {
		return nil, err
	}
	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(hashLockKeypair.PrivKey.PubKey()))
	builder.AddOp(txscript.OP_CHECKSIG)
	hashLockScript, err := builder.Script()
	if err != nil {
		return nil, err
	}

	tapLeaf := txscript.NewBaseTapLeaf(hashLockScript)
	tapScriptTree := txscript.AssembleTaprootScriptTree(tapLeaf)
	tapScriptRootHash := tapScriptTree.RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(
		wif.PrivKey.PubKey(), tapScriptRootHash[:],
	)
	scriptAddr, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), &chaincfg.SimNetParams)
	if err != nil {
		return nil, fmt.Errorf("error building script: %v", err)
	}
	fmt.Println(p2pktrAddr)
	fmt.Println(scriptAddr)

	/* ============================= COMMIT TX ================================== */
	commitTx, err := NewTx()
	if err != nil {
		return nil, err
	}
	outPoint := wire.NewOutPoint(hash, uint32(idx))
	txIn := wire.NewTxIn(outPoint, nil, nil)
	commitTx.AddTxIn(txIn)
	scriptAddrScript, _ := txscript.PayToAddrScript(scriptAddr)
	commitTxOut := wire.NewTxOut(fundTx.MsgTx().TxOut[0].Value*80/100, scriptAddrScript)
	commitTx.AddTxOut(commitTxOut)

	//After witness
	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(wif.PrivKey.PubKey())
	ctrlBlockBytes, _ := ctrlBlock.ToBytes()

	inputFetcher := txscript.NewCannedPrevOutputFetcher(fundTx.MsgTx().TxOut[0].PkScript, fundTx.MsgTx().TxOut[0].Value)
	sigHashes := txscript.NewTxSigHashes(commitTx, inputFetcher)
	sig, err := txscript.RawTxInTaprootSignature(
		commitTx, sigHashes, 0, fundTx.MsgTx().TxOut[0].Value,
		fundTx.MsgTx().TxOut[0].PkScript, tapScriptRootHash[:], txscript.SigHashDefault,
		wif.PrivKey,
	)
	if err != nil {
		return nil, err
	}

	commitTx.TxIn[0].Witness = wire.TxWitness{sig, hashLockScript, ctrlBlockBytes}

	return commitTx, nil
}
