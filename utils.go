package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
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

func GetBitcoinRPCClient() (*rpcclient.Client, error) {
	certs, err := LoadCerts("btcd")
	if err != nil {
		return nil, err
	}

	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         "localhost:8334",
		Endpoint:     "ws",
		User:         "4bmeiF7E3ny8cGf8Ok6QJZy/0pk=",
		Pass:         "2oljjSoRFzC5Go7hCGDID6xWi+c=",
		Certificates: certs,
	}, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
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
