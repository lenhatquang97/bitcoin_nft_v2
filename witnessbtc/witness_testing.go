package witnessbtc

import (
	"bitcoin_nft_v2/utils"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
)

func DoFirstTestCaseWithPNG() {
	embeddedData, err := PrepareInscriptionData("./README.md", false)
	if err != nil {
		fmt.Println(err)
		return
	}
	body, _ := DeserializeWitnessDataIntoInscription(embeddedData, ON_CHAIN)
	fmt.Println(string(body))
	WriteData(body, "./GG.md")
}
func DoSecondTestCaseWithText() {
	privKey, _ := btcec.NewPrivateKey()
	embeddedData, _ := utils.CreateInscriptionScript(privKey.PubKey(), []byte("Hello World. It's me Mario!"))
	body, _ := DeserializeWitnessDataIntoInscription(embeddedData, ON_CHAIN)
	fmt.Println(string(body))
}
