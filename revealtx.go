package main

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

func RevealTx(embeddedData []byte, commitTxHash chainhash.Hash, commitOutput wire.TxOut, txOutIndex uint32, randPriv *btcec.PrivateKey, params *chaincfg.Params) (*wire.MsgTx, *btcutil.AddressTaproot, error) {
	pubKey := randPriv.PubKey()
	pkScript, err := utils.CreateInscriptionScript(pubKey, embeddedData)

	if err != nil {
		return nil, nil, fmt.Errorf("error building script: %v", err)
	}
	outputKey, tapScriptTree, tapLeaf := utils.CreateOutputKeyBasedOnScript(pubKey, pkScript)

	outputScriptBuilder := txscript.NewScriptBuilder()
	outputScriptBuilder.AddOp(txscript.OP_1)
	outputScriptBuilder.AddData(schnorr.SerializePubKey(outputKey))
	outputScript, _ := outputScriptBuilder.Script()

	address, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(outputKey), params)
	if err != nil {
		return nil, nil, err
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

	// opReturnScript, err := txscript.NullDataScript([]byte("https://example.com"))
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("error creating op return script: %v", err)
	// }

	anotherAddress, _ := btcutil.DecodeAddress("mv4rnyY3Su5gjcDNzbMLKBQkBicCtHUtFB", &chaincfg.TestNet3Params)
	anotherAddressScript, _ := txscript.PayToAddrScript(anotherAddress)

	txOut := &wire.TxOut{
		Value: 1000, PkScript: anotherAddressScript,
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
		return nil, nil, fmt.Errorf("error signing tapscript: %v", err)
	}

	// Now that we have the sig, we'll make a valid witness
	// including the control block.
	ctrlBlockBytes, err := ctrlBlock.ToBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("error including control block: %v", err)
	}
	tx.TxIn[0].Witness = wire.TxWitness{
		sig, pkScript, ctrlBlockBytes,
	}

	fmt.Println("==================================Validation=============================================")
	engine, err := txscript.NewEngine(outputScript, tx, 0, txscript.StandardVerifyFlags, nil, sigHashes, 8000, inputFetcher)
	if err != nil {
		return nil, nil, err
	}
	err = engine.Execute()
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("===================================Success validation====================================")
	return tx, address, nil
}

type RevealTxInput struct {
	CommitTxHash *chainhash.Hash
	Idx          uint32
	Wif          *btcutil.WIF
	CommitOutput *wire.TxOut
	ChainConfig  *chaincfg.Params
}

func ExecuteRevealTransaction(client *rpcclient.Client, revealTxInput *RevealTxInput, data []byte) (*chainhash.Hash, error) {
	revealTx, _, err := RevealTx(data, *revealTxInput.CommitTxHash, *revealTxInput.CommitOutput, revealTxInput.Idx, revealTxInput.Wif.PrivKey, revealTxInput.ChainConfig)
	if err != nil {
		return nil, err
	}

	revealTxHash, err := client.SendRawTransaction(revealTx, true)
	if err != nil {
		return nil, err
	}
	return revealTxHash, nil
}
