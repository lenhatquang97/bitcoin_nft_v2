package business

import (
	"bitcoin_nft_v2/config"
	"bitcoin_nft_v2/db"
	"bitcoin_nft_v2/db/sqlc"
	"bitcoin_nft_v2/gobcy"
	"bitcoin_nft_v2/nft_tree"
	"bitcoin_nft_v2/utils"
	"bitcoin_nft_v2/witnessbtc"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/btcsuite/btcd/mempool"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
)

const (
	PassphraseInWallet = "12345"
	PassphraseTimeout  = 3
	ON_CHAIN           = "on_chain"
	OFF_CHAIN          = "off_chain"
	USER_PASWORD       = "user_password"
	LEFT_STR           = "Your wallet generation seed is:"
	RIGHT_STR          = "IMPORTANT: Keep the seed in a safe place as you"
)

type Server struct {
	client *rpcclient.Client
	mode   string
	Config *config.NetworkConfig
	DB     *db.PostgresStore
}

func NewServer(networkCfg *config.NetworkConfig, mode string) (*Server, error) {
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

	return &Server{
		client: client,
		mode:   mode,
		Config: networkCfg,
		DB:     store,
	}, nil
}

func (sv *Server) CalculateFee(toAddress string, amount int64, isRef bool, data interface{}, passphrase string) (int64, error) {
	var dataSend []byte
	var err error
	if sv.mode == OFF_CHAIN {
		dataSend, err = sv.GetDataSendOffChain(data, isRef)
		if err != nil {
			fmt.Println("Compute root hash for receiver error")
			fmt.Println(err)
			return 0, err
		}
	} else {
		dataSend, err = sv.GetDataSendOnChain(data, isRef)
		if err != nil {
			return 0, err
		}
	}

	//Step 2: Open passphrase
	err = sv.client.WalletPassphrase(passphrase, PassphraseTimeout)
	if err != nil {
		return 0, err
	}

	//Step 3: Calculate fee (commit and revealTxFee)
	estimatedCommitTxFee, err := FakeCommitTxFee(sv, dataSend, amount, isRef)
	if err != nil {
		return 0, err
	}

	estimatedRevealTxFee, err := FakeRevealTxFee(sv, dataSend, isRef, toAddress, amount)
	if err != nil {
		return 0, err
	}

	return estimatedCommitTxFee + estimatedRevealTxFee, nil
}

