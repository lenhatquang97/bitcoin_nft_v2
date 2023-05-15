package main

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcwallet/waddrmgr"
)

const (
	PubKeyFormatCompressedOdd byte = 0x03
)

func TapscriptFullTree(internalKey *btcec.PublicKey, allTreeLeaves ...txscript.TapLeaf) *waddrmgr.Tapscript {
	tree := txscript.AssembleTaprootScriptTree(allTreeLeaves...)
	rootHash := tree.RootNode.TapHash()
	tapKey := txscript.ComputeTaprootOutputKey(internalKey, rootHash[:])

	var outputKeyYIsOdd bool
	if tapKey.SerializeCompressed()[0] == PubKeyFormatCompressedOdd {
		outputKeyYIsOdd = true
	}

	return &waddrmgr.Tapscript{
		Type: waddrmgr.TapscriptTypeFullTree,
		ControlBlock: &txscript.ControlBlock{
			InternalKey:     internalKey,
			OutputKeyYIsOdd: outputKeyYIsOdd,
			LeafVersion:     txscript.BaseLeafVersion,
		},
		Leaves: allTreeLeaves,
	}
}
