package business

import (
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

/*
Reveal transaction
*/

func ExecuteRevealTransaction(client *rpcclient.Client, revealTxInput *RevealTxInput, data []byte, isRef bool, toAddress string, fee int64, amount int64) (*chainhash.Hash, error) {
	revealTx, err := RevealTx(client, data, isRef, *revealTxInput.CommitTxHash, *revealTxInput.CommitOutput, revealTxInput.Idx, revealTxInput.Wif.PrivKey, revealTxInput.ChainConfig, toAddress, fee, amount)
	if err != nil {
		return nil, err
	}

	revealTxHash, err := client.SendRawTransaction(revealTx, true)
	if err != nil {
		return nil, err
	}
	return revealTxHash, nil
}

func RevealTx(client *rpcclient.Client, embeddedData []byte, isRef bool, commitTxHash chainhash.Hash, commitOutput wire.TxOut, txOutIndex uint32, randPriv *btcec.PrivateKey, params *chaincfg.Params, toAddress string, fee int64, amount int64) (*wire.MsgTx, error) {
	tx, outputScript, sigHashes, inputFetcher, err := CreateRevealTxObj(client, embeddedData, isRef, commitTxHash, commitOutput, txOutIndex, randPriv, params, toAddress, amount)
	if err != nil {
		return nil, err
	}

	err = ValidateRevealTx(outputScript, tx, sigHashes, inputFetcher)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func FakeRevealTxFee(sv *Server, dataSend []byte, isRef bool, toAddress string, amount int64) (int64, error) {
	fakeCommitTxHash, err := chainhash.NewHashFromStr("932012f4b18bad5f1e8ece085bac68dae6b8213b58cdf6a38f52752df81d0663")
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	toAddressObj, _ := btcutil.DecodeAddress(toAddress, sv.Config.ParamsObject)
	fakeOutputScript, _ := txscript.PayToAddrScript(toAddressObj)
	fakeCommitOutput := wire.NewTxOut(0, fakeOutputScript)
	randPriv, err := btcec.NewPrivateKey()
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	estimatedFee, err := EstimatedFeeForRevealTx(sv.client, dataSend, isRef, *fakeCommitTxHash, *fakeCommitOutput, 0, randPriv, sv.Config.ParamsObject, toAddress, amount)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return estimatedFee, nil
}

func EstimatedFeeForRevealTx(client *rpcclient.Client, embeddedData []byte, isRef bool, commitTxHash chainhash.Hash, commitOutput wire.TxOut, txOutIndex uint32, randPriv *btcec.PrivateKey, params *chaincfg.Params, toAddress string, amount int64) (int64, error) {
	tx, _, _, _, err := CreateRevealTxObj(client, embeddedData, isRef, commitTxHash, commitOutput, txOutIndex, randPriv, params, toAddress, amount)
	if err != nil {
		return 0, err
	}

	smartFeeRate, err := client.EstimateFee(1)
	if err != nil {
		return 0, err
	}

	smartFeeRate = smartFeeRate * 100_000
	fmt.Println("smartFeeRate: ", smartFeeRate)
	fmt.Println("tx.SerializeSize(): ", tx.SerializeSize())
	fee := int64(smartFeeRate * float64(tx.SerializeSize()) * 2)
	if err != nil {
		return 0, err
	}
	return fee, nil
}

/*
For support in reveal tx
*/

func CreateRevealTxObj(client *rpcclient.Client, embeddedData []byte, isRef bool, commitTxHash chainhash.Hash, commitOutput wire.TxOut, txOutIndex uint32, randPriv *btcec.PrivateKey, params *chaincfg.Params, toAddress string, amount int64) (*wire.MsgTx, []byte, *txscript.TxSigHashes, txscript.PrevOutputFetcher, error) {
	pubKey := randPriv.PubKey()
	pkScript, err := utils.CreateInscriptionScriptV2(pubKey, embeddedData, isRef)

	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error building script: %v", err)
	}
	outputKey, tapScriptTree, tapLeaf := utils.CreateOutputKeyBasedOnScript(pubKey, pkScript)

	outputScriptBuilder := txscript.NewScriptBuilder()
	outputScriptBuilder.AddOp(txscript.OP_1)
	outputScriptBuilder.AddData(schnorr.SerializePubKey(outputKey))
	outputScript, _ := outputScriptBuilder.Script()

	address, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), params)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	fmt.Println(address.EncodeAddress())

	ctrlBlock := tapScriptTree.LeafMerkleProofs[0].ToControlBlock(pubKey)

	tx := wire.NewMsgTx(2)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  commitTxHash,
			Index: txOutIndex,
		},
	})

	customAddress, err := btcutil.DecodeAddress(toAddress, params)
	if err != nil {
		fmt.Println("Decode address error", err)
		return nil, nil, nil, nil, err
	}

	customAddrScript, err := txscript.PayToAddrScript(customAddress)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error creating op return script: %v", err)
	}

	//TODO: amount is right?
	txOut := &wire.TxOut{
		Value: amount, PkScript: customAddrScript,
	}
	tx.AddTxOut(txOut)

	inputFetcher := txscript.NewCannedPrevOutputFetcher(
		commitOutput.PkScript,
		commitOutput.Value,
	)
	sigHashes := txscript.NewTxSigHashes(tx, inputFetcher)

	sig, err := txscript.RawTxInTapscriptSignature(
		tx, sigHashes, 0, txOut.Value,
		txOut.PkScript, *tapLeaf, txscript.SigHashDefault,
		randPriv,
	)

	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error signing tapscript: %v", err)
	}

	// Now that we have the sig, we'll make a valid witness
	// including the control block.
	ctrlBlockBytes, err := ctrlBlock.ToBytes()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error including control block: %v", err)
	}
	tx.TxIn[0].Witness = wire.TxWitness{
		sig, pkScript, ctrlBlockBytes,
	}
	return tx, outputScript, sigHashes, inputFetcher, nil
}

func ValidateRevealTx(outputScript []byte, tx *wire.MsgTx, sigHashes *txscript.TxSigHashes, inputFetcher txscript.PrevOutputFetcher) error {
	fmt.Println("==================================Validation=============================================")
	engine, err := txscript.NewEngine(outputScript, tx, 0, txscript.StandardVerifyFlags, nil, sigHashes, 8000, inputFetcher)
	if err != nil {
		return err
	}
	err = engine.Execute()
	if err != nil {
		return err
	}
	fmt.Println("===================================Success validation====================================")
	return nil
}
