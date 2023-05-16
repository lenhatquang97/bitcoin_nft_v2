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

	tx, wif, err := CreateTxV2(10000, client)
	if err != nil {
		fmt.Println(err)
		return
	}

	hashTx, err := client.SendRawTransaction(tx, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	rawTx, err := client.GetRawTransaction(hashTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("===================================Checkpoint 1====================================")

	finalTx, _, err := RevealTx([]byte("Hello World"), *hashTx, *rawTx.MsgTx().TxOut[0], 0, wif.PrivKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	finalHash, err := client.SendRawTransaction(finalTx, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(finalHash)

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

func CreateTxV2(amount int64, client *rpcclient.Client) (*wire.MsgTx, *btcutil.WIF, error) {
	defaultAddress, err := btcutil.DecodeAddress(senderAddress, &chaincfg.SimNetParams)
	if err != nil {
		return nil, nil, err
	}

	wif, err := client.DumpPrivKey(defaultAddress)
	if err != nil {
		return nil, nil, err
	}

	utxos, err := client.ListUnspent()
	if err != nil {
		return nil, nil, err
	}

	sendUtxos := GetManyUtxo(utxos, defaultAddress.EncodeAddress(), float64(amount))
	if len(sendUtxos) == 0 {
		return nil, nil, fmt.Errorf("no utxos")
	}

	//PrintLogUtxos(sendUtxos)

	var balance float64
	for _, item := range sendUtxos {
		balance += item.Amount
	}

	pkScript, _ := txscript.PayToAddrScript(defaultAddress)

	if err != nil {
		return nil, nil, err
	}

	// checking for sufficiency of account
	if int64(balance) < amount {
		return nil, nil, fmt.Errorf("the balance of the account is not sufficient")
	}

	// extracting destination address as []byte from function argument (destination string)
	tweakedPubKey, err := schnorr.ParsePubKey(schnorr.SerializePubKey(wif.PrivKey.PubKey()))
	if err != nil {
		return nil, nil, err
	}
	destinationAddr, err := btcutil.NewAddressTaproot(tweakedPubKey.X().Bytes(), &chaincfg.SimNetParams)
	if err != nil {
		return nil, nil, err
	}

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return nil, nil, err
	}

	redeemTx, err := NewTx()
	if err != nil {
		return nil, nil, err
	}

	for _, utxo := range sendUtxos {
		utxoHash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, nil, err
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
		return nil, nil, err
	}

	return finalRawTx, wif, nil
}
