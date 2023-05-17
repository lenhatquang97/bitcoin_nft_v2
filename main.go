package main

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

func main() {
	DoCommitRevealTransaction()
}

func ExecuteCommitTransaction(client *rpcclient.Client) (*chainhash.Hash, *btcutil.WIF, error) {
	commitTx, wif, err := CreateCommitTx(CoinsToSend, client, EmbeddedData, &TestNetConfig)
	if err != nil {
		return nil, nil, err
	}

	commitTxHash, err := client.SendRawTransaction(commitTx, false)
	if err != nil {
		return nil, nil, err
	}
	return commitTxHash, wif, nil
}
func ExecuteRevealTransaction(client *rpcclient.Client, commitTxHash *chainhash.Hash, idx uint32, wif *btcutil.WIF, commitOutput *wire.TxOut, config *chaincfg.Params) (*chainhash.Hash, error) {
	revealTx, _, err := RevealTx(EmbeddedData, *commitTxHash, *commitOutput, idx, wif.PrivKey, config)
	if err != nil {
		return nil, err
	}

	revealTxHash, err := client.SendRawTransaction(revealTx, false)
	if err != nil {
		return nil, err
	}
	return revealTxHash, nil
}
