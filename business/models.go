package business

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type NftData struct {
	ID     string
	Url    string
	Memo   string
	Binary string
}

type RevealTxInput struct {
	CommitTxHash *chainhash.Hash
	Idx          uint32
	Wif          *btcutil.WIF
	CommitOutput *wire.TxOut
	ChainConfig  *chaincfg.Params
}
