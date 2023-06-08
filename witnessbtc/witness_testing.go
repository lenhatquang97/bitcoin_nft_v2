package witnessbtc

import (
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
)

func DoFirstTestCaseWithPNG() {
	embeddedData, err := PrepareInscriptionData("./README.md")
	if err != nil {
		fmt.Println(err)
		return
	}
	body := DeserializeWitnessDataIntoInscription(embeddedData)
	WriteData(body, "./GG.md")
}
func DoSecondTestCaseWithText() {
	privKey, _ := btcec.NewPrivateKey()
	embeddedData, _ := utils.CreateInscriptionScript(privKey.PubKey(), []byte("Hello World. It's me Mario!"))
	body := DeserializeWitnessDataIntoInscription(embeddedData)
	fmt.Println(string(body))
}
