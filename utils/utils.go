package utils

import (
	"bitcoin_nft_v2/config"
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

func GetBitcoinWalletRpcClient(certName string, networkConfig config.NetworkConfig) (*rpcclient.Client, error) {
	certs, _ := LoadCerts(certName)
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         networkConfig.Host,
		Endpoint:     networkConfig.Endpoint,
		User:         networkConfig.User,
		Pass:         networkConfig.Pass,
		Params:       networkConfig.Params,
		Certificates: certs,
	}, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}
