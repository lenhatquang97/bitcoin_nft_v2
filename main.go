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

	rawTx, wif, err := CreateTxV2("SZnK16oMnqQt8Q1qLvrTpYLpkpkFG9eVRi", 120, client)

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

type MyUtxo struct {
	TxID   string
	Vout   uint32
	Amount float64
}

func GetManyUtxo(utxos []btcjson.ListUnspentResult, address string, amount float64) []*MyUtxo {
	var myUtxos []*MyUtxo
	for i := 0; i < len(utxos); i++ {
		if utxos[i].Address == address {
			myUtxos = append(myUtxos, &MyUtxo{
				TxID:   utxos[i].TxID,
				Vout:   utxos[i].Vout,
				Amount: utxos[i].Amount,
			})
		}
	}
	var res []*MyUtxo
	//sort.Slice(myUtxos, func(i, j int) bool {
	//	return myUtxos[i].Amount < myUtxos[j].Amount
	//})
	for _, utxo := range myUtxos {
		res = append(res, utxo)
		amount -= utxo.Amount
		if amount <= 0 {
			break
		}
	}

	return res
}

func CreateTxV2(destination string, amount int64, client *rpcclient.Client) (*wire.MsgTx, *btcutil.WIF, error) {
	defaultAddress, err := btcutil.DecodeAddress("SeTCfjeSQYevShUDEqo59GH1V5kqnP4dg5", &chaincfg.SimNetParams)
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

func CreateTx(destination string, amount int64, client *rpcclient.Client) (*wire.MsgTx, *btcutil.WIF, error) {
	defaultAddress, err := btcutil.DecodeAddress("SeTCfjeSQYevShUDEqo59GH1V5kqnP4dg5", &chaincfg.SimNetParams)
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
	for i, _ := range redeemTx.TxIn {
		signature, err := txscript.SignatureScript(redeemTx, i, pkScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return nil, err
		}

		redeemTx.TxIn[i].SignatureScript = signature
	}
	return redeemTx, nil
}
