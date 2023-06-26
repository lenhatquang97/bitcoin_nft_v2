package business

import (
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/mempool"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func EstimateFeeForCommitTx(sv *Server, amount int64, dataSend []byte, isRef bool) (int64, error) {
	defaultAddress, err := sv.client.GetAccountAddress("default")
	if err != nil {
		return 0, err
	}

	wif, err := sv.client.DumpPrivKey(defaultAddress)
	if err != nil {
		return 0, err
	}

	hashLockScript, err := utils.CreateInscriptionScriptV2(wif.PrivKey.PubKey(), dataSend, isRef, ON_CHAIN)
	if err != nil {
		return 0, fmt.Errorf("error building script: %v", err)
	}

	outputKey, _, _ := utils.CreateOutputKeyBasedOnScript(wif.PrivKey.PubKey(), hashLockScript)
	outputScriptBuilder := txscript.NewScriptBuilder()
	outputScriptBuilder.AddOp(txscript.OP_1)
	outputScriptBuilder.AddData(schnorr.SerializePubKey(outputKey))
	outputScript, _ := outputScriptBuilder.Script()
	redeemTx := wire.NewMsgTx(wire.TxVersion)
	utxos, err := sv.client.ListUnspent()
	if err != nil {
		return 0, err
	}

	sendUtxos := utils.GetManyUtxo(sv.client, utxos, float64(amount), "")
	for _, utxo := range sendUtxos {
		utxoHash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return 0, err
		}

		outPoint := wire.NewOutPoint(utxoHash, utxo.Vout)

		// making the input, and adding it to transaction
		txIn := wire.NewTxIn(outPoint, nil, nil)
		redeemTx.AddTxIn(txIn)
	}

	redeemTxOut := wire.NewTxOut(amount, outputScript)
	redeemTx.AddTxOut(redeemTxOut)

	//Fake change money to estimate transaction fee
	changeAddressScript, _ := txscript.PayToAddrScript(defaultAddress)
	fakeChangeTxOut := wire.NewTxOut(100, changeAddressScript)
	redeemTx.AddTxOut(fakeChangeTxOut)

	//txSize := int64(tx.SerializeSize())
	feeRate, err := sv.client.EstimateFee(1)
	if err != nil {
		return 0, err
	}
	txSize := mempool.GetTxVirtualSize(btcutil.NewTx(redeemTx))
	fee := txSize * int64(feeRate*100_000) * 3
	return fee, nil
}

func EstimatedFeeForRevealTx(client *rpcclient.Client, embeddedData []byte, isRef bool, commitTxHash chainhash.Hash, commitOutput wire.TxOut, txOutIndex uint32, randPriv *btcec.PrivateKey, params *chaincfg.Params, toAddress string, amount int64) (int64, error) {
	tx, _, _, _, err := CreateRevealTxObj(client, embeddedData, isRef, commitTxHash, commitOutput, txOutIndex, randPriv, params, toAddress, amount, "on_chain")
	if err != nil {
		return 0, err
	}

	//txSize := int64(tx.SerializeSize())
	feeRate, err := client.EstimateFee(1)
	if err != nil {
		return 0, err
	}
	txSize := mempool.GetTxVirtualSize(btcutil.NewTx(tx))
	fee := txSize * int64(feeRate*100_000)

	if err != nil {
		return 0, err
	}
	return fee, nil
}
