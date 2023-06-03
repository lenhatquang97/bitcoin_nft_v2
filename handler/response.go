package handler

type SendResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// Data ?
}

type CheckBalanceResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    int64  `json:"data"`
}

type ImportProofResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ExportProofResponse struct {
	Code    string `json:"code"`
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
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Data    []NftData `json:"data"`
}