// Send
// if on-chain mode data is file path
// else if off-chain mode data is list nft data (list by get data from db)
// if don't have data in DB --> import nft
func (sv *Server) Send(toAddress string, amount int64, isSendNft bool, isRef bool, data interface{}, passphrase string) (string, string, int64, error) {
	//nftUrls := []string{
	//	"https://genk.mediacdn.vn/k:thumb_w/640/2016/photo-1-1473821552147/top6suthatcucsocvepikachu.jpg",
	//	"https://pianofingers.vn/wp-content/uploads/2020/12/organ-casio-ct-s100-1.jpg",
	//	"https://amnhacvietthanh.vn/wp-content/uploads/2020/10/Yamaha-C40.jpg",
	//}
	//nameSpace := DefaultNameSpace
	// Get Nft Data
	var dataSend []byte
	//var contentType string
	var err error
	txIdRef := ""
	var keys [][32]byte
	var leafHash []nft_tree.NodeHash
	if isSendNft {
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
					return "", "", 0, err
				}

				nftData = append(nftData, &NftData{
					ID:   item.ID,
					Url:  item.Url,
					Memo: item.Memo,
				})
			}

			dataSend, keys, leafHash, err = NewRootHashForReceiver(nftData)
			if err != nil {
				fmt.Println("Compute root hash for receiver error")
				fmt.Println(err)
				return "", "", 0, err
			}
		} else {
			if isRef {
				txIdRef = data.([]string)[0]
				originTxId := data.([]string)[1]
				dataSend = []byte(originTxId)
			} else {
				stringArr := data.([]string)
				dataSend, _, err = witnessbtc.ReadFile(stringArr[0])
			}

			if err != nil {
				// log error
				return "", "", 0, err
			}
		}
	}

	//Step 2: Open passphrase
	err = sv.client.WalletPassphrase(passphrase, PassphraseTimeout)
	if err != nil {
		return "", "", 0, err
	}

	//Step 3: Calculate fee (commit and revealTxFee)
	estimatedCommitTxFee, err := FakeCommitTxFee(sv, dataSend, amount, isRef)
	if err != nil {
		return "", "", 0, err
	}

	estimatedRevealTxFee, err := FakeRevealTxFee(sv, dataSend, isRef, toAddress, amount)
	if err != nil {
		return "", "", 0, err
	}

	fmt.Println("Commit tx fee is: ", estimatedCommitTxFee)
	fmt.Println("Reveal tx fee is: ", estimatedRevealTxFee)
	fmt.Println("Estimated fee is: ", estimatedCommitTxFee+estimatedRevealTxFee)

	commitTxHash, wif, err := ExecuteCommitTransaction(sv, dataSend, isRef, txIdRef, estimatedRevealTxFee+amount, estimatedCommitTxFee)
	if err != nil {
		return "", "", 0, err
	}

	fmt.Printf("Your commit tx hash is: %s\n", commitTxHash.String())

	retrievedCommitTx, err := sv.client.GetRawTransaction(commitTxHash)
	if err != nil {
		fmt.Println(err)
		return "", "", 0, err
	}

	fmt.Println("===================================Checkpoint 1====================================")

	revealTxInput := RevealTxInput{
		CommitTxHash: commitTxHash,
		Idx:          0,
		Wif:          wif,
		CommitOutput: retrievedCommitTx.MsgTx().TxOut[0],
		ChainConfig:  sv.Config.ParamsObject,
	}

	revealTxHash, err := ExecuteRevealTransaction(sv.client, &revealTxInput, dataSend, isRef, toAddress, estimatedRevealTxFee, amount)
	if err != nil {
		fmt.Println(err)
		return "", "", 0, err
	}

	if sv.mode == OFF_CHAIN {
		for i, key := range keys {

			txCreator := func(tx *sql.Tx) db.TreeStore {
				return sv.DB.WithTx(tx)
			}

			treeDB := db.NewTransactionExecutor[db.TreeStore](sv.DB, txCreator)

			taroTreeStore := db.NewTaroTreeStore(treeDB)

			tree := nft_tree.NewFullTree(taroTreeStore)

			_, err = tree.Delete(context.Background(), key)
			if err != nil {
				fmt.Println("Delete leaf after reveal tx failed", err)
			}

			destArray := make([]byte, len(leafHash[i]))

			// Copy the elements from source array to destination array
			copy(destArray, leafHash[i][:])

			_, err = sv.DB.DeleteNode(context.Background(), destArray)
			if err != nil {
				fmt.Println("Delete leaf after reveal tx failed", err)
			}
		}
	}

	fmt.Println("===================================Checkpoint 2====================================")
	fmt.Printf("Your reveal tx hash is: %s\n", revealTxHash.String())
	fmt.Println("===================================Success====================================")
	return commitTxHash.String(), revealTxHash.String(), estimatedCommitTxFee + estimatedRevealTxFee, nil
}

func (sv *Server) CheckBalance(address string) (int, error) {
	utxos, err := sv.client.ListUnspent()
	if err != nil {
		return -1, err
	}
	amount := 0

	for i := 0; i < len(utxos); i++ {
		if utxos[i].Address == address {
			//100_000_000 is because it's testnet
			amount += int(utxos[i].Amount * 100_000_000)
		}
	}
	return amount, nil
}

