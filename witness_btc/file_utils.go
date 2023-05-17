package witnessbtc

import (
	"bitcoin_nft_v2/utils"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
)

func PrepareData(filePath string) ([]byte, error) {
	rawData, contentType, err := ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fmt.Println(contentType)

	privKey, _ := btcec.NewPrivateKey()
	embeddedData, _ := utils.CreateInscriptionScript(privKey.PubKey(), rawData)
	return embeddedData, nil
}

func WriteData(body []byte, outputFilePath string) {
	err := os.WriteFile(outputFilePath, body, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
}
