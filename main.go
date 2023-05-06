package main

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const FirstSeed = "d94155d877b8150f6215ad5bc6917989fd88888c045a21791fed17e0ae916bec"
const ReceiverAddress = "mv4rnyY3Su5gjcDNzbMLKBQkBicCtHUtFB"

func CreateCommitTx(txId string, outputIndex uint32, receiverAddress string) *wire.MsgTx {
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	hash, _ := chainhash.NewHashFromStr(txId)

	outPoint := wire.NewOutPoint(hash, outputIndex)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	rcvScript := GetPayToAddrScript(receiverAddress)
	txOut := wire.NewTxOut(1000, rcvScript)
	redeemTx.AddTxOut(txOut)

	return redeemTx
}

func SignTx(redeemTx *wire.MsgTx, subscript []byte, privKey *secp256k1.PrivateKey) {
	sig, err := txscript.SignatureScript(redeemTx, 0, subscript, txscript.SigHashAll, privKey, false)
	if err != nil {
		log.Fatalf("could not generate signature: %v", err)
	}
	redeemTx.TxIn[0].SignatureScript = sig
}

func main() {
	client, err := GetBitcoinWalletRpcClient()
	if err != nil {
		fmt.Println(err)
		return
	}
	result, err := client.ListUnspent()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result[0].Address)
}
