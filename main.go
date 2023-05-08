package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const FirstSeed = "d94155d877b8150f6215ad5bc6917989fd88888c045a21791fed17e0ae916bec"
const FirstMiningAddress = "SZnK16oMnqQt8Q1qLvrTpYLpkpkFG9eVRi"
const FirstReceiverAddress = "SVX3xWuraCUkRd7kg788LKrNNiCVvHoGsq"

const SecondSeed = "ffdeb8d75f9fe2021e154c431510d08e1e95918cc28cdcb8f90b0119577d7a40"
const SecondMiningAddress = "SeTCfjeSQYevShUDEqo59GH1V5kqnP4dg5"

func NewTx() (*wire.MsgTx, error) {
	return wire.NewMsgTx(wire.TxVersion), nil
}

func SignTx(privKey *secp256k1.PrivateKey, pkScript string, redeemTx *wire.MsgTx) (*wire.MsgTx, string, error) {

	//wif, err := btcutil.DecodeWIF(privKey)
	//if err != nil {
	//	return nil, "", err
	//}

	sourcePKScript, err := hex.DecodeString(pkScript)
	if err != nil {
		return nil, "", nil
	}

	// since there is only one input in our transaction
	// we use 0 as second argument, if the transaction
	// has more args, should pass related index
	signature, err := txscript.SignatureScript(redeemTx, 0, sourcePKScript, txscript.SigHashAll, privKey, false)
	if err != nil {
		return nil, "", nil
	}

	// since there is only one input, and want to add
	// signature to it use 0 as index
	redeemTx.TxIn[0].SignatureScript = signature

	var signedTx bytes.Buffer
	redeemTx.Serialize(&signedTx)

	hexSignedTx := hex.EncodeToString(signedTx.Bytes())

	return redeemTx, hexSignedTx, nil
}

func GetUTXO(address string) (string, float64, string, error) {
	// Provide your url to get UTXOs, read the response
	// unmarshal it, and extract necessary data
	client, err := GetBitcoinWalletRpcClient()
	if err != nil {
		fmt.Println(err)
		return "", 0, "", err
	}

	utxos, err := client.ListUnspent()
	if err != nil {
		fmt.Println(err)
		return "", 0, "", err
	}

	txId1Ref := "66cd068c04778bf07734bf358982221cfb3edad30b41b11eebbf49b1318c6876"
	txId2Ref := "67b35b34dfbd0cb25c0779877e8c062094112ddb1c0f9fd81c51b6cafe4e02b0"
	fmt.Println(txId1Ref)

	for _, utxo := range utxos {
		if utxo.TxID == txId2Ref {
			fmt.Println("found....")
		}
	}
	for _, utxo := range utxos {
		return utxo.TxID, utxo.Amount, utxo.ScriptPubKey, nil
	}

	return "", 0, "", nil
}

func CreateCommitTx(privKey *secp256k1.PrivateKey, destination string, amount int64) (*wire.MsgTx, string, error) {
	//wif, err := btcutil.DecodeWIF(privKey)
	//if err != nil {
	//	return nil, "", err
	//}

	// use TestNet3Params for interacting with bitcoin testnet
	// if we want to interact with main net should use MainNetParams
	addrPubKey, err := btcutil.NewAddressPubKey(privKey.PubKey().SerializeUncompressed(), &chaincfg.SimNetParams)
	if err != nil {
		return nil, "", err
	}

	txid, balance, pkScript, err := GetUTXO(addrPubKey.EncodeAddress())
	if err != nil {
		return nil, "", err
	}

	/*
	 * 1 or unit-amount in Bitcoin is equal to 1 satoshi and 1 Bitcoin = 100000000 satoshi
	 */

	// checking for sufficiency of account
	if balance < float64(amount) {
		return nil, "", fmt.Errorf("the balance of the account is not sufficient")
	}

	// extracting destination address as []byte from function argument (destination string)
	destinationAddr, err := btcutil.DecodeAddress(destination, &chaincfg.SimNetParams)
	if err != nil {
		return nil, "", err
	}

	destinationAddrByte, err := txscript.PayToAddrScript(destinationAddr)
	if err != nil {
		return nil, "", err
	}

	// creating a new bitcoin transaction, different sections of the tx, including
	// input list (contain UTXOs) and outputlist (contain destination address and usually our address)
	// in next steps, sections will be field and pass to sign
	redeemTx, err := NewTx()
	if err != nil {
		return nil, "", err
	}

	utxoHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return nil, "", err
	}

	// the second argument is vout or Tx-index, which is the index
	// of spending UTXO in the transaction that Txid referred to
	// in this case is 0, but can vary different numbers
	outPoint := wire.NewOutPoint(utxoHash, 0)

	// making the input, and adding it to transaction
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	// adding the destination address and the amount to
	// the transaction as output
	redeemTxOut := wire.NewTxOut(amount, destinationAddrByte)
	redeemTx.AddTxOut(redeemTxOut)

	// now sign the transaction
	signedTx, finalRawTx, err := SignTx(privKey, pkScript, redeemTx)

	return signedTx, finalRawTx, nil
}

func main() {
	client, err := GetBitcoinWalletRpcClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	//result, err := client.ListUnspent()
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	privKey, _, _ := GetPrivateKey(FirstSeed)
	//commitTx := CreateCommitTx(result[0].TxID, result[0].Vout, FirstReceiverAddress)
	//SignTx(commitTx, GetPayToAddrScript(FirstMiningAddress), privKey)
	//hash, err := client.SendRawTransaction(commitTx, false)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//randomPrivateKey, _ := secp256k1.GeneratePrivateKey()
	//revealTx, taprootAddr, err := RevealTx([]byte("Hello World"), *hash, *commitTx.TxOut[0], 0, randomPrivateKey)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//fmt.Println(taprootAddr)
	//fmt.Println(revealTx)
	signTx, rawTx, err := CreateCommitTx(privKey,
		FirstReceiverAddress, 10)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Raw tx is: ", rawTx)
	_, err = client.SendRawTransaction(signTx, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Perfect hahaha")
}
