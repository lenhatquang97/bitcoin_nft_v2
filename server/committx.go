package server

import (
	"bitcoin_nft_v2/config"
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func CreateCommitTx(amount int64, client *rpcclient.Client, embeddedData []byte, networkConfig *config.NetworkConfig) (*wire.MsgTx, *btcutil.WIF, error) {
	defaultAddress, err := utils.GetDefaultAddress(client, networkConfig.SenderAddress, networkConfig.ParamsObject)
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

	sendUtxos := utils.GetManyUtxo(utxos, defaultAddress.EncodeAddress(), float64(amount))
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
	if networkConfig.Params == "testnet3" && int64(balance*float64(TESTNET_1_BTC)) < amount+DefaultFee {
		return nil, nil, fmt.Errorf("the balance of the account is not sufficient")
	} else if networkConfig.Params == "simnet" && int64(balance) < amount+DefaultFee {
		return nil, nil, fmt.Errorf("the balance of the account is not sufficient")
	}

	// extracting destination address as []byte from function argument (destination string)
	hashLockScript, err := utils.CreateInscriptionScript(wif.PrivKey.PubKey(), embeddedData)
	if err != nil {
		return nil, nil, fmt.Errorf("error building script: %v", err)
	}
	outputKey, _, _ := utils.CreateOutputKeyBasedOnScript(wif.PrivKey.PubKey(), hashLockScript)

	address, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), networkConfig.ParamsObject)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(address.EncodeAddress())

	outputScriptBuilder := txscript.NewScriptBuilder()
	outputScriptBuilder.AddOp(txscript.OP_1)
	outputScriptBuilder.AddData(schnorr.SerializePubKey(outputKey))
	outputScript, _ := outputScriptBuilder.Script()

	redeemTx, err := utils.NewTx()
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

	fmt.Println(outputScript)

	redeemTxOut := wire.NewTxOut(amount, outputScript)
	redeemTx.AddTxOut(redeemTxOut)

	var changeCoin int64
	if networkConfig.Params == "testnet3" && int64(balance*float64(TESTNET_1_BTC)) > amount+DefaultFee {
		changeCoin = int64(balance*float64(TESTNET_1_BTC)) - amount - DefaultFee
	} else if networkConfig.Params == "simnet" && int64(balance) > amount+DefaultFee {
		changeCoin = int64(balance) - amount - DefaultFee
	}

	if changeCoin > 0 {
		changeAddressScript, _ := txscript.PayToAddrScript(defaultAddress)
		rawChangeTxOut := wire.NewTxOut(changeCoin, changeAddressScript)
		redeemTx.AddTxOut(rawChangeTxOut)
	}

	// now sign the transaction
	finalRawTx, err := utils.SignTx(wif, pkScript, redeemTx)
	if err != nil {
		return nil, nil, err
	}

	return finalRawTx, wif, nil
}

func ExecuteCommitTransaction(client *rpcclient.Client, data []byte, netConfig *config.NetworkConfig) (*chainhash.Hash, *btcutil.WIF, error) {
	commitTx, wif, err := CreateCommitTx(CoinsToSend, client, data, netConfig)
	if err != nil {
		return nil, nil, err
	}

	commitTxHash, err := client.SendRawTransaction(commitTx, false)
	if err != nil {
		return nil, nil, err
	}
	return commitTxHash, wif, nil
}
