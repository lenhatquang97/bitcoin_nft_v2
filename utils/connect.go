package utils

import (
	"bitcoin_nft_v2/config"

	"github.com/btcsuite/btcd/rpcclient"
)

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
