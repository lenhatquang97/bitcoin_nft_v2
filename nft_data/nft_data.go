package nft_data

import (
	"crypto/sha256"
	"encoding/json"
)

type NftData struct {
	ID   string `json:"id" bson:"id"`
	Url  string `json:"url" bson:"url"`
	Memo string `json:"memo" bson:"memo"`
}

func GetSampleNftData() *NftData {
	//id := uuid.New().String()
	id := "123456789"
	return &NftData{
		ID:   id,
		Url:  "https://upload.wikimedia.org/wikipedia/en/b/bd/Doraemon_character.png",
		Memo: "",
	}
}

func GetSampleDataByte() ([]byte, [32]byte) {
	sampleData := GetSampleNftData()
	h := sha256.New()
	_, _ = h.Write([]byte(sampleData.ID))
	_, _ = h.Write([]byte(sampleData.Url))
	_, _ = h.Write([]byte(sampleData.Memo))

	rawData, err := json.Marshal(sampleData)
	if err != nil {
		return nil, [32]byte{}
	}

	return rawData, *(*[32]byte)(h.Sum(nil))
}
