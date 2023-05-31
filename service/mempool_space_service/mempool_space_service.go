package mempool_space_service

import (
	"fmt"
	"ordbook-aggregation/config"
	"ordbook-aggregation/tool"
)

//Get TxDetail
func GetTxHex(net, txId string) (string, int, error) {
	var (
		url        string
		code int
		result        string
		err        error
		query map[string]string = map[string]string{}
		headers map[string]string = map[string]string{
			"Ok-Access-Key":config.OklinkKey,
		}
	)
	url = fmt.Sprintf("%s/api/tx/%s/hex", config.MempoolSpace, txId)
	if net != "mainnet" {
		url = fmt.Sprintf("%s/%s/api/tx/%s/hex", config.MempoolSpace, net, txId)
	}

	result, code, err = tool.GetUrlAndCode(url, query, headers)
	if err != nil {
		return "", code, err
	}
	return result, code, nil
}