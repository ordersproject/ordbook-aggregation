package task

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/create_key"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
)

//monitor 2 address:
//platformAddressReceiveBidValue
//platformAddressReceiveDummyValue  - 1200

const (
	MaxTotalUtxoAmount int64 = 50000
)

func LoopCheckPlatformAddressForBidValue(net string) {
	var (
		LoopName                                                          string = "BidValue"
		platformPrivateKeyReceiveBidValue, platformAddressReceiveBidValue string = order_brc20_service.GetPlatformKeyAndAddressReceiveBidValue(net)
		err                                                               error
		utxoResp                                                          *oklink_service.OklinkUtxoDetails
		totalUtxoAmount                                                   int64                   = 0
		netParams                                                         *chaincfg.Params        = order_brc20_service.GetNetParams(net)
		fromPriKeyHex, fromSegwitAddress                                  string                  = "", ""
		txRaw                                                             string                  = ""
		utxoList                                                          []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		startIndex                                                        int64                   = order_brc20_service.GetSaveStartIndex(net, model.UtxoTypeBidY)
		changeAddress                                                     string                  = platformAddressReceiveBidValue
		count                                                             int64                   = 0
		//todo 5w/utxo 10w/utxo, 20w/utxo, 50w/utxo, 100w/utxo
		perAmount         uint64        = 10000
		feeRate           int64         = 20
		totalSize         int64         = 0
		utxoInterfaceList []interface{} = make([]interface{}, 0)
	)
	utxoResp, err = oklink_service.GetAddressUtxo(platformAddressReceiveBidValue, 1, 50)
	if err != nil {
		fmt.Printf("[LOOP][%s]address utxo list err:%s\n", LoopName, err.Error())
		return
	}
	if utxoResp.UtxoList == nil || len(utxoResp.UtxoList) == 0 {
		fmt.Printf("[LOOP][%s]address utxo list: empty, waiting for next time\n", LoopName)
		return
	}

	for _, v := range utxoResp.UtxoList {
		amountDe, _ := decimal.NewFromString(v.UnspentAmount)
		amount := amountDe.Mul(decimal.New(1, 8)).IntPart()
		totalUtxoAmount = totalUtxoAmount + amount
	}

	if totalUtxoAmount <= MaxTotalUtxoAmount {
		fmt.Printf("[LOOP][%s]address utxo list: totalUtxoAmount[%d], not enough, waiting for next time\n", LoopName, totalUtxoAmount)
		return
	}

	fromPriKeyHex, fromSegwitAddress, err = create_key.CreateSegwitKey(netParams)
	if err != nil {
		fmt.Printf("[LOOP][%s] CreateSegwitKey err:%s\n", LoopName, err.Error())
		return
	}

	inputs := make([]*order_brc20_service.TxInputUtxo, 0)
	for _, v := range utxoResp.UtxoList {
		amountDe, _ := decimal.NewFromString(v.UnspentAmount)
		amount := amountDe.Mul(decimal.New(1, 8)).IntPart()
		addr, err := btcutil.DecodeAddress(v.Address, netParams)
		if err != nil {
			fmt.Printf("[LOOP][%s] DecodeAddress in input err:%s\n", LoopName, err.Error())
			return
		}
		pkScriptByte, err := txscript.PayToAddrScript(addr)
		if err != nil {
			fmt.Printf("[LOOP][%s] PayToAddrScript in input err:%s\n", LoopName, err.Error())
			return
		}
		txIndex, _ := strconv.ParseInt(v.Index, 10, 64)
		inputs = append(inputs, &order_brc20_service.TxInputUtxo{
			TxId:     v.TxId,
			TxIndex:  txIndex,
			PkScript: hex.EncodeToString(pkScriptByte),
			Amount:   uint64(amount),
			PriHex:   platformPrivateKeyReceiveBidValue,
		})
	}

	addr, err := btcutil.DecodeAddress(fromSegwitAddress, netParams)
	if err != nil {
		fmt.Printf("[LOOP][%s] DecodeAddress in output err:%s\n", LoopName, err.Error())
		return
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		fmt.Printf("[LOOP][%s] PayToAddrScript in output err:%s\n", LoopName, err.Error())
		return
	}
	pkScript := hex.EncodeToString(pkScriptByte)
	outputs := make([]*order_brc20_service.TxOutput, 0)

	totalSize = int64(len(inputs))*order_brc20_service.SpendSize + count*order_brc20_service.OutSize + order_brc20_service.OtherSize
	for totalSize*feeRate+count*int64(perAmount) < totalUtxoAmount {
		count++
		totalSize = int64(len(inputs))*order_brc20_service.SpendSize + count*order_brc20_service.OutSize + order_brc20_service.OtherSize
		fmt.Printf("[Cal][%s]inLen:%d, outLen:%d, totalSize:%d, totalFee:%d, totalOutAmount:%d, totalInAmount:%d\n", LoopName, len(inputs), count, totalSize, totalSize*feeRate, totalUtxoAmount, totalUtxoAmount)
	}
	count = count - 1
	fmt.Printf("[Final][%s]inLen:%d, outLen:%d, totalSize:%d, totalFee:%d, totalOutAmount:%d, totalInAmount:%d\n", LoopName, len(inputs), count, totalSize, totalSize*feeRate, totalUtxoAmount, totalUtxoAmount)

	if count <= 0 {
		fmt.Printf("[LOOP][%s] Count of outputs is 0, not enough, waiting for next time\n", LoopName)
		return
	}

	for i := int64(0); i < count; i++ {
		outputs = append(outputs, &order_brc20_service.TxOutput{
			Address: fromSegwitAddress,
			Amount:  int64(perAmount),
		})

		utxoList = append(utxoList, &model.OrderUtxoModel{
			Net:           net,
			UtxoType:      model.UtxoTypeBidY,
			Amount:        perAmount,
			Address:       fromSegwitAddress,
			PrivateKeyHex: fromPriKeyHex,
			TxId:          "",
			Index:         i,
			PkScript:      pkScript,
			UsedState:     model.UsedNo,
			SortIndex:     startIndex + i + 1,
			Timestamp:     tool.MakeTimestamp(),
		})
	}

	tx, err := order_brc20_service.BuildCommonTx(netParams, inputs, outputs, changeAddress, feeRate)
	if err != nil {
		fmt.Printf("[LOOP][%s]BuildCommonTx err:%s\n", LoopName, err.Error())
		return
	}
	txRaw, err = order_brc20_service.ToRaw(tx)
	if err != nil {
		fmt.Printf("[LOOP][%s]ToRaw err:%s\n", LoopName, err.Error())
		return
	}
	for _, u := range utxoList {
		u.TxId = tx.TxHash().String()
		u.UtxoId = fmt.Sprintf("%s_%d", u.TxId, u.Index)

		//_, err := mongo_service.SetOrderUtxoModel(u)
		//if err != nil {
		//	major.Println(fmt.Sprintf("SetOrderUtxoModel for cold down err:%s", err.Error()))
		//	return
		//}

		utxoInterfaceList = append(utxoInterfaceList, u)
	}

	txId := ""
	sendJop := func() error {
		//if net == "testnet" {
		//	txResp, err := mempool_space_service.BroadcastTx(net, txRaw)
		//	if err != nil {
		//		fmt.Printf("[LOOP][%s] testnet-BroadcastTx err:%s\n", LoopName, err.Error())
		//		return err
		//	}
		//	txId = txResp
		//}else {
		//	txResp, err := oklink_service.BroadcastTx(txRaw)
		//	if err != nil {
		//		fmt.Printf("[LOOP][%s] mainnet-BroadcastTx err:%s\n", LoopName, err.Error())
		//		return err
		//	}
		//	txId = txResp.
		//}

		txResp, err := unisat_service.BroadcastTx(net, txRaw)
		if err != nil {
			fmt.Printf("[LOOP][%s] [%s]-BroadcastTx err:%s\n", LoopName, net, err.Error())
			return err
		}
		txId = txResp.Result

		//txResp, err := node.BroadcastTx(net, txRaw)
		//if err != nil {
		//	fmt.Printf("[LOOP][%s] [%s]-BroadcastTx err:%s\n", LoopName, net, err.Error())
		//	return err
		//}
		//txId = txResp

		return nil
	}

	err = mongo_service.SetManyUtxoInSession(utxoList, sendJop)
	if err != nil {
		fmt.Printf("[LOOP][%s]SetManyUtxoInSession in send err:%s\n", LoopName, err.Error())
		return
	}

	fmt.Printf("[LOOP][%s] Replenish Utxo success, txId:%s\n", LoopName, txId)
}

