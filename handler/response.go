package handler

import "bitcoin_nft_v2/business"

type SendResponseData struct {
	RevealTxID string `json:"revealTxId"`
	CommitTxID string `json:"commitTxId"`
	Fee        int64  `json:"fee"`
}

type SendResponse struct {
	Code    int32            `json:"code"`
	Message string           `json:"message"`
	Data    SendResponseData `json:"data"`
}

type CheckBalanceResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    int64  `json:"data"`
}

type ImportProofResponse struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type ExportProofResponse struct {
	Code    int32            `json:"code"`
	Message string           `json:"message"`
	Data    business.NftData `json:"data"`
}

type ViewNftDataResponse struct {
	Code    int32              `json:"code"`
	Message string             `json:"message"`
	Data    []business.NftData `json:"data"`
}

type CreateWalletResponseData struct {
	Seed string `json:"seed"`
}

type CreateWalletResponse struct {
	Code    int32                     `json:"code"`
	Message string                    `json:"message"`
	Data    *CreateWalletResponseData `json:"data"`
}

type GetTxResponse struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type GetNftFromUtxoRes struct {
	Code    int32         `json:"code"`
	Message string        `json:"message"`
	Data    []NftFromUtxo `json:"data"`
}

type NftFromUtxo struct {
	HexData    string `json:"hexData"`
	MimeType   string `json:"mimeType"`
	TxId       string `json:"txId"`
	OriginTxId string `json:"originTxId"`
}
