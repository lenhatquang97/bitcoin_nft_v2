package main

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	senderAddress = "SeZdpbs8WBuPHMZETPWajMeXZt1xzCJNAJ"
)

func main() {
	client, err := GetBitcoinWalletRpcClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.WalletPassphrase("12345", 5)
	if err != nil {
		fmt.Println(err)
		return
	}

	senderAddressObj, err := btcutil.DecodeAddress(senderAddress, &chaincfg.SimNetParams)

	if err != nil {
		fmt.Println(err)
		return
	}
	wif, err := client.DumpPrivKey(senderAddressObj)
	if err != nil {
		fmt.Println(err)
		return
	}

	firstTx, err := SendMoneyToTaprootAddress(10000, client, senderAddressObj)
	if err != nil {
		fmt.Println(err)
		return
	}

	hashFirstTx, err := client.SendRawTransaction(firstTx, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("===================================Checkpoint 1====================================")

	availableUtxos, err := client.GetRawTransaction(hashFirstTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = StartTapTree(client, wif, []byte("Hello World"), hashFirstTx, 0, availableUtxos.MsgTx().TxOut[0])
	if err != nil {
		fmt.Println(err)
		return
	}
}

func MakeRandomKeyPair() (*btcutil.WIF, error) {
	randPriv, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	wif, err := btcutil.NewWIF(randPriv, &chaincfg.SimNetParams, true)
	if err != nil {
		return nil, err
	}
	return wif, nil
}

func StartTapTree(client *rpcclient.Client, keyPair *btcutil.WIF, data []byte, hash *chainhash.Hash, index uint32, commitOutput *wire.TxOut) error {
	// tweakedPubKey, err := schnorr.ParsePubKey(schnorr.SerializePubKey(keyPair.PrivKey.PubKey()))
	// if err != nil {
	// 	return err
	// }
	// p2pktr, err := btcutil.NewAddressTaproot(tweakedPubKey.X().Bytes(), &chaincfg.SimNetParams)
	// if err != nil {
	// 	return err
	// }

	hashLockKeypair, err := MakeRandomKeyPair()
	if err != nil {
		return err
	}

	builder := txscript.NewScriptBuilder()
	builder.AddData(hashLockKeypair.PrivKey.PubKey().X().Bytes())
	builder.AddOp(txscript.OP_CHECKSIG)
	hashLockScript, err := builder.Script()
	if err != nil {
		return err
	}

	var allTreeLeaves []txscript.TapLeaf
	tapLeaf := txscript.NewBaseTapLeaf(hashLockScript)
	allTreeLeaves = append(allTreeLeaves, tapLeaf)
	tapTree := TapscriptFullTree(keyPair.PrivKey.PubKey(), allTreeLeaves...)
	taprootKey, err := tapTree.TaprootKey()
	if err != nil {
		return err
	}
	scriptAddr, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(taprootKey), &chaincfg.SimNetParams)
	if err != nil {
		return err
	}

	// p2pkP2tr, err := btcutil.NewAddressTaproot(keyPair.PrivKey.PubKey().X().Bytes(), &chaincfg.SimNetParams)
	// if err != nil {
	// 	return err
	// }

	//Commit transaction
	commitTx, err := NewTx()
	if err != nil {
		return err
	}
	commitTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: *wire.NewOutPoint(hash, index),
	})
	scriptAddrScript, _ := txscript.PayToAddrScript(scriptAddr)
	commitTx.AddTxOut(&wire.TxOut{
		Value:    commitOutput.Value * 80 / 100,
		PkScript: scriptAddrScript,
	})

	inputFetcher := txscript.NewCannedPrevOutputFetcher(scriptAddrScript, commitOutput.Value*80/100)
	sigHashes := txscript.NewTxSigHashes(commitTx, inputFetcher)

	sig, err := txscript.RawTxInTapscriptSignature(commitTx, sigHashes, 0, commitOutput.Value*80/100, scriptAddrScript, tapLeaf, txscript.SigHashDefault, keyPair.PrivKey)
	if err != nil {
		return err
	}

	controlBlock, err := tapTree.ControlBlock.ToBytes()
	if err != nil {
		return err
	}
	commitTx.TxIn[0].Witness = wire.TxWitness{sig, hashLockScript, controlBlock}

	finalHash, err := client.SendRawTransaction(commitTx, false)
	if err != nil {
		return err
	}
	fmt.Println(finalHash)
	return nil
}
