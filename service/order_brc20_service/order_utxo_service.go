package order_brc20_service

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/create_key"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
	"strings"
)

func ColdDownUtxo(req *request.ColdDownUtxo) (string, error){
	var (
		netParams *chaincfg.Params = &chaincfg.MainNetParams
		err error
		fromPriKeyHex, fromTaprootAddress string = "", ""
		txRaw string = ""
		latestUtxo *model.OrderUtxoModel
		utxoList []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		startIndex int64 = 0
	)
	switch strings.ToLower(req.Net) {
	case "mainnet":
		netParams = &chaincfg.MainNetParams
		break
	case "signet":
		netParams = &chaincfg.SigNetParams
		break
	case "testnet":
		netParams = &chaincfg.TestNet3Params
		break
	}

	fromPriKeyHex, fromTaprootAddress, err = create_key.CreateTaprootKey(netParams)
	if err != nil {
		return "", err
	}

	latestUtxo, _ = mongo_service.GetLatestStartIndexUtxo(req.Net, req.UtxoType)
	if latestUtxo != nil {
		startIndex = latestUtxo.SortIndex
	}

	inputs := make([]*TxInputUtxo, 0)
	inputs = append(inputs, &TxInputUtxo{
		TxId:     req.TxId,
		TxIndex:  req.Index,
		PkScript: req.PkScript,
		Amount:   req.Amount,
		PriHex:   req.PriKeyHex,
	})


	//count := req.Amount/req.PerAmount
	outputs := make([]*TxOutput, 0)
	for i := int64(0); i < req.Count; i++ {
		outputs = append(outputs, &TxOutput{
			Address: fromTaprootAddress,
			Amount:  int64(req.PerAmount),
		})

		utxoList = append(utxoList, &model.OrderUtxoModel{
			//UtxoId:     "",
			Net:           req.Net,
			UtxoType:      req.UtxoType,
			Amount:        req.PerAmount,
			Address:       fromTaprootAddress,
			PrivateKeyHex: fromPriKeyHex,
			TxId:          "",
			Index:         i,
			PkScript:      "",
			UsedState:     model.UsedNo,
			//UseTx:      "",
			SortIndex: startIndex + i,
			Timestamp: tool.MakeTimestamp(),
		})
	}


	if req.ChangeAddress == "" {
		req.ChangeAddress = req.Address
	}
	tx, err := BuildCommonTx(netParams, inputs, outputs, req.ChangeAddress, req.FeeRate)
	if err != nil {
		return "", err
	}
	txRaw, err = ToRaw(tx)
	if err != nil {
		return "", err
	}
	for _, u := range utxoList {
		u.TxId = tx.TxHash().String()
		u.UtxoId = fmt.Sprintf("%s_%d", u.TxId, u.Index)

		_, err := mongo_service.SetOrderUtxoModel(u)
		if err != nil {
			major.Println(fmt.Sprintf("SetOrderUtxoModel for cold down err:%s", err.Error()))
			return "", nil
		}
	}

	txResp, err := oklink_service.BroadcastTx(txRaw)
	if err != nil {
		return "", err
	}
	return txResp.TxId, nil
}