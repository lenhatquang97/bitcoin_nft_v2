package utils

import "fmt"

func PrintLogUtxos(sendUtxos []*MyUtxo) {
	for _, item := range sendUtxos {
		fmt.Printf("%s:%d with %f\n", item.TxID, item.Vout, item.Amount)
	}
}
