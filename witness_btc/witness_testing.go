package witnessbtc

import (
	"bitcoin_nft_v2/utils"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
)

func PrepareData(filePath string) ([]byte, error) {
	rawData, err := ReadFile(filePath)
	if err != nil {
		return nil, err
	}

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

func DoFirstTestCaseWithMP3() {
	embeddedData, err := PrepareData("./sample-15s.mp3")
	if err != nil {
		fmt.Println(err)
		return
	}
	body := DeserializeWitnessDataIntoInscription(embeddedData)
	WriteData(body, "./final.mp3")
}
func DoSecondTestCaseWithText() {
	privKey, _ := btcec.NewPrivateKey()
	embeddedData, _ := utils.CreateInscriptionScript(privKey.PubKey(), []byte("Hello World. It's me Mario!"))
	body := DeserializeWitnessDataIntoInscription(embeddedData)
	fmt.Println(string(body))
}
