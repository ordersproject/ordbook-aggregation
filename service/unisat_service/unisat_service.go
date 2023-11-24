package unisat_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/config"
	"ordbook-aggregation/tool"
	"strings"
)

const (
	UnisatCodeSuccess string = "1"
)

func BroadcastTx(net, hex string) (*BroadcastTxResp, error) {
	var (
		url    string
		result string
		resp   *BroadcastTxResp
		err    error
		req    map[string]string = map[string]string{
			"rawtx": hex,
		}
		headers map[string]string = map[string]string{
			"X-Client": "UniSat Wallet",
		}
	)

	fmt.Println(hex)
	url = fmt.Sprintf("%s/wallet-v4/tx/broadcast", config.UnisatDomain)
	if strings.ToLower(net) == "testnet" {
		url = fmt.Sprintf("%s/testnet/wallet-v4/tx/broadcast", config.UnisatDomain)
	}
	fmt.Println(url)
	result, err = tool.PostUrl(url, req, headers)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)
	if err = tool.JsonToObject(result, &resp); err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	if resp.Status != UnisatCodeSuccess {
		return nil, errors.New(fmt.Sprintf("Msg:%s", resp.Message))
	}

	return resp, nil
}
