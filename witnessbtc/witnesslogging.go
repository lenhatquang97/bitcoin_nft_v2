package witnessbtc

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

func IterateWitness(client *rpcclient.Client, revealTxHash *chainhash.Hash) {
	retrievedTx, err := client.GetRawTransaction(revealTxHash)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range retrievedTx.MsgTx().TxIn {
		for _, witnessItem := range item.Witness {
			fmt.Println(witnessItem)
		}
	}
}
