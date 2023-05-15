package main

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func SendMoneyToTaprootAddress(amount int64, client *rpcclient.Client, senderAddress btcutil.Address) (*wire.MsgTx, error) {
	wif, err := client.DumpPrivKey(senderAddress)
	if err != nil {
		return nil, err
	}

	utxos, err := client.ListUnspent()
	if err != nil {
		return nil, err
	}

	sendUtxos := GetManyUtxo(utxos, senderAddress.EncodeAddress(), float64(amount))
	if len(sendUtxos) == 0 {
		return nil, fmt.Errorf("no utxos")
	}

	var balance float64
	for _, item := range sendUtxos {
		balance += item.Amount
	}

	pkScript, _ := txscript.PayToAddrScript(senderAddress)

	if err != nil {
		return nil, err
	}

	// checking for sufficiency of account
	if int64(balance) < amount {
		return nil, fmt.Errorf("the balance of the account is not sufficient")
	}

	tweakedPubKey, err := schnorr.ParsePubKey(schnorr.SerializePubKey(wif.PrivKey.PubKey()))
	if err != nil {
		return nil, err
	}
	destinationAddr, err := btcutil.NewAddressTaproot(tweakedPubKey.X().Bytes(), &chaincfg.SimNetParams)
	if err != nil {
		return nil, err
	}

	fmt.Println(destinationAddr)

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return nil, err
	}

	redeemTx, err := NewTx()
	if err != nil {
		return nil, err
	}

	for _, utxo := range sendUtxos {
		utxoHash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, err
		}

		outPoint := wire.NewOutPoint(utxoHash, utxo.Vout)

		// making the input, and adding it to transaction
		txIn := wire.NewTxIn(outPoint, nil, nil)
		redeemTx.AddTxIn(txIn)
	}

	// adding the destination address and the amount to
	// the transaction as output
	redeemTxOut := wire.NewTxOut(amount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	// now sign the transaction
	finalRawTx, err := SignTx(wif, pkScript, redeemTx)
	if err != nil {
		return nil, err
	}

	return finalRawTx, nil
}
