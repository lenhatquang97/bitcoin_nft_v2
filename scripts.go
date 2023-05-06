package main

import (
	"encoding/hex"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

func GetPayToAddrScript(address string) []byte {
	rcvAddress, _ := btcutil.DecodeAddress(address, &chaincfg.TestNet3Params)
	rcvScript, _ := txscript.PayToAddrScript(rcvAddress)
	return rcvScript
}

func GetKeyAddressFromPrivateKey(privKey string) (*btcec.PrivateKey, string) {
	privByte, err := hex.DecodeString(privKey)

	if err != nil {
		log.Panic(err)
	}

	priv, pubKey := btcec.PrivKeyFromBytes(privByte) //secp256k1

	address, _ := btcutil.NewAddressPubKeyHash(
		btcutil.Hash160(pubKey.SerializeUncompressed()), &chaincfg.TestNet3Params)

	return priv, address.EncodeAddress()
}
