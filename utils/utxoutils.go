package utils

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

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
	for _, utxo := range myUtxos {
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
			amount += int(utxos[i].Amount)
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
	simNetAddress, err := btcutil.DecodeAddress(senderAddress, config)
	if err != nil {
		return nil, err
	}
	return simNetAddress, nil
}
