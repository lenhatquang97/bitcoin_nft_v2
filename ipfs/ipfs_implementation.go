package ipfs

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

const YourPublicKey = "k51qzi5uqu5dil4952pc6hw4063hgsxyx8lae8txbpxkf2iajvj43u71gwudyu"

func ReadFileForIpfs(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	binFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return binFile, nil
}

func DownloadOnIpfs(fileLink string) ([]byte, error) {
	sh := shell.NewShell("localhost:5001")
	//Get cid based on file link
	cid := fileLink[strings.LastIndex(fileLink, "/")+1 : strings.LastIndex(fileLink, "?")]
	err := sh.Get(cid, "./")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	binFile, err := ReadFileForIpfs("./" + cid)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	os.Remove("./" + cid)

	return binFile, nil
}

func GetIpfsLink(filePath string) (string, error) {
	sh := shell.NewShell("localhost:5001")
	fileBin, err := ReadFileForIpfs(filePath)
	if err != nil {
		return "", err
	}
	//Get filename from filePath by cutting the last slash
	filename := filePath[strings.LastIndex(filePath, "/")+1:]
	reader := bytes.NewReader(fileBin)
	cid, err := sh.Add(reader)
	if err != nil {
		return "", err
	}

	//Add to IPNS
	var lifetime time.Duration = 10 * time.Hour
	var ttl time.Duration = 1 * time.Microsecond
	_, err = sh.PublishWithDetails(cid, YourPublicKey, lifetime, ttl, true)
	if err != nil {
		return "", err
	}
	_, err = sh.Resolve(YourPublicKey)
	if err != nil {
		return "", err
	}
	link := "https://ipfs.io/ipfs/" + cid + "?filename=" + filename
	return link, nil
}
