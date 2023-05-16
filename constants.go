package main

type NetworkConfig struct {
	Host     string
	Endpoint string
	User     string
	Pass     string
	CertName string
	Params   string
}

const (
	SenderAddress      = "SeZdpbs8WBuPHMZETPWajMeXZt1xzCJNAJ"
	PassphraseInWallet = "12345"
	PassphraseTimeout  = 3
	CoinsToSend        = 10000
)

var EmbeddedData = []byte("Hello World")
var SimNetConfig = NetworkConfig{
	Host:     "localhost:18554",
	Endpoint: "ws",
	User:     "youruser",
	Pass:     "SomeDecentp4ssw0rd",
	Params:   "simnet",
}
