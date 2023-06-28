package handler

import (
	"bitcoin_nft_v2/business"
	"bitcoin_nft_v2/config"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/gin-gonic/gin"
)

const (
	ON_CHAIN  = "on_chain"
	OFF_CHAIN = "off_chain"
	TESTNET   = "testnet3"
	SIMNET    = "simnet"
)

var sv *business.Server

func Init(conf config.NetworkConfig, mode string) (*business.Server, error) {
	var err error
	sv, err = business.NewServer(&conf, mode)
	if err != nil {
		return nil, err
	}

	return sv, nil
}

func RegisterRoutes(rg *gin.RouterGroup) {
	router := rg.Group("/btc_nft")
	router.POST("/predefine", WrapperPredefineEstimatedFee)
	router.POST("/send", WrapperSend)
	router.POST("/import", WrapperImportProof)
	router.POST("/wallet", WrapperCreateWallet)
	router.PUT("/mode", WrapperSetMode)
	router.POST("/export", WrapperExportProof)
	router.GET("/view-data", WrapperViewNftData)
	router.GET("/balance", WrapperCheckBalance)
	router.GET("/on-chain-nft", WrapperGetNftFromUtxo)
	router.GET("/tx/size", WrapperGetTxSize)
	router.GET("/render", WrapperRenderTree)
	router.GET("/ipfs-link", WrapperIpfsLink)
}

func Run(config *Config) {
	err := ValidateConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	networkCfg := CreateNetworkConfig(config)
	_, err = Init(networkCfg, config.Mode)
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

func ValidateConfig(conf *Config) error {
	if conf.Mode != ON_CHAIN && conf.Mode != OFF_CHAIN {
		return errors.New("MODE_IS_INVALID")
	}

	if conf.Host == "" {
		return errors.New("HOST_ADDRESS_IS_EMPTY")
	}

	flagCheck := strings.Split(conf.Host, ":")
	if len(flagCheck) != 2 {
		return errors.New("HOST_FORMAT_INVALID")
	}

	if conf.User == "" {
		return errors.New("USER_IS_EMPTY")
	}

	if conf.Password == "" {
		return errors.New("PASSWORD_IS_EMPTY")
	}

	if conf.Network != TESTNET && conf.Network != SIMNET {
		return errors.New("NETWORK_IS_INVALID")
	}

	return nil
}

func CreateNetworkConfig(conf *Config) config.NetworkConfig {
	var networkParams *chaincfg.Params
	if conf.Network == TESTNET {
		networkParams = &chaincfg.TestNet3Params
	} else {
		networkParams = &chaincfg.SimNetParams
	}

	return config.NetworkConfig{
		Host:         conf.Host,
		Endpoint:     "ws",
		User:         conf.User,
		Params:       conf.Network,
		Pass:         conf.Password,
		ParamsObject: networkParams,
	}
}
