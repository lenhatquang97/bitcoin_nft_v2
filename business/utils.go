package business

import (
	"bitcoin_nft_v2/config"
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/utils"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"io"
	"os"
)

func ComputeNftDataByte(data *NftData) ([]byte, [32]byte) {
	h := sha256.New()
	_, _ = h.Write([]byte(data.ID))
	_, _ = h.Write([]byte(data.Url))
	_, _ = h.Write([]byte(data.Memo))

	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, [32]byte{}
	}

	return rawData, *(*[32]byte)(h.Sum(nil))
}

func WrapperError(errStr string) error {
	return errors.New(errStr)
}

func NewRootHashForReceiver(nftData []*NftData) ([]byte, error) {
	tree := nft_tree.NewCompactedTree(nft_tree.NewDefaultStore())

	var updatedRoot *nft_tree.BranchNode
	for _, item := range nftData {
		// Compute Nft Data Info
		dataByte, key := ComputeNftDataByte(item)

		// Init Root Hash For Receiver
		leaf := nft_tree.NewLeafNode(dataByte, 0) // CoinsToSend
		leaf.NodeHash()

		updatedTree, err := tree.Insert(context.TODO(), key, leaf)
		if err != nil {
			return nil, err
		}

		updatedRoot, err = updatedTree.Root(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return utils.GetNftRoot(updatedRoot), nil
}

func ExecuteRevealTransaction(client *rpcclient.Client, revealTxInput *RevealTxInput, data []byte, toAddress string) (*chainhash.Hash, error) {
	revealTx, _, err := RevealTx(data, *revealTxInput.CommitTxHash, *revealTxInput.CommitOutput, revealTxInput.Idx, revealTxInput.Wif.PrivKey, revealTxInput.ChainConfig, toAddress)
	if err != nil {
		return nil, err
	}

	revealTxHash, err := client.SendRawTransaction(revealTx, true)
	if err != nil {
		return nil, err
	}
	return revealTxHash, nil
}

func RevealTx(embeddedData []byte, commitTxHash chainhash.Hash, commitOutput wire.TxOut, txOutIndex uint32, randPriv *btcec.PrivateKey, params *chaincfg.Params, toAddress string) (*wire.MsgTx, *btcutil.AddressTaproot, error) {
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

	customAddress, err := btcutil.DecodeAddress(toAddress, params)
	if err != nil {
		fmt.Println("Decode address error", err)
		return nil, nil, err
	}

	opReturnScript, err := txscript.PayToAddrScript(customAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating op return script: %v", err)
	}

	txOut := &wire.TxOut{
		Value: 0, PkScript: opReturnScript,
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

func ExecuteCommitTransaction(client *rpcclient.Client, data []byte, netConfig *config.NetworkConfig, amount int64) (*chainhash.Hash, *btcutil.WIF, error) {
	commitTx, wif, err := CreateCommitTx(amount, client, data, netConfig)
	if err != nil {
		return nil, nil, err
	}

	commitTxHash, err := client.SendRawTransaction(commitTx, false)
	if err != nil {
		return nil, nil, err
	}
	return commitTxHash, wif, nil
}

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

	balance, err := utils.GetActualBalance(client, networkConfig.SenderAddress)
	if err != nil {
		return nil, nil, err
	}

	pkScript, _ := txscript.PayToAddrScript(defaultAddress)

	if err != nil {
		return nil, nil, err
	}

	// checking for sufficiency of account
	if networkConfig.Params == "testnet3" && int64(float64(balance)*float64(TESTNET_1_BTC)) < amount+DefaultFee {
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
	if networkConfig.Params == "testnet3" && int64(float64(balance)*float64(TESTNET_1_BTC)) > amount+DefaultFee {
		changeCoin = int64(float64(balance)*float64(TESTNET_1_BTC)) - amount - DefaultFee
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

func FileSha256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new SHA256 hash
	hash := sha256.New()

	// Copy the file contents to the hash calculator
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Get the hash sum as a byte slice
	hashSum := hash.Sum(nil)
	return hex.EncodeToString(hashSum), nil

}
