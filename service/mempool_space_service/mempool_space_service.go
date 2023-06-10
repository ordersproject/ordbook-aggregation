package mempool_space_service

import (
	"errors"
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
		}
	)
	url = fmt.Sprintf("%s/api/tx/%s/hex", config.MempoolSpace, txId)
	if net != "mainnet" && net != "livenet" {
		url = fmt.Sprintf("%s/%s/api/tx/%s/hex", config.MempoolSpace, net, txId)
	}

	result, code, err = tool.GetUrlAndCode(url, query, headers)
	if err != nil {
		return "", code, err
	}
	return result, code, nil
}


func BroadcastTx(net, hex string) (string, error) {
	var (
		url        string
		code int
		result        string
		err        error
		headers map[string]string = map[string]string{
		}
	)

	fmt.Println(hex)
	url = fmt.Sprintf("%s/api/tx", config.MempoolSpace)
	if net != "mainnet" && net != "livenet" {
		url = fmt.Sprintf("%s/%s/api/tx", config.MempoolSpace, net)
	}
	fmt.Println(url)
	result, code, err = tool.PostUrlAndCode(url, hex, headers)
	if err != nil {
		return "", err
	}
	fmt.Println(result)
	fmt.Println(code)
	//return result, nil
	if code != 200 {
		return "", errors.New(fmt.Sprintf("Post err: code not 200, msg:%s", result))
	}

	return result, nil
}