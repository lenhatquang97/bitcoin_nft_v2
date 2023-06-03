package handler

// SendRequest Post
type SendRequest struct {
	Address string   `json:"address"`
	Amount  int64    `json:"amount"`
	Urls    []string `json:"urls"`
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

// Config config
type Config struct {
	Mode          string
	Network       string
	Host          string
	User          string
	Password      string
	SenderAddress string
	Port          string
}
