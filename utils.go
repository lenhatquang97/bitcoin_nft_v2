package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
)

func ChunkSlice(slice []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func LoadCerts(baseFolder string) ([]byte, error) {
	certHomeDir := btcutil.AppDataDir(baseFolder, false)
	certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return nil, err
	}
	return certs, nil
}

func GetBitcoinWalletRpcClient() (*rpcclient.Client, error) {
	certs, _ := LoadCerts("btcwallet")
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         "localhost:18554",
		Endpoint:     "ws",
		User:         "youruser",
		Pass:         "SomeDecentp4ssw0rd",
		Certificates: certs,
		Params:       "simnet",
	}, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func MakeRandomKeyPair() (*btcutil.WIF, error) {
	randPriv, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}
	wif, err := btcutil.NewWIF(randPriv, &chaincfg.SimNetParams, true)
	if err != nil {
		return nil, err
	}
	return wif, nil
}