func (sv *Server) ViewNftData() ([]*NftData, error) {
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

func (sv *Server) CreateWallet(passphrase string) (string, error) {
	//res, err := sv.client.CreateW	allet(name, rpcclient.WithCreateWalletPassphrase(passphrase))
	//if err != nil {
	//	return err
	//}

	data, err := os.ReadFile("./template_create_wallet.exp")
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	dataStr := string(data)

	res := strings.Replace(dataStr, USER_PASWORD, passphrase, -1)

	err = os.WriteFile("create_wallet_result.exp", []byte(res), 0666)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	cmd := &exec.Cmd{
		Path:   "./create_wallet_result.exp",
		Stderr: os.Stderr,
	}

	output, err := cmd.Output()
	fmt.Println(err)
	fmt.Println(string(output))

	resStr := string(output)
	l := strings.Index(resStr, LEFT_STR)
	r := strings.Index(resStr, RIGHT_STR)

	if l+len(LEFT_STR) > r {
		return "", errors.New("seed is empty")
	}

	err = os.Remove("./create_wallet_result.exp")
	if err != nil {
		return "", err
	}

	return resStr[l+len(LEFT_STR) : r], err
}

func (sv *Server) GetNftData() {

}

func (sv *Server) ImportProof(id, url, memo string) error {
	if sv.mode != OFF_CHAIN {
		return errors.New("SERVER_MODE_IS_ON_CHAIN")
	}

	// import nft data and merge tree
	dataByte, key := ComputeNftDataByte(&NftData{
		ID:   id,
		Url:  url,
		Memo: memo,
	})

	fmt.Println("Key here: ", key)

	// Init Root Hash For Receiver
	leaf := nft_tree.NewLeafNode(dataByte, 0) // CoinsToSend
	leaf.NodeHash()

	txCreator := func(tx *sql.Tx) db.TreeStore {
		return sv.DB.WithTx(tx)
	}

	treeDB := db.NewTransactionExecutor(sv.DB, txCreator)

	taroTreeStore := db.NewTaroTreeStore(treeDB)

	tree := nft_tree.NewFullTree(taroTreeStore)

	//We use the default, in-memory store that doesn't actually use the
	//context.

	fmt.Println("Hash is: ", leaf.NodeHash().String())
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

func (sv *Server) ExportProof(url string) (*NftData, error) {
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

	dataByte, key := ComputeNftDataByte(nftDataRes)

	leaf := nft_tree.NewLeafNode(dataByte, 0) // C
	txCreator := func(tx *sql.Tx) db.TreeStore {
		return sv.DB.WithTx(tx)
	}

	treeDB := db.NewTransactionExecutor[db.TreeStore](sv.DB, txCreator)

	taroTreeStore := db.NewTaroTreeStore(treeDB)

	tree := nft_tree.NewFullTree(taroTreeStore)

	_, err = tree.Delete(context.Background(), key)
	if err != nil {
		fmt.Println("Delete leaf after reveal tx failed", err)
	}

	nodeHash := leaf.NodeHash()
	destArray := make([]byte, len(nodeHash))

	// Copy the elements from source array to destination array
	copy(destArray, nodeHash[:])

	_, err = sv.DB.DeleteNode(context.Background(), destArray)
	if err != nil {
		fmt.Println("Delete leaf after reveal tx failed", err)
	}

	return nftDataRes, nil
}

func (sv *Server) SetMode(mode string) error {
	if mode != ON_CHAIN && mode != OFF_CHAIN {
		return WrapperError("MODE_INVALID")
	}

	sv.mode = mode
	return nil
}

func (sv *Server) GetTx(txId string) (interface{}, error) {
	//using a struct literal
	bc := gobcy.API{Token: "0e3279a9ec4e4859ba55945c6a29a6ec", Coin: "btc", Chain: "test3"}

	res, err := bc.GetTX(txId, make(map[string]string))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return res, nil
}
func (sv *Server) GetDataSendOffChain(data interface{}, isRef bool) ([]byte, error) {
	var nftData []*NftData
	fmt.Println(data)
	item := data.([]string)[0]
	fmt.Println("Item test", item)
	for _, url := range data.([]string) {
		item, err := sv.DB.GetNFtDataByUrl(context.Background(), url)
		if err != nil {
			print("Get Nft Data Failed")
			fmt.Println(err)
			return nil, err
		}

		nftData = append(nftData, &NftData{
			ID:   item.ID,
			Url:  item.Url,
			Memo: item.Memo,
		})
	}

	dataSend, _, _, err := NewRootHashForReceiver(nftData)
	if err != nil {
		fmt.Println("Compute root hash for receiver error")
		fmt.Println(err)
		return nil, err
	}
	return dataSend, nil
}

func (sv *Server) GetDataSendOnChain(data interface{}, isRef bool) ([]byte, error) {
	if isRef {
		customData, err := RawDataEncode(data.([]string)[0])
		if err != nil {
			return nil, err
		}

		return []byte(customData), nil
	} else {
		dataSend, err := hex.DecodeString(data.([]string)[0])
		if err != nil {
			return nil, err
		}
		return dataSend, nil
	}
}

func (sv *Server) GetNftFromUtxo(address string) ([][]byte, []string, []string, error) {
	utxos, err := sv.client.ListUnspent()
	if err != nil {
		return nil, nil, nil, err
	}

	res := make([][]byte, 0)
	txIds := make([]string, 0)
	orginalTxIds := make([]string, 0)
	for i := 0; i < len(utxos); i++ {
		if utxos[i].Address == address {
			//100_000_000 is because it's testnet
			hashId, err := chainhash.NewHashFromStr(utxos[i].TxID)
			if err != nil {
				return nil, nil, nil, err
			}
			tx, err := sv.client.GetRawTransaction(hashId)
			if err != nil {
				return nil, nil, nil, err
			}

			witness := tx.MsgTx().TxIn[0].Witness
			if len(witness) != 3 {
				continue
			}

			txId := utxos[i].TxID
			data, isRef := witnessbtc.DeserializeWitnessDataIntoInscription(witness[1])
			if isRef {
				hashId, err = chainhash.NewHashFromStr(string(data))
				if err != nil {
					return nil, nil, nil, err
				}
				tx, err = sv.client.GetRawTransaction(hashId)
				if err != nil {
					return nil, nil, nil, err
				}

				witness = tx.MsgTx().TxIn[0].Witness
				if len(witness) != 3 {
					continue
				}

				originTxId := string(data)
				orginalTxIds = append(orginalTxIds, originTxId)
				data, _ = witnessbtc.DeserializeWitnessDataIntoInscription(witness[1])
			} else {
				orginalTxIds = append(orginalTxIds, utxos[i].TxID)
			}
			if data != nil {
				res = append(res, data)
				txIds = append(txIds, txId)
			}
		}
	}

	return res, txIds, orginalTxIds, nil
}

func (sv *Server) GetTxSize(txId string) (int64, int, error) {
	hashId, err := chainhash.NewHashFromStr(txId)
	if err != nil {
		return 0, 0, err
	}
	tx, err := sv.client.GetRawTransaction(hashId)
	if err != nil {
		return 0, 0, err
	}

	virtualSize := mempool.GetTxVirtualSize(tx)
	serializeSize := tx.MsgTx().SerializeSize()

	return virtualSize, serializeSize, nil
}

func (sv *Server) RenderTree() error {
	// get all data
	nftData, err := sv.ViewNftData()
	if err != nil {
		return err
	}

	input := make(map[[sha256.Size]byte]nft_tree.NftData)
	for _, item := range nftData {
		_, key := ComputeNftDataByte(item)
		input[key] = nft_tree.NftData{
			ID:   item.ID,
			Url:  item.Url,
			Memo: item.Memo,
		}
	}

	txCreator := func(tx *sql.Tx) db.TreeStore {
		return sv.DB.WithTx(tx)
	}

	treeDB := db.NewTransactionExecutor[db.TreeStore](sv.DB, txCreator)

	taroTreeStore := db.NewTaroTreeStore(treeDB)

	tree := nft_tree.NewFullTree(taroTreeStore)

	renderedTree, err := tree.RenderTree(context.Background(), input)
	if err != nil {
		return err
	}

	printTree(renderedTree, 3)
	return nil
}

func getMaxWidth(root *nft_tree.VirtualTree, level int) int {
	if root == nil {
		return 0
	}

	if level == 1 {
		return 1
	}

	leftWidth := getMaxWidth(root.Left, level-1)
	rightWidth := getMaxWidth(root.Right, level-1)

	return leftWidth + rightWidth
}

func printTree(root *nft_tree.VirtualTree, level int) {
	if root == nil {
		return
	}

	height := getHeight(root)
	maxWidth := getMaxWidth(root, height)
	printTreeHelper(root, "", maxWidth, height, level)
}

func printTreeHelper(root *nft_tree.VirtualTree, prefix string, maxWidth, currLevel, targetLevel int) {
	if root == nil {
		return
	}

	if currLevel == targetLevel {
		currStr := "———"
		prevStr := prefix

		fmt.Print(prefix)
		fmt.Printf("%s %v\n", currStr, root.Data)

		newPrefix := prefix
		if prevStr != "" {
			newPrefix = strings.Replace(prevStr, "———", "   |", 1)
		}

		printTreeHelper(root.Left, newPrefix+"   |", maxWidth, currLevel-1, targetLevel-1)
		printTreeHelper(root.Right, prevStr, maxWidth, currLevel-1, targetLevel-1)
	} else {
		printTreeHelper(root.Left, prefix, maxWidth, currLevel-1, targetLevel)
		printTreeHelper(root.Right, prefix, maxWidth, currLevel-1, targetLevel)
	}
}

func getHeight(root *nft_tree.VirtualTree) int {
	if root == nil {
		return 0
	}

	leftHeight := getHeight(root.Left)
	rightHeight := getHeight(root.Right)

	if leftHeight > rightHeight {
		return leftHeight + 1
	}
	return rightHeight + 1
}