func LoopCheckPlatformAddressForDummyValue(net string) {
	var (
		LoopName                                                              string = "DummyValue"
		platformPrivateKeyReceiveDummyValue, platformAddressReceiveDummyValue string = order_brc20_service.GetPlatformKeyAndAddressReceiveDummyValue(net)
		err                                                                   error
		utxoResp                                                              *oklink_service.OklinkUtxoDetails
		totalUtxoAmount                                                       int64                   = 0
		netParams                                                             *chaincfg.Params        = order_brc20_service.GetNetParams(net)
		fromPriKeyHex, fromSegwitAddress                                      string                  = "", ""
		txRaw                                                                 string                  = ""
		utxoList                                                              []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		startIndex                                                            int64                   = order_brc20_service.GetSaveStartIndex(net, model.UtxoTypeBidY)
		changeAddress                                                         string                  = platformAddressReceiveDummyValue
		count                                                                 int64                   = 0
		perAmount                                                             uint64                  = 10000
		feeRate                                                               int64                   = 20
		totalSize                                                             int64                   = 0
		utxoInterfaceList                                                     []interface{}           = make([]interface{}, 0)
	)
	utxoResp, err = oklink_service.GetAddressUtxo(platformAddressReceiveDummyValue, 1, 50)
	if err != nil {
		fmt.Printf("[LOOP][%s]address utxo list err:%s\n", LoopName, err.Error())
		return
	}
	if utxoResp.UtxoList == nil || len(utxoResp.UtxoList) == 0 {
		fmt.Printf("[LOOP][%s]address utxo list: empty, waiting for next time\n", LoopName)
		return
	}

	for _, v := range utxoResp.UtxoList {
		amountDe, _ := decimal.NewFromString(v.UnspentAmount)
		amount := amountDe.Mul(decimal.New(1, 8)).IntPart()
		totalUtxoAmount = totalUtxoAmount + amount
	}

	if totalUtxoAmount <= MaxTotalUtxoAmount {
		fmt.Printf("[LOOP][%s]address utxo list: totalUtxoAmount[%d], Not enough, waiting for next time\n", LoopName, totalUtxoAmount)
		return
	}

	fromPriKeyHex, fromSegwitAddress, err = create_key.CreateSegwitKey(netParams)
	if err != nil {
		fmt.Printf("[LOOP][%s] CreateSegwitKey err:%s\n", LoopName, err.Error())
		return
	}

	inputs := make([]*order_brc20_service.TxInputUtxo, 0)
	for _, v := range utxoResp.UtxoList {
		amountDe, _ := decimal.NewFromString(v.UnspentAmount)
		amount := amountDe.Mul(decimal.New(1, 8)).IntPart()
		addr, err := btcutil.DecodeAddress(v.Address, netParams)
		if err != nil {
			fmt.Printf("[LOOP][%s] DecodeAddress in input err:%s\n", LoopName, err.Error())
			return
		}
		pkScriptByte, err := txscript.PayToAddrScript(addr)
		if err != nil {
			fmt.Printf("[LOOP][%s] PayToAddrScript in input err:%s\n", LoopName, err.Error())
			return
		}
		txIndex, _ := strconv.ParseInt(v.Index, 10, 64)
		inputs = append(inputs, &order_brc20_service.TxInputUtxo{
			TxId:     v.TxId,
			TxIndex:  txIndex,
			PkScript: hex.EncodeToString(pkScriptByte),
			Amount:   uint64(amount),
			PriHex:   platformPrivateKeyReceiveDummyValue,
		})
	}

	addr, err := btcutil.DecodeAddress(fromSegwitAddress, netParams)
	if err != nil {
		fmt.Printf("[LOOP][%s] DecodeAddress in output err:%s\n", LoopName, err.Error())
		return
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		fmt.Printf("[LOOP][%s] PayToAddrScript in output err:%s\n", LoopName, err.Error())
		return
	}
	pkScript := hex.EncodeToString(pkScriptByte)
	outputs := make([]*order_brc20_service.TxOutput, 0)

	totalSize = int64(len(inputs))*order_brc20_service.SpendSize + count*order_brc20_service.OutSize + order_brc20_service.OtherSize
	for totalSize*feeRate+count*int64(perAmount) < totalUtxoAmount {
		count++
		totalSize = int64(len(inputs))*order_brc20_service.SpendSize + count*order_brc20_service.OutSize + order_brc20_service.OtherSize
		fmt.Printf("[Cal][%s]inLen:%d, outLen:%d, totalSize:%d\n, totalFee:%d, totalOutAmount:%d, totalInAmount:%d\n", LoopName, len(inputs), count, totalSize, totalSize*feeRate, totalUtxoAmount, totalUtxoAmount)
	}
	count = count - 1
	fmt.Printf("[Final][%s]inLen:%d, outLen:%d, totalSize:%d\n, totalFee:%d, totalOutAmount:%d, totalInAmount:%d\n", LoopName, len(inputs), count, totalSize, totalSize*feeRate, totalUtxoAmount, totalUtxoAmount)

	if count <= 0 {
		fmt.Printf("[LOOP][%s] Count of outputs is 0, not enough, waiting for next time\n", LoopName)
		return
	}

	for i := int64(0); i < count; i++ {
		outputs = append(outputs, &order_brc20_service.TxOutput{
			Address: fromSegwitAddress,
			Amount:  int64(perAmount),
		})

		utxoList = append(utxoList, &model.OrderUtxoModel{
			Net:           net,
			UtxoType:      model.UtxoTypeBidY,
			Amount:        perAmount,
			Address:       fromSegwitAddress,
			PrivateKeyHex: fromPriKeyHex,
			TxId:          "",
			Index:         i,
			PkScript:      pkScript,
			UsedState:     model.UsedNo,
			SortIndex:     startIndex + i + 1,
			Timestamp:     tool.MakeTimestamp(),
		})
	}

	tx, err := order_brc20_service.BuildCommonTx(netParams, inputs, outputs, changeAddress, feeRate)
	if err != nil {
		fmt.Printf("[LOOP][%s]BuildCommonTx err:%s\n", LoopName, err.Error())
		return
	}
	txRaw, err = order_brc20_service.ToRaw(tx)
	if err != nil {
		fmt.Printf("[LOOP][%s]ToRaw err:%s\n", LoopName, err.Error())
		return
	}
	for _, u := range utxoList {
		u.TxId = tx.TxHash().String()
		u.UtxoId = fmt.Sprintf("%s_%d", u.TxId, u.Index)

		//_, err := mongo_service.SetOrderUtxoModel(u)
		//if err != nil {
		//	major.Println(fmt.Sprintf("SetOrderUtxoModel for cold down err:%s", err.Error()))
		//	return
		//}

		utxoInterfaceList = append(utxoInterfaceList, u)
	}

	txId := ""
	sendJop := func() error {
		//if net == "testnet" {
		//	txResp, err := mempool_space_service.BroadcastTx(net, txRaw)
		//	if err != nil {
		//		fmt.Printf("[LOOP][%s] testnet-BroadcastTx err:%s\n", LoopName, err.Error())
		//		return err
		//	}
		//	txId = txResp
		//}else {
		//	txResp, err := oklink_service.BroadcastTx(txRaw)
		//	if err != nil {
		//		fmt.Printf("[LOOP][%s] mainnet-BroadcastTx err:%s\n", LoopName, err.Error())
		//		return err
		//	}
		//	txId = txResp.TxId
		//}

		txResp, err := unisat_service.BroadcastTx(net, txRaw)
		if err != nil {
			fmt.Printf("[LOOP][%s] [%s]-BroadcastTx err:%s\n", LoopName, net, err.Error())
			return err
		}
		txId = txResp.Result

		//txResp, err := node.BroadcastTx(net, txRaw)
		//if err != nil {
		//	fmt.Printf("[LOOP][%s] [%s]-BroadcastTx err:%s\n", LoopName, net, err.Error())
		//	return err
		//}
		//txId = txResp
		return nil
	}

	err = mongo_service.SetManyUtxoInSession(utxoList, sendJop)
	if err != nil {
		fmt.Printf("[LOOP][%s]SetManyUtxoInSession in send err:%s\n", LoopName, err.Error())
		return
	}

	fmt.Printf("[LOOP][%s] Replenish Utxo success, txId:%s\n", LoopName, txId)
}
