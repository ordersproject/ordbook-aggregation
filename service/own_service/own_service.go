package own_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/config"
	"ordbook-aggregation/tool"
)

const (
	OwnServiceCodeSuccess int64 = 2000
)

func CheckUtxoInfo(outPoints []string) (map[string]*UtxoInfo, error) {
	var (
		url    string
		result string
		resp   *OwnServiceResp
		data   map[string]*UtxoInfo
		err    error
		req    map[string]interface{} = map[string]interface{}{
			"outPoints": outPoints,
		}
	)

	url = fmt.Sprintf("%s/tx/btc-utxo/check", config.OwnDomain)
	fmt.Println(url)
	result, err = tool.PostUrl(url, req, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)
	if err = tool.JsonToObject(result, &resp); err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	if resp.Code != OwnServiceCodeSuccess {
		return nil, errors.New(fmt.Sprintf("Msg:%v", resp.Data))
	}

	if err = tool.JsonToAny(resp.Data, &data); err != nil {
		return nil, errors.New(fmt.Sprintf("Get request err:%s", err))
	}

	return data, nil
}
