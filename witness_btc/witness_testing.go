package witnessbtc

import (
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
)

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
