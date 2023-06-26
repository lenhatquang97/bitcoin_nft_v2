package witnessbtc

import (
	"fmt"
	"net/http"
	"os"
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
	if fileSize >= 2.5*1024*1024 {
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
