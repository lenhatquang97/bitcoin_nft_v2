package witnessbtc

import (
	"bitcoin_nft_v2/utils"

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

func DeserializeWitnessDataIntoInscription(embeddedData []byte) []byte {
	fixedBytes := []byte{txscript.OP_CHECKSIG, txscript.OP_0, txscript.OP_IF}
	validPosition := -1
	for i := range embeddedData {
		if i+2 < len(embeddedData) && embeddedData[i] == fixedBytes[0] && embeddedData[i+1] == fixedBytes[1] && embeddedData[i+2] == fixedBytes[2] {
			validPosition = i
			break
		}
	}
	var body = make([]byte, 0)
	if validPosition != -1 {
		multipleIndexes := utils.FindMultiplePartsOfByteArray([]byte(CONTENT_TAG), embeddedData)
		for i := 0; i < len(multipleIndexes)-1; i++ {
			startChunkWithPadding := multipleIndexes[i] + len([]byte(CONTENT_TAG))
			endChunk := multipleIndexes[i+1] - GetPaddingInAddData([]byte(CONTENT_TAG))
			padding := GetPaddingInAddData(embeddedData[startChunkWithPadding:endChunk])
			actualStartChunk := startChunkWithPadding + padding
			actualEndChunk := endChunk
			body = append(body, embeddedData[actualStartChunk:actualEndChunk]...)
		}
		startChunkWithPadding := multipleIndexes[len(multipleIndexes)-1] + len([]byte(CONTENT_TAG))
		endChunk := len(embeddedData) - 1
		padding := GetPaddingInAddData(embeddedData[startChunkWithPadding:endChunk])

		actualStartBody := startChunkWithPadding + padding
		actualEndBody := endChunk
		body = append(body, embeddedData[actualStartBody:actualEndBody]...)
		return body
	}
	return nil
}
