package business

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

func ExecuteCommitTransaction(sv *Server, data []byte, isRef bool, amount int64, fee int64) (*chainhash.Hash, *btcutil.WIF, error) {
	commitTx, wif, err := CreateCommitTx(amount, sv.client, data, isRef, sv.Config, fee)
	if err != nil {
		return nil, nil, err
	}

	commitTxHash, err := sv.client.SendRawTransaction(commitTx, false)
	if err != nil {
		return nil, nil, err
	}
	return commitTxHash, wif, nil
}

func CreateCommitTx(amount int64, client *rpcclient.Client, embeddedData []byte, isRef bool, networkConfig *config.NetworkConfig, fee int64) (*wire.MsgTx, *btcutil.WIF, error) {
	//Step 1: Get private key
	defaultAddress, err := utils.GetDefaultAddress(client, networkConfig.SenderAddress, networkConfig.ParamsObject)
	if err != nil {
		return nil, nil, err
	}

	wif, err := client.DumpPrivKey(defaultAddress)
	if err != nil {
		return nil, nil, err
	}

	//Step 2: Get utxo and balance (all for testnet)
	utxos, err := client.ListUnspent()
	if err != nil {
		return nil, nil, err
	}

	sendUtxos := utils.GetManyUtxo(utxos, defaultAddress.EncodeAddress(), float64(amount))
	if len(sendUtxos) == 0 {
		return nil, nil, fmt.Errorf("no utxos")
	}

	balance, err := utils.GetActualBalance(client, networkConfig.SenderAddress)
	if err != nil {
		return nil, nil, err
	}

	pkScript, _ := txscript.PayToAddrScript(defaultAddress)

	if err != nil {
		return nil, nil, err
	}

	// Step 3: extracting destination address as []byte from function argument (destination string)
	hashLockScript, err := utils.CreateInscriptionScriptV2(wif.PrivKey.PubKey(), embeddedData, isRef)
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

	//Step 4: Create new transaction
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

	redeemTxOut := wire.NewTxOut(amount, outputScript)
	redeemTx.AddTxOut(redeemTxOut)

	if int64(balance) < amount+fee {
		return nil, nil, fmt.Errorf("the balance of the account is not sufficient")
	}

	// Output with satoshi change
	if int64(balance) > amount+fee {
		changeCoin := int64(balance) - amount - fee
		changeAddressScript, _ := txscript.PayToAddrScript(defaultAddress)
		fakeChangeTxOut := wire.NewTxOut(changeCoin, changeAddressScript)
		redeemTx.AddTxOut(fakeChangeTxOut)
	}

	// now sign the transaction
	finalRawTx, err := utils.SignTx(wif, pkScript, redeemTx)
	if err != nil {
		return nil, nil, err
	}

	return finalRawTx, wif, nil
}

func FakeCommitTxFee(sv *Server, dataSend []byte, amount int64, isRef bool) (int64, error) {
	fee, err := EstimateFeeForCommitTx(sv, sv.Config, amount, dataSend, isRef)
	if err != nil {
		return 0, err
	}
	return fee, nil
}

func EstimateFeeForCommitTx(sv *Server, networkConfig *config.NetworkConfig, amount int64, dataSend []byte, isRef bool) (int64, error) {
	defaultAddress, err := utils.GetDefaultAddress(sv.client, networkConfig.SenderAddress, networkConfig.ParamsObject)
	if err != nil {
		return 0, err
	}

	wif, err := sv.client.DumpPrivKey(defaultAddress)
	if err != nil {
		return 0, err
	}

	hashLockScript, err := utils.CreateInscriptionScriptV2(wif.PrivKey.PubKey(), dataSend, isRef)
	if err != nil {
		return 0, fmt.Errorf("error building script: %v", err)
	}

	outputKey, _, _ := utils.CreateOutputKeyBasedOnScript(wif.PrivKey.PubKey(), hashLockScript)
	outputScriptBuilder := txscript.NewScriptBuilder()
	outputScriptBuilder.AddOp(txscript.OP_1)
	outputScriptBuilder.AddData(schnorr.SerializePubKey(outputKey))
	outputScript, _ := outputScriptBuilder.Script()
	redeemTx, err := utils.NewTx()
	if err != nil {
		return 0, err
	}
	utxos, err := sv.client.ListUnspent()
	if err != nil {
		return 0, err
	}

	sendUtxos := utils.GetManyUtxo(utxos, defaultAddress.EncodeAddress(), float64(amount))
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

	smartFeeRate, err := sv.client.EstimateFee(1)
	if err != nil {
		return 0, err
	}
	smartFeeRate = smartFeeRate * 100_000
	txSize := int64(redeemTx.SerializeSize())

	fee := int64(smartFeeRate * float64(txSize) * 2)
	return fee, nil
}
