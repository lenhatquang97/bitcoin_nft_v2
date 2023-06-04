package business

import (
	"bitcoin_nft_v2/config"
	"bitcoin_nft_v2/db"
	"bitcoin_nft_v2/db/sqlc"
	"bitcoin_nft_v2/gobcy"
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"io"
	"os"
	"os/exec"
	"time"
)

const (
	DefaultNameSpace   = "default"
	PassphraseInWallet = "12345"
	PassphraseTimeout  = 3
	ON_CHAIN           = "on_chain"
	OFF_CHAIN          = "off_chain"
)

type ServerOffChain struct {
	client *rpcclient.Client
	mode   string
	Config *config.NetworkConfig
	DB     *db.PostgresStore
}

func NewServer(networkCfg *config.NetworkConfig, mode string) (*ServerOffChain, error) {
	client, err := utils.GetBitcoinWalletRpcClient("btcwallet", networkCfg)
	if err != nil {
		return nil, err
	}

	var store *db.PostgresStore
	if mode == OFF_CHAIN {
		sqlFixture := db.NewTestPgFixture()
		store, err = db.NewPostgresStore(sqlFixture.GetConfig())
		if err != nil {
			return nil, err
		}
	}

	return &ServerOffChain{
		client: client,
		mode:   mode,
		Config: networkCfg,
		DB:     store,
	}, nil
}

// Send
// if on-chain mode data is file path
// else if off-chain mode data is list nft data (list by get data from db)
// if don't have data in DB --> import nft
func (sv *ServerOffChain) Send(toAddress string, amount int64, data interface{}, passphrase string) error {
	//nftUrls := []string{
	//	"https://genk.mediacdn.vn/k:thumb_w/640/2016/photo-1-1473821552147/top6suthatcucsocvepikachu.jpg",
	//	"https://pianofingers.vn/wp-content/uploads/2020/12/organ-casio-ct-s100-1.jpg",
	//	"https://amnhacvietthanh.vn/wp-content/uploads/2020/10/Yamaha-C40.jpg",
	//}
	//nameSpace := DefaultNameSpace

	// Get Nft Data
	var dataSend []byte
	var err error
	if sv.mode == OFF_CHAIN {
		var nftData []*NftData
		fmt.Println(data)
		item := data.([]string)[0]
		fmt.Println("Item test", item)
		for _, url := range data.([]string) {
			item, err := sv.DB.GetNFtDataByUrl(context.Background(), url)
			if err != nil {
				print("Get Nft Data Failed")
				fmt.Println(err)
				return err
			}

			nftData = append(nftData, &NftData{
				ID:   item.ID,
				Url:  item.Url,
				Memo: item.Memo,
			})
		}

		dataSend, err = NewRootHashForReceiver(nftData)
		if err != nil {
			fmt.Println("Compute root hash for receiver error")
			fmt.Println(err)
			return err
		}
	} else {
		var customData string
		customData, err = FileSha256(data.(string))
		if err != nil {
			// log error
			return err
		}

		dataSend = []byte(customData)
	}

	// root hash for sender
	//rootHashForSender, err := sv.PreComputeRootHashForSender(context.Background(), key, leaf, nameSpace)
	//if err != nil {
	//	fmt.Println("Compute root hash for sender error")
	//	fmt.Println(err)
	//	return
	//}
	//
	//fmt.Println("Sender root hash update is: ", rootHashForSender)

	err = sv.client.WalletPassphrase(passphrase, PassphraseTimeout)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("===================================Checkpoint 0====================================")

	//customData, err := offchainnft.FileSha256("./README.md")
	if err != nil {
		fmt.Println(err)
		return err
	}

	commitTxHash, wif, err := ExecuteCommitTransaction(sv.client, dataSend, sv.Config, amount)
	if err != nil {
		fmt.Println("commitLog")
		fmt.Println(err)
		return err
	}

	fmt.Printf("Your commit tx hash is: %s\n", commitTxHash.String())

	retrievedCommitTx, err := sv.client.GetRawTransaction(commitTxHash)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("===================================Checkpoint 1====================================")

	revealTxInput := RevealTxInput{
		CommitTxHash: commitTxHash,
		Idx:          0,
		Wif:          wif,
		CommitOutput: retrievedCommitTx.MsgTx().TxOut[0],
		ChainConfig:  sv.Config.ParamsObject,
	}

	revealTxHash, err := ExecuteRevealTransaction(sv.client, &revealTxInput, dataSend, toAddress)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//txCreator := func(tx *sql.Tx) db.TreeStore {
	//	return sv.PostgresDB.WithTx(tx)
	//}
	//
	//treeDB := db.NewTransactionExecutor[db.TreeStore](sv.PostgresDB, txCreator)
	//
	//taroTreeStore := db.NewTaroTreeStore(treeDB, DefaultNameSpace)
	//
	//tree := nft_tree.NewFullTree(taroTreeStore)
	//
	//_, err = tree.Delete(context.Background(), key)
	//if err != nil {
	//	fmt.Println("Delete leaf after reveal tx failed", err)
	//}
	// We use the default, in-memory store that doesn't actually use the
	// context.
	//updatedTree, err := tree.Insert(context.Background(), key, leaf)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//updatedRoot, err := updatedTree.Root(context.Background())
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	//rootHash := utils.GetNftRoot(updatedRoot)
	//EmbeddedData = rootHash
	fmt.Println("===================================Checkpoint 2====================================")
	fmt.Printf("Your reveal tx hash is: %s\n", revealTxHash.String())
	fmt.Println("===================================Success====================================")
	return nil
}

