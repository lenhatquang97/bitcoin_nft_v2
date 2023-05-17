package witnessbtc

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
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

func FindMultiplePartsOfByteArray(part []byte, array []byte) []int {
	m := len(part)
	n := len(array)

	result := make([]int, 0)

	for i := 0; i <= n-m; i++ {
		var j = 0
		for j = 0; j < m; j++ {
			if array[i+j] != part[j] {
				break
			}
		}
		if j == m {
			result = append(result, i)
		}
	}
	return result
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
		multipleIndexes := FindMultiplePartsOfByteArray([]byte("m25"), embeddedData)
		for i := 0; i < len(multipleIndexes)-1; i++ {
			startChunkWithPadding := multipleIndexes[i] + len([]byte("m25"))
			endChunk := multipleIndexes[i+1] - GetPaddingInAddData([]byte("m25"))
			padding := GetPaddingInAddData(embeddedData[startChunkWithPadding:endChunk])
			actualStartChunk := startChunkWithPadding + padding
			actualEndChunk := endChunk
			body = append(body, embeddedData[actualStartChunk:actualEndChunk]...)
		}
		startChunkWithPadding := multipleIndexes[len(multipleIndexes)-1] + len([]byte("m25"))
		endChunk := len(embeddedData) - 1
		padding := GetPaddingInAddData(embeddedData[startChunkWithPadding:endChunk])

		actualStartBody := startChunkWithPadding + padding
		actualEndBody := endChunk
		body = append(body, embeddedData[actualStartBody:actualEndBody]...)
		return body
	}
	return nil
}

func IterateWitness(client *rpcclient.Client, revealTxHash *chainhash.Hash) {
	retrievedTx, err := client.GetRawTransaction(revealTxHash)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, item := range retrievedTx.MsgTx().TxIn {
		for _, witnessItem := range item.Witness {
			fmt.Println(witnessItem)
		}
	}
}
