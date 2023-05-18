package offchainnft

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func FileSha256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Create a new SHA256 hash
	hash := sha256.New()

	// Copy the file contents to the hash calculator
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Get the hash sum as a byte slice
	hashSum := hash.Sum(nil)
	return hex.EncodeToString(hashSum), nil

}
