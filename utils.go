package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/btcsuite/btcwallet/walletdb"
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

func GetDataDir(dataDir string) string {
	path := ""
	if dataDir != "" {
		path = dataDir
	} else {
		dirname, err := os.UserConfigDir()
		if err != nil {
			return ""
		}
		path = dirname
	}

	return path + "simnet"
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

func InitializeWallet(passPhrase string, create bool) (*wallet.Wallet, error) {
	privPass := []byte("password")
	pubPass := []byte(wallet.InsecurePubPassphrase)

	basePath := btcutil.AppDataDir("btcwallet", false)
	dbPath := filepath.Join(basePath+"/simnet/", wallet.WalletDBName)
	if create {
		db, err := walletdb.Create("bdb", dbPath, true, 3*time.Second)
		if err != nil {
			return nil, err
		}
		defer db.Close()
		err = wallet.Create(db, pubPass, privPass, nil, &chaincfg.SimNetParams, time.Now())
		if err != nil {
			return nil, err
		}
	} else {
		db, err := walletdb.Open("bdb", dbPath, true, 3*time.Second)
		if err != nil {
			return nil, err
		}
		defer db.Close()
		wallet := InitWallet(dbPath)
		return wallet, nil
	}

	fmt.Println("The wallet has been created successfully.")
	return nil, nil
}

func InitWallet(dbDir string) *wallet.Wallet {
	loader := wallet.NewLoader(&chaincfg.SimNetParams, dbDir, true, 10*time.Second, 250)
	w, loaded := loader.LoadedWallet()
	fmt.Println(loaded)
	return w
}

func GetBitcoinWalletRpcClient() (*rpcclient.Client, error) {
	certs, _ := LoadCerts("btcwallet")
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         "localhost:18554",
		Endpoint:     "ws",
		User:         "youruser",
		Pass:         "SomeDecentp4ssw0rd",
		Certificates: certs,
	}, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}
