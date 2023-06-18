package business

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type NftData struct {
	ID     string `json:"id"`
	Url    string `json:"url"`
	Memo   string `json:"memo"`
	TxID   string `json:"txId"`
	Binary string `json:"binary"`
}

type RevealTxInput struct {
	CommitTxHash *chainhash.Hash
	Idx          uint32
	Wif          *btcutil.WIF
	CommitOutput *wire.TxOut
	ChainConfig  *chaincfg.Params
}
