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

type InscriptionResult struct {
	List  []*InscriptionDetailItem `json:"list"`
	Total int64                    `json:"total"`
}

type InscriptionDetailItem struct {
	InscriptionId      string `json:"inscriptionId"`
	InscriptionNumber  int64  `json:"inscriptionNumber"`
	Address            string `json:"address"`
	OutputValue        int64  `json:"outputValue"`
	Preview            string `json:"preview"`
	Content            string `json:"content"`
	ContentLength      int    `json:"contentLength"`
	ContentType        string `json:"contentType"`
	ContentBody        string `json:"contentBody"`
	Timestamp          int64  `json:"timestamp"`
	GenesisTransaction string `json:"genesisTransaction"`
	Location           string `json:"location"`
	Output             string `json:"output"`
	Offset             int    `json:"offset"`
	UtxoHeight         int64  `json:"utxoHeight"`
	UtxoConfirmation   int64  `json:"utxoConfirmation"`
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

	url = fmt.Sprintf("%s/wallet-v4/address/btc-utxo", config.UnisatDomain)
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

func GetAddressInscriptions(address string) ([]*InscriptionDetailItem, error) {
	var (
		url        string
		result     string
		resp       *UtxoDetailResp
		resultData *InscriptionResult
		//data   []*InscriptionDetailItem = make([]*InscriptionDetailItem, 0)
		err   error
		query map[string]string = map[string]string{
			"address": address,
			"cursor":  "0",
			"size":    "100",
		}
		headers map[string]string = map[string]string{
			"Content-Type": "application/json",
			"X-Client":     "UniSat Wallet",
			"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		}
	)

	url = fmt.Sprintf("%s/wallet-v4/address/inscriptions", config.UnisatDomain)
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

	if err = tool.JsonToAny(resp.Result, &resultData); err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}
	if len(resultData.List) == 0 {
		return nil, errors.New("No Data. ")
	}

	return resultData.List, nil
}

// Get fee detail
func GetFeeDetail() (*FeeSummary, error) {
	var (
		url     string
		result  string
		resp    *UtxoDetailResp
		data    *FeeSummary = &FeeSummary{}
		err     error
		query   map[string]string = map[string]string{}
		headers map[string]string = map[string]string{
			"Content-Type": "application/json",
			"X-Client":     "UniSat Wallet",
			"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		}
	)

	url = fmt.Sprintf("%s/wallet-v4/default/fee-summary", config.UnisatDomain)
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

	return data, nil
}
