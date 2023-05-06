package main

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

func GetPayToAddrScript(address string) []byte {
	rcvAddress, _ := btcutil.DecodeAddress(address, &chaincfg.SimNetParams)
	rcvScript, _ := txscript.PayToAddrScript(rcvAddress)
	return rcvScript
}

func GetPrivateKey(privKey string) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	privByte, err := hex.DecodeString(privKey)

	if err != nil {
		return nil, nil, err
	}

	priv, pubKey := btcec.PrivKeyFromBytes(privByte) //secp256k1
	return priv, pubKey, nil
}
