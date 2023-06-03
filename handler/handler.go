package handler

import (
	"bitcoin_nft_v2/business"
	"bitcoin_nft_v2/config"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

const (
	ON_CHAIN  = "on_chain"
	OFF_CHAIN = "off_chain"
)

var sv *business.ServerOffChain

var SimNetConfig = config.NetworkConfig{
	Host:          "localhost:18554",
	Endpoint:      "ws",
	User:          "youruser",
	Pass:          "SomeDecentp4ssw0rd",
	Params:        "simnet",
	ParamsObject:  &chaincfg.SimNetParams,
	SenderAddress: "SeZdpbs8WBuPHMZETPWajMeXZt1xzCJNAJ",
}

var TestNetConfig = config.NetworkConfig{
	Host:         "localhost:18332",
	Endpoint:     "ws",
	User:         "DeW+bgKg011pJHZnaBvgv/lMRks=",
	Pass:         "wD9aohGo2f5LwVg7fdj1ntHQcfY=",
	Params:       "testnet3",
	ParamsObject: &chaincfg.TestNet3Params,
	//Note: in testnet, address is not reused so you need to use default address
	//Another note: Default address has changed everytime you init the server => In UI, you need a mechanism to
	//choose address anyway.
	SenderAddress: "mntb2RxQhyXqXRZV5GE1bDkP6615EPXLHF",
}

func Init(mode string) (*business.ServerOffChain, error) {
	if mode == ON_CHAIN {

	} else {
		var err error
		sv, err = business.NewServerOffChain(&TestNetConfig, OFF_CHAIN)
		if err != nil {
			return nil, err
		}
	}

	return sv, nil
}

func RegisterRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/btc_nft")
	router.POST("/send", WrapperSend)
	router.POST("/import", WrapperImportProof)
	router.POST("/export", WrapperExportProof)
	router.GET("/view-data", WrapperViewNftData)
	router.GET("/balance", WrapperCheckBalance)
}

func Run() {
	mode := OFF_CHAIN
	_, err := Init(mode)
	if err != nil {
		fmt.Print(err)
		return
	}

	controller := gin.Default()
	basepath := controller.Group("/v1")
	RegisterRoutes(basepath)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(controller.Run(":" + port))

}
