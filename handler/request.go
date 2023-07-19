package handler

import "bitcoin_nft_v2/business"

// SendRequest Post
type SendRequest struct {
	Address    string            `json:"address"`
	Passphrase string            `json:"passphrase"`
	IsSendNFT  bool              `json:"isSendNft"`
	IsRef      bool              `json:"isRef"`
	IsMint     bool              `json:"isMint"`
	Urls       []string          `json:"urls"`
	TxID       string            `json:"txId"`
	Data       *business.NftData `json:"data"`
}

// CheckBalanceRequest Get
type CheckBalanceRequest struct {
	Address string `json:"address"`
}

// ViewNftDataRequest Get
type ViewNftDataRequest struct {
}

// ImportNftDataRequest post
type ImportNftDataRequest struct {
	ID   string `json:"id"`
	Url  string `json:"url"`
	Memo string `json:"memo"`
}

// ExportNftDataRequest post
type ExportNftDataRequest struct {
	Url string
}

// SwitchModeRequest put
type SwitchModeRequest struct {
	Mode string `json:"mode"`
}

type CreateWalletRequest struct {
	Name       string `json:"name"`
	Passphrase string `json:"passphrase"`
}

// Config config
type Config struct {
	Mode     string
	Network  string
	Host     string
	User     string
	Password string
	Port     string
}

type GetTxRequest struct {
	TxID string `json:"txId" form:"txId"`
}
