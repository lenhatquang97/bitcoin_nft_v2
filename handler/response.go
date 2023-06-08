package handler

type SendResponseData struct {
	TxID string `json:"txId"`
	Fee  int64  `json:"fee"`
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
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID   string `json:"id"`
		Url  string `json:"url"`
		Memo string `json:"memo"`
	} `json:"data"`
}

type NftData struct {
	ID   string `json:"id"`
	Url  string `json:"url"`
	Memo string `json:"memo"`
}

type ViewNftDataResponse struct {
	Code    int32     `json:"code"`
	Message string    `json:"message"`
	Data    []NftData `json:"data"`
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
