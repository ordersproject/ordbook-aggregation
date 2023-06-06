package order_brc20_service

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/create_key"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
)

func ColdDownUtxo(req *request.ColdDownUtxo) (string, error){
	var (
		netParams *chaincfg.Params = GetNetParams(req.Net)
		err error
		fromPriKeyHex, fromSegwitAddress string = "", ""
		txRaw string = ""
		latestUtxo *model.OrderUtxoModel
		utxoList []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		startIndex int64 = 0
	)

	fromPriKeyHex, fromSegwitAddress, err = create_key.CreateSegwitKey(netParams)
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
	addr, err := btcutil.DecodeAddress(fromSegwitAddress, netParams)
	if err != nil {
		return "", err
	}
	addrHash, err := btcutil.NewAddressWitnessPubKeyHash(addr.ScriptAddress(), netParams)
	if err != nil {
		fmt.Printf("NewAddressWitnessPubKeyHash err: %s\n", err.Error())
		return "", err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addrHash)
	if err != nil {
		return "", err
	}
	pkScript := hex.EncodeToString(pkScriptByte)
	//count := req.Amount/req.PerAmount
	outputs := make([]*TxOutput, 0)
	for i := int64(0); i < req.Count; i++ {
		outputs = append(outputs, &TxOutput{
			Address: fromSegwitAddress,
			Amount:  int64(req.PerAmount),
		})

		utxoList = append(utxoList, &model.OrderUtxoModel{
			//UtxoId:     "",
			Net:           req.Net,
			UtxoType:      req.UtxoType,
			Amount:        req.PerAmount,
			Address:       fromSegwitAddress,
			PrivateKeyHex: fromPriKeyHex,
			TxId:          "",
			Index:         i,
			PkScript:      pkScript,
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
		fmt.Printf("BuildCommonTx err:%s\n", err.Error())
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

func saveNewDummyFromBid(net string, out Output, priKeyHex string, index int64, txId string) error {
	startIndex := int64(0)
	latestUtxo, _ := mongo_service.GetLatestStartIndexUtxo(net, model.UtxoTypeDummy)
	if latestUtxo != nil {
		startIndex = latestUtxo.SortIndex
	}
	netParams := GetNetParams(net)
	addr, err := btcutil.DecodeAddress(out.Address, netParams)
	if err != nil {
		return err
	}
	addrHash, err := btcutil.NewAddressPubKeyHash(addr.ScriptAddress(), netParams)
	if err != nil {
		return err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addrHash)
	if err != nil {
		return err
	}
	pkScript := hex.EncodeToString(pkScriptByte)

	newDummy := &model.OrderUtxoModel{
		UtxoId:        fmt.Sprintf("%s_%d", txId, index),
		Net:           net,
		UtxoType:      model.UtxoTypeDummy,
		Amount:        out.Amount,
		Address:       out.Address,
		PrivateKeyHex: priKeyHex,
		TxId:          txId,
		Index:         index,
		PkScript:      pkScript,
		UsedState:     model.UsedNo,
		SortIndex:     startIndex + 1,
		Timestamp:     tool.MakeTimestamp(),
	}

	_, err = mongo_service.SetOrderUtxoModel(newDummy)
	if err != nil {
		major.Println(fmt.Sprintf("SetOrderUtxoModel from bid err:%s", err.Error()))
		return nil
	}
	return nil
}