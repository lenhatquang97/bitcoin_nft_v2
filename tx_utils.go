package main

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func SignTx(wif *btcutil.WIF, pkScript []byte, redeemTx *wire.MsgTx) (*wire.MsgTx, error) {
	for i, _ := range redeemTx.TxIn {
		signature, err := txscript.SignatureScript(redeemTx, i, pkScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return nil, err
		}

		redeemTx.TxIn[i].SignatureScript = signature
	}
	return redeemTx, nil
}
