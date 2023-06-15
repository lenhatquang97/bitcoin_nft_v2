package utils

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

type MyUtxo struct {
	TxID    string
	Vout    uint32
	Amount  float64
	Address string
}

func GetManyUtxo(client *rpcclient.Client, utxos []btcjson.ListUnspentResult, amount float64, specialTxId string) []*MyUtxo {
	var myUtxos []*MyUtxo
	for i := 0; i < len(utxos); i++ {

		myUtxos = append(myUtxos, &MyUtxo{
			TxID:    utxos[i].TxID,
			Vout:    utxos[i].Vout,
			Amount:  utxos[i].Amount,
			Address: utxos[i].Address,
		})
	}
	var res []*MyUtxo
	// fistly choose utxo with tx id
	for _, utxo := range myUtxos {
		if utxo.TxID == specialTxId {
			res = append(res, utxo)
			amount -= utxo.Amount
			break
		}
	}

	if amount <= 0 {
		return res
	}

	for _, utxo := range myUtxos {
		isNft, _ := CheckTransactionHasNft(client, utxo.TxID)
		if utxo.TxID == specialTxId || isNft {
			continue
		}

		res = append(res, utxo)
		amount -= utxo.Amount
		if amount <= 0 {
			break
		}
	}

	return res
}

func GetActualBalance(client *rpcclient.Client, actualAddress string) (int, error) {
	utxos, err := client.ListUnspent()
	if err != nil {
		return -1, err
	}
	amount := 0

	for i := 0; i < len(utxos); i++ {
		if utxos[i].Address == actualAddress {
			//Note: This is balance for testnet
			amount += int(utxos[i].Amount * 100_000_000)
		}
	}
	return amount, nil
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
func GetDefaultAddress(client *rpcclient.Client, senderAddress string, config *chaincfg.Params) (btcutil.Address, error) {
	if len(senderAddress) == 0 {
		testNetAddress, err := client.GetAccountAddress("default")
		if err != nil {
			return nil, err
		}
		return testNetAddress, nil
	}
	customAddress, err := btcutil.DecodeAddress(senderAddress, config)
	if err != nil {
		return nil, err
	}
	return customAddress, nil
}

func CheckTransactionHasNft(client *rpcclient.Client, txId string) (bool, error) {
	hashId, err := chainhash.NewHashFromStr(txId)
	if err != nil {
		return false, err
	}

	tx, err := client.GetRawTransaction(hashId)
	if err != nil {
		return false, err
	}
	witness := tx.MsgTx().TxIn[0].Witness
	if len(witness) != 3 {
		return false, nil
	}

	maxLen := 100
	script := witness[1]

	//check whether script contains []byte("m25start"), if yes, it means this transaction has nft and return true
	if len(script) < maxLen {
		maxLen = len(script)
	}
	for i := 0; i <= maxLen-8; i++ {
		if string(script[i:i+8]) == "m25start" {
			return true, nil
		}
	}
	return false, nil
}
