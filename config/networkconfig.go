package config

import "github.com/btcsuite/btcd/chaincfg"

type NetworkConfig struct {
	Host         string
	Endpoint     string
	User         string
	Pass         string
	CertName     string
	Params       string
	ParamsObject *chaincfg.Params
}
