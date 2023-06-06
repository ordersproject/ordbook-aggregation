package oklink_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/config"
	"ordbook-aggregation/tool"
	"strconv"
)

const (
	OklinkCodeSuccess = "0"
)

//Get brc20Balance
func GetAddressBrc20BalanceResult(address, tick string, page, limit int64) (*OklinkBrc20BalanceDetails, error){
	var (
		url        string
		result        string
		resp        *OklinkResp
		data        []*OklinkBrc20BalanceDetails = make([]*OklinkBrc20BalanceDetails, 0)
		err        error
		query map[string]string = map[string]string{
			"address":address,
			"token":tick,
			"page":strconv.FormatInt(page, 10),
			"limit":strconv.FormatInt(limit, 10),
		}
		headers map[string]string = map[string]string{
			"Ok-Access-Key":config.OklinkKey,
		}
	)

	url = fmt.Sprintf("%s/api/v5/explorer/btc/address-balance-details", config.OklinkDomain)
	result, err = tool.GetUrl(url, query, headers)
	if err != nil {
		return nil, err
	}
	fmt.Println(result)
	if err = tool.JsonToObject(result, &resp) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	if resp.Code != OklinkCodeSuccess {
		return nil, errors.New(fmt.Sprintf("Msg:%s", resp.Msg))
	}

	if err = tool.JsonToAny(resp.Data, &data) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}
	if len(data) == 0 {
		return nil, errors.New("No Data. ")
	}

	return data[0], nil
}

//Get Inscriptions
func GetInscriptions(token, inscriptionId, inscriptionNumber string, page, limit int64) (*OklinkInscriptionDetails, error) {
	var (
		url        string
		result        string
		resp        *OklinkResp
		data        []*OklinkInscriptionDetails = make([]*OklinkInscriptionDetails, 0)
		err        error
		query map[string]string = map[string]string{
			"token":token,
			"inscriptionId":inscriptionId,
			"inscriptionNumber":inscriptionNumber,
			"page":strconv.FormatInt(page, 10),
			"limit":strconv.FormatInt(limit, 10),
		}
		headers map[string]string = map[string]string{
			"Ok-Access-Key":config.OklinkKey,
		}
	)

	url = fmt.Sprintf("%s/api/v5/explorer/btc/inscriptions-list", config.OklinkDomain)
	result, err = tool.GetUrl(url, query, headers)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)
	if err = tool.JsonToObject(result, &resp) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	if resp.Code != OklinkCodeSuccess {
		return nil, errors.New(fmt.Sprintf("Msg:%s", resp.Msg))
	}

	if err = tool.JsonToAny(resp.Data, &data) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}
	if len(data) == 0 {
		return nil, errors.New("No Data. ")
	}

	return data[0], nil
}

//Get TxDetail
func GetTxDetail(txId string) (*TxDetail, error) {
	var (
		url        string
		result        string
		resp        *OklinkResp
		data        []*TxDetail = make([]*TxDetail, 0)
		err        error
		query map[string]string = map[string]string{
			"chainShortName":"btc",
			"txid":txId,
		}
		headers map[string]string = map[string]string{
			"Ok-Access-Key":config.OklinkKey,
		}
	)

	url = fmt.Sprintf("%s/api/v5/explorer/transaction/transaction-fills", config.OklinkDomain)
	result, err = tool.GetUrl(url, query, headers)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)
	if err = tool.JsonToObject(result, &resp) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	if resp.Code != OklinkCodeSuccess {
		return nil, errors.New(fmt.Sprintf("Msg:%s", resp.Msg))
	}

	if err = tool.JsonToAny(resp.Data, &data) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}
	if len(data) == 0 {
		return nil, errors.New("No Data. ")
	}

	return data[0], nil
}

func BroadcastTx(hex string) (*BroadcastTxResp, error) {
	var (
		url        string
		result        string
		resp        *OklinkResp
		data        []*BroadcastTxResp = make([]*BroadcastTxResp, 0)
		err        error
		req map[string]string = map[string]string{
			"chainShortName":"btc",
			"signedTx":hex,
		}
		headers map[string]string = map[string]string{
			"Ok-Access-Key":config.OklinkKey,
		}
	)

	fmt.Println(hex)
	url = fmt.Sprintf("%s/api/v5/explorer/transaction/publish-tx", config.OklinkDomain)
	result, err = tool.PostUrl(url, req, headers)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)
	if err = tool.JsonToObject(result, &resp) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	if resp.Code != OklinkCodeSuccess {
		return nil, errors.New(fmt.Sprintf("Msg:%s", resp.Msg))
	}

	if err = tool.JsonToAny(resp.Data, &data) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}
	if len(data) == 0 {
		return nil, errors.New("No Data. ")
	}
	return data[0], nil
}