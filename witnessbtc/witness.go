package witnessbtc

import (
	"bitcoin_nft_v2/utils"
	"fmt"
	"github.com/btcsuite/btcd/txscript"
)

const (
	CONTENT_TAG = "m25"
)

func GetPaddingInAddData(data []byte) int {
	dataLen := len(data)
	//Case it's a opcode
	condition1 := dataLen == 0 || dataLen == 1 && data[0] == 0
	condition2 := dataLen == 1 && data[0] <= 16
	condition3 := dataLen == 1 && data[0] == 0x81
	if condition1 || condition2 || condition3 {
		return 1
	}

	if dataLen < txscript.OP_PUSHDATA1 {
		return 1
	} else if dataLen <= 0xff {
		return 2
	} else if dataLen <= 0xffff {
		return 3
	} else {
		return 5
	}
}

func DeserializeWitnessDataIntoInscription(embeddedData []byte) ([]byte, bool) {
	fixedBytes := []byte{txscript.OP_CHECKSIG, txscript.OP_0, txscript.OP_IF}
	validPosition := -1
	for i := range embeddedData {
		if i+2 < len(embeddedData) && embeddedData[i] == fixedBytes[0] && embeddedData[i+1] == fixedBytes[1] && embeddedData[i+2] == fixedBytes[2] {
			validPosition = i
			break
		}
	}
	var body = make([]byte, 0)
	flagEnd := "m25end"
	flagData := "m25start-data"
	flagRef := "m25start-ref"
	isRef := false
	fmt.Println(validPosition)
	if validPosition != -1 {
		startBodyPos1 := utils.FindStartOfByteArray([]byte(flagData), embeddedData) //+ len([]byte("m25start-data"))
		startBodyPos2 := utils.FindStartOfByteArray([]byte(flagRef), embeddedData)  //+ len([]byte("m25start-ref"))
		startBodyPos := startBodyPos1
		if startBodyPos1 == -1 {
			startBodyPos = startBodyPos2 + len([]byte(flagRef))
			flagEnd += "-ref"
			isRef = true
		} else {
			startBodyPos += len([]byte(flagData))
			flagEnd += "-data"
		}
		endBodyPos := startBodyPos + 500
		if startBodyPos == -1 {
			return nil, false
		}

		if endBodyPos < len(embeddedData) {
			for endBodyPos < len(embeddedData) {
				body = append(body, embeddedData[startBodyPos:endBodyPos]...)
				padding := GetPaddingInAddData(embeddedData[startBodyPos:endBodyPos])
				startBodyPos = endBodyPos + padding
				endBodyPos = startBodyPos + 500
			}
			finalBodyPos := utils.FindStartOfByteArrayFromEnd([]byte(flagEnd), embeddedData, len(embeddedData)-1)
			body = append(body, embeddedData[startBodyPos:finalBodyPos]...)
		} else {
			finalBodyPos := utils.FindStartOfByteArrayFromEnd([]byte(flagEnd), embeddedData, len(embeddedData)-1)
			//body = append(body, embeddedData[startBodyPos:len(embeddedData)-1]...)
			fmt.Println("flag end", finalBodyPos)
			body = append(body, embeddedData[startBodyPos:finalBodyPos]...)
		}
	}

	return body, isRef
}
