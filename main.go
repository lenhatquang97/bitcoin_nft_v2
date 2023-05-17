package main

import (
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
)

func main() {
	client, err := utils.GetBitcoinWalletRpcClient("btcwallet", TestNetConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.WalletPassphrase(PassphraseInWallet, PassphraseTimeout)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("===================================Checkpoint 0====================================")

	commitTxHash, wif, err := ExecuteCommitTransaction(client)
	if err != nil {
		fmt.Println(err)
	}

	retrievedCommitTx, err := client.GetRawTransaction(commitTxHash)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("===================================Checkpoint 1====================================")

	revealTxHash, err := ExecuteRevealTransaction(client, commitTxHash, 0, wif, retrievedCommitTx.MsgTx().TxOut[0], TestNetConfig.ParamsObject)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("===================================Checkpoint 2====================================")
	fmt.Println(revealTxHash)
	IterateWitness(client, revealTxHash)
	fmt.Println("===================================Success====================================")

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
