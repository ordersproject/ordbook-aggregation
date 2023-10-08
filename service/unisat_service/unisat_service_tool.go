package unisat_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/config"
	"ordbook-aggregation/tool"
)

type UtxoDetailResp struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

type UtxoDetailItem struct {
	TxId         string        `json:"txId"`
	OutputIndex  int64         `json:"outputIndex"`
	Satoshis     int64         `json:"satoshis"`
	ScriptPk     string        `json:"scriptPk"`
	AddressType  int           `json:"addressType"`
	Inscriptions []interface{} `json:"inscriptions"`
}

func GetAddressUtxo(address string) ([]*UtxoDetailItem, error) {
	var (
		url    string
		result string
		resp   *UtxoDetailResp
		data   []*UtxoDetailItem = make([]*UtxoDetailItem, 0)
		err    error
		query  map[string]string = map[string]string{
			"address": address,
		}
		headers map[string]string = map[string]string{
			"Content-Type": "application/json",
			"X-Client":     "UniSat Wallet",
			"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		}
	)

	url = fmt.Sprintf("%s/wallet-api-v4/address/btc-utxo", config.UnisatDomain)
	result, err = tool.GetUrl(url, query, headers)
	if err != nil {
		return nil, err
	}

	//fmt.Println(result)
	if err = tool.JsonToObject(result, &resp); err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	if resp.Status != UnisatCodeSuccess {
		return nil, errors.New(fmt.Sprintf("Msg:%s", resp.Message))
	}

	if err = tool.JsonToAny(resp.Result, &data); err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}
	if len(data) == 0 {
		return nil, errors.New("No Data. ")
	}

	return data, nil
}