func (sv *ServerOffChain) CheckBalance(address string) (int, error) {
	utxos, err := sv.client.ListUnspent()
	if err != nil {
		return -1, err
	}
	amount := 0

	for i := 0; i < len(utxos); i++ {
		if utxos[i].Address == address {
			amount += int(utxos[i].Amount)
		}
	}
	return amount, nil
}

func (sv *ServerOffChain) ViewNftData() ([]*NftData, error) {
	// get nft data from db
	if sv.mode != OFF_CHAIN {
		return nil, errors.New("SERVER_MODE_IS_ON_CHAIN")
	}

	nftDatas, err := sv.DB.GetAllNft(context.Background())
	if err != nil {
		fmt.Println("[ViewNftData] Get nft data error ", err)
		fmt.Println(err)
		return nil, err
	}

	var res []*NftData

	for _, item := range nftDatas {
		res = append(res, &NftData{
			ID:   item.ID,
			Url:  item.Url,
			Memo: item.Memo,
		})
	}

	return res, nil
}

func (sv *ServerOffChain) CreateWallet(name string, passphrase string) error {
	//res, err := sv.client.CreateWallet(name, rpcclient.WithCreateWalletPassphrase(passphrase))
	//if err != nil {
	//	return err
	//}

	app := "btcwallet"

	arg0 := "--simnet"
	arg1 := "--username=" + sv.Config.User
	arg2 := "--password=" + sv.Config.Pass
	arg3 := "--create"

	cmd := exec.Command(app, arg0, arg1, arg2, arg3)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	go func() {
		time.Sleep(1 * time.Second)
		io.WriteString(os.Stdin, "12345\n")
		time.Sleep(1 * time.Second)

		io.WriteString(os.Stdin, "12345\n")
		time.Sleep(1 * time.Second)

		io.WriteString(os.Stdin, "n\n")
		time.Sleep(1 * time.Second)

		io.WriteString(os.Stdin, "n\n")
		time.Sleep(1 * time.Second)

		io.WriteString(os.Stdin, "OK\n")

		//time.Sleep(5 * time.Second)
		//io.WriteString(os.Stdin, "echo hello again\n")
	}()

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error run ", err)
		return err
	}

	return nil
}

func (sv *ServerOffChain) GetNftData() {

}

func (sv *ServerOffChain) ImportProof(id, url, memo string) error {
	if sv.mode != OFF_CHAIN {
		return errors.New("SERVER_MODE_IS_ON_CHAIN")
	}

	// import nft data and merge tree
	dataByte, key := ComputeNftDataByte(&NftData{
		ID:   id,
		Url:  url,
		Memo: memo,
	})

	// Init Root Hash For Receiver
	leaf := nft_tree.NewLeafNode(dataByte, 0) // CoinsToSend
	leaf.NodeHash()

	txCreator := func(tx *sql.Tx) db.TreeStore {
		return sv.DB.WithTx(tx)
	}

	treeDB := db.NewTransactionExecutor[db.TreeStore](sv.DB, txCreator)

	taroTreeStore := db.NewTaroTreeStore(treeDB, DefaultNameSpace)

	tree := nft_tree.NewFullTree(taroTreeStore)

	//We use the default, in-memory store that doesn't actually use the
	//context.
	updatedTree, err := tree.Insert(context.Background(), key, leaf)
	if err != nil {
		fmt.Println(err)
		return err
	}

	updatedRoot, err := updatedTree.Root(context.Background())
	if err != nil {
		fmt.Println(err)
		return err
	}

	rootHash := utils.GetNftRoot(updatedRoot)
	err = sv.DB.InsertNftData(context.Background(), sqlc.InsertNftDataParams{
		ID:   id,
		Url:  url,
		Memo: memo,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("New Root Hash: ", rootHash)

	return err
}

func (sv *ServerOffChain) ExportProof(url string) (*NftData, error) {
	if sv.mode != OFF_CHAIN {
		return nil, errors.New("SERVER_MODE_IS_ON_CHAIN")
	}

	// get nft by url
	if url == "" {
		return nil, WrapperError("[ExportProof] _NFT_URL_REQUIRED_")
	}

	nftData, err := sv.DB.GetNFtDataByUrl(context.Background(), url)
	if err != nil {
		fmt.Println("[ExportProof] Get nft data error ", err)
		fmt.Println(err)
		return nil, err
	}

	// export data and delete
	nftDataRes := &NftData{
		ID:   nftData.ID,
		Url:  nftData.Url,
		Memo: nftData.Memo,
	}

	// delete data
	err = sv.DB.DeleteNftDataByUrl(context.Background(), url)
	if err != nil {
		fmt.Println("[ExportProof] Delete nft data error ", err)
		return nil, err
	}

	return nftDataRes, nil
}

func (sv *ServerOffChain) SetMode(mode string) error {
	if mode != ON_CHAIN && mode != OFF_CHAIN {
		return WrapperError("MODE_INVALID")
	}

	sv.mode = mode
	return nil
}

func (sv *ServerOffChain) GetTx(txId string) (interface{}, error) {
	//using a struct literal
	bc := gobcy.API{"0e3279a9ec4e4859ba55945c6a29a6ec", "btc", "test3"}

	//query away
	fmt.Println(bc.GetChain())

	res, err := bc.GetTX(txId, make(map[string]string))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return res, nil

	//fmt.Println("Success create add key chain")
	//fmt.Println(res)
	//bc.CreateWallet()
	//fmt.Println(bc.GetBlock(300000, "", nil))
}
