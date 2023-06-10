package witnessbtc

import (
	"bitcoin_nft_v2/utils"
	"fmt"
	"net/http"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
)

func GetFileContentType(out *os.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

func ReadFile(filePath string) ([]byte, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, "", err
	}

	fileSize := fileInfo.Size()
	if fileSize >= 3*1024*1024 {
		return nil, "", fmt.Errorf("too much %d bytes for embedding NFT data", fileSize)
	}

	binFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}

	contentType, err := GetFileContentType(file)
	if err != nil {
		return nil, "", err
	}

	return binFile, contentType, nil
}

func PrepareInscriptionData(data string, isRef bool) ([]byte, error) {
	var rawData []byte
	var err error
	if isRef {
		rawData = []byte(data)
	} else {
		rawData, _, err = ReadFile(data)
	}
	if err != nil {
		return nil, err
	}

	privKey, _ := btcec.NewPrivateKey()
	embeddedData, _ := utils.CreateInscriptionScriptV2(privKey.PubKey(), rawData, isRef)
	return embeddedData, nil
}

func WriteData(body []byte, outputFilePath string) {
	err := os.WriteFile(outputFilePath, body, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
}
