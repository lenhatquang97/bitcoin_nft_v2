package main

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
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

	rawTx, wif, err := CreateTx("SZnK16oMnqQt8Q1qLvrTpYLpkpkFG9eVRi", 40, client)

	if err != nil {
		fmt.Println(err)
		return
	}

	hash, err := client.SendRawTransaction(rawTx, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Success")

	commitTx, err := client.GetRawTransaction(hash)
	if err != nil {
		fmt.Println(err)
		return
	}

	tx, address, err := RevealTx([]byte("Hello World"), *hash, *commitTx.MsgTx().TxOut[0], 0, wif.PrivKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(address)
	revealTx, err := client.SendRawTransaction(tx, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(revealTx)
}

func NewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(wire.TxVersion), nil
}

func GetUtxo(utxos []btcjson.ListUnspentResult, address string) (string, uint32, float64) {
	for i := 0; i < len(utxos); i++ {
		if utxos[i].Address == address {
			return utxos[i].TxID, utxos[i].Vout, utxos[i].Amount
		}
	}
	return "", 0, -1
}

func CreateTx(destination string, amount int64, client *rpcclient.Client) (*wire.MsgTx, *btcutil.WIF, error) {
	defaultAddress, err := btcutil.DecodeAddress("SeZdpbs8WBuPHMZETPWajMeXZt1xzCJNAJ", &chaincfg.SimNetParams)
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

	txid, vout, balance := GetUtxo(utxos, defaultAddress.EncodeAddress())
	if len(txid) == 0 {
		return nil, nil, fmt.Errorf("no utxos")
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
	destinationAddr, err := btcutil.DecodeAddress(destination, &chaincfg.SimNetParams)
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

	utxoHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return nil, nil, err
	}

	outPoint := wire.NewOutPoint(utxoHash, vout)

	// making the input, and adding it to transaction
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

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

func SignTx(wif *btcutil.WIF, pkScript []byte, redeemTx *wire.MsgTx) (*wire.MsgTx, error) {
	signature, err := txscript.SignatureScript(redeemTx, 0, pkScript, txscript.SigHashAll, wif.PrivKey, true)
	if err != nil {
		return nil, err
	}

	redeemTx.TxIn[0].SignatureScript = signature

	return redeemTx, nil
}
