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
const FirstMiningAddress = "SZnK16oMnqQt8Q1qLvrTpYLpkpkFG9eVRi"
const FirstReceiverAddress = "SVX3xWuraCUkRd7kg788LKrNNiCVvHoGsq"

func CreateCommitTx(txId string, outputIndex uint32, receiverAddress string) *wire.MsgTx {
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	hash, _ := chainhash.NewHashFromStr(txId)

	outPoint := wire.NewOutPoint(hash, outputIndex)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	redeemTx.AddTxIn(txIn)

	rcvScript := GetPayToAddrScript(receiverAddress)
	txOut := wire.NewTxOut(30, rcvScript)
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
	privKey, _, _ := GetPrivateKey(FirstSeed)
	commitTx := CreateCommitTx(result[0].TxID, result[0].Vout, FirstReceiverAddress)
	SignTx(commitTx, GetPayToAddrScript(FirstMiningAddress), privKey)
	hash, err := client.SendRawTransaction(commitTx, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	randomPrivateKey, _ := secp256k1.GeneratePrivateKey()
	revealTx, taprootAddr, err := RevealTx([]byte("Hello World"), *hash, *commitTx.TxOut[0], 0, randomPrivateKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(taprootAddr)
	fmt.Println(revealTx)

}
