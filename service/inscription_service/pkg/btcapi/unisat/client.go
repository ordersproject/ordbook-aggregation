package unisat

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"io"
	"log"
	"ordbook-aggregation/service/inscription_service/pkg/btcapi"
)

type UniSatClient struct {
	baseURL string
}

func NewClient(netParams *chaincfg.Params) *UniSatClient {
	baseURL := ""
	if netParams.Net == wire.MainNet {
		baseURL = "https://unisat.io"
	} else if netParams.Net == wire.TestNet3 {
		baseURL = "https://unisat.io/testnet"
	} else {
		log.Fatal("UniSat don't support other netParams")
	}
	return &UniSatClient{
		baseURL: baseURL,
	}
}

func (c *UniSatClient) request(method, subPath string, requestBody io.Reader) ([]byte, error) {
	return btcapi.Request(method, c.baseURL, subPath, requestBody)
}

var _ btcapi.BTCAPIClient = (*UniSatClient)(nil)
