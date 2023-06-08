package utils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
)

func ChunkSlice(slice []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

func LoadCerts(baseFolder string) ([]byte, error) {
	certHomeDir := btcutil.AppDataDir(baseFolder, false)
	certs, err := ioutil.ReadFile(filepath.Join(certHomeDir, "rpc.cert"))
	if err != nil {
		return nil, err
	}
	return certs, nil
}

func FindStartOfByteArray(part []byte, array []byte) int {
	m := len(array)
	n := len(part)
	for i := 0; i <= m-n; i++ {
		found := true
		for j := 0; j < n; j++ {
			if array[i+j] != part[j] {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
}

// Find first occurence of part in array from end index
func FindStartOfByteArrayFromEnd(part []byte, array []byte, end int) int {
	n := len(part)
	for i := end; i >= 0; i-- {
		found := true
		for j := 0; j < n; j++ {
			if array[i+j] != part[j] {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
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
