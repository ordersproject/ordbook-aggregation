package lp_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/inscription_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
)

// add one lp in step1
func AddOneLpStep1(req *request.LpAddOneStep1Request) (*respond.Brc20LpAddBatchStep1Resp, error) {
	var (
		netParams            *chaincfg.Params = order_brc20_service.GetNetParams(req.Net)
		_, platformAddressLp string           = order_brc20_service.GetPlatformKeyAndAddressForLp(req.Net)
		//_, platformAddressReceiveValueForAsk                              string           = GetPlatformKeyAndAddressReceiveValueForAsk(req.Net)
		transferContent                     string   = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, req.Tick, req.InscribeTransferAmount)
		commitTxHash                        string   = ""
		revealTxHashList, inscriptionIdList []string = make([]string, 0), make([]string, 0)
		lpOrderIdList                       []string = make([]string, 0)
		err                                 error
		brc20BalanceResult                  *oklink_service.OklinkBrc20BalanceDetails
		availableBalance                    int64                               = 0
		fees                                int64                               = 0
		inscribeUtxoList                    []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
		commonCoinAmount                    uint64                              = uint64(req.InscribeTransferAmount)
		brc20InValue                        int64                               = req.Brc20InValue
	)
	inscribeUtxoList = append(inscribeUtxoList, &inscription_service.InscribeUtxo{
		OutTx:     req.TxId,
		OutIndex:  req.Index,
		OutAmount: int64(req.Amount),
	})

	fmt.Println(transferContent)
	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressLp, req.Tick, 1, 50)
	if err != nil {
		return nil, err
	}
	availableBalance, _ = strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	fmt.Printf("availableBalance:%d, req.InscribeTransferAmount*req.Count: %d\n", availableBalance, req.InscribeTransferAmount*req.Count)
	if availableBalance < req.InscribeTransferAmount*req.Count {
		return nil, errors.New("AvailableBalance not enough. ")
	}
	commitTxHash, revealTxHashList, inscriptionIdList, fees, err =
		inscription_service.InscribeMultiDataFromUtxo(netParams, req.PriKeyHex, platformAddressLp,
			transferContent, req.FeeRate, req.ChangeAddress, req.Count, inscribeUtxoList, false, req.OutAddressType, req.IsOnlyCal, brc20InValue)
	if err != nil {
		return nil, err
	}

	if brc20InValue == 0 {
		brc20InValue = 546
	}

	if err != nil {
		return nil, err
	}
	for _, v := range inscriptionIdList {
		orderId := fmt.Sprintf("%s_%s_%s_%s_%d", req.Net, req.Tick, v, platformAddressLp, commonCoinAmount)
		orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
		lpOrder := &model.LpBrc20Model{
			Net:        req.Net,
			Tick:       req.Tick,
			OrderId:    orderId,
			Address:    platformAddressLp,
			FeeAddress: req.Address,
			//PoolOrderId:        "",
			Brc20InscriptionId: v,
			Brc20CoinAmount:    req.InscribeTransferAmount,
			Brc20InValue:       brc20InValue,
			Brc20ConfirmStatus: model.Unconfirmed,
			//BtcUtxoId:          "",
			//BtcAmount:          0,
			//BtcConfirmStatus:   0,
			PoolOrderState: 0,
			//CoinPrice:           coinPrice,
			//CoinPriceDecimalNum: coinPriceDecimalNum,
			Timestamp: tool.MakeTimestamp(),
		}
		_, err := mongo_service.SetLpBrc20Model(lpOrder)
		if err != nil {
			major.Println(fmt.Sprintf("SetLpBrc20Model err:%s", err.Error()))
			continue
		}
		lpOrderIdList = append(lpOrderIdList, orderId)
	}

	return &respond.Brc20LpAddBatchStep1Resp{
		Fees:              fees,
		CommitTxHash:      commitTxHash,
		RevealTxHashList:  revealTxHashList,
		InscriptionIdList: inscriptionIdList,
		LpOrderIdList:     lpOrderIdList,
	}, nil
}

// add one lp in step2
func AddOneLpStep2(req *request.LpAddOneStep2Request) (*respond.Brc20LpAddStep2Resp, error) {
	var (
		netParams         *chaincfg.Params
		orderEntity       *model.LpBrc20Model
		marketPrice       uint64 = 0
		btcAmount         int64  = 0
		platformAddressLp string
		txRaw             string
	)
	orderEntity, _ = mongo_service.FindLpBrc20ModelByOrderId(req.LpOrderId)
	if orderEntity == nil {
		return nil, errors.New("no order")
	}
	netParams = order_brc20_service.GetNetParams(orderEntity.Net)
	_, platformAddressLp = order_brc20_service.GetPlatformKeyAndAddressForLp(orderEntity.Net)
	//marketPrice = order_brc20_service.GetMarketPrice(orderEntity.Net, orderEntity.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(orderEntity.Tick)))
	//if marketPrice == 0 {
	//	return nil, errors.New("no marketPrice")
	//}
	marketPrice = 75

	switch req.Ratio {
	case 12, 15, 18:
		break
	default:
		return nil, errors.New("ratio error")
	}

	btcAmount = int64(marketPrice) * orderEntity.Brc20CoinAmount * req.Ratio / 10

	inputs := make([]*order_brc20_service.TxInputUtxo, 0)
	inputs = append(inputs, &order_brc20_service.TxInputUtxo{
		TxId:     req.TxId,
		TxIndex:  req.Index,
		PkScript: req.PkScript,
		Amount:   req.Amount,
		PriHex:   req.PriKeyHex,
	})

	outputs := make([]*order_brc20_service.TxOutput, 0)
	outputs = append(outputs, &order_brc20_service.TxOutput{
		Address: platformAddressLp,
		Amount:  btcAmount,
	})

	if req.ChangeAddress == "" {
		req.ChangeAddress = req.Address
	}
	tx, err := order_brc20_service.BuildCommonTx(netParams, inputs, outputs, req.ChangeAddress, req.FeeRate)
	if err != nil {
		fmt.Printf("[LP]BuildCommonTx err:%s\n", err.Error())
		return nil, err
	}
	txRaw, err = order_brc20_service.ToRaw(tx)
	txResp, err := unisat_service.BroadcastTx(orderEntity.Net, txRaw)
	if err != nil {
		fmt.Printf("[LP][%s] [%s]-BroadcastTx err:%s\n", "step2", orderEntity.Net, err.Error())
		return nil, err
	}
	txId := txResp.Result

	btcAmountDe := decimal.NewFromInt(btcAmount)
	coinAmountDe := decimal.NewFromInt(orderEntity.Brc20CoinAmount)
	poolCoinRatePrice := btcAmountDe.Div(coinAmountDe).IntPart()

	orderEntity.BtcUtxoId = fmt.Sprintf("%s_%d", txId, 0)
	orderEntity.BtcAmount = btcAmount
	orderEntity.BtcOutValue = req.BtcOutValue
	orderEntity.BtcConfirmStatus = model.Unconfirmed
	orderEntity.PoolCoinRatePrice = uint64(poolCoinRatePrice)
	orderEntity.CoinRatePrice = marketPrice
	orderEntity.Ratio = req.Ratio
	_, err = mongo_service.SetLpBrc20Model(orderEntity)
	if err != nil {
		major.Println(fmt.Sprintf("SetLpBrc20Model err:%s", err.Error()))
		return nil, err
	}

	return &respond.Brc20LpAddStep2Resp{
		Fees:          0,
		TxId:          txId,
		CoinPrice:     0,
		LpOrderId:     orderEntity.OrderId,
		BtcUtxoId:     orderEntity.BtcUtxoId,
		BtcAmount:     orderEntity.BtcAmount,
		CoinRatePrice: orderEntity.CoinRatePrice,
		Ratio:         orderEntity.Ratio,
	}, nil
}

func AddOneLpStep2Batch(req *request.LpAddOneStep2BatchRequest) (*respond.Brc20LpAddStep2BatchResp, error) {
	var (
		netParams         *chaincfg.Params
		marketPrice       uint64 = 0
		platformAddressLp string
		txRaw             string
		orderEntityList   []*model.LpBrc20Model                   = make([]*model.LpBrc20Model, 0)
		list              []*respond.Brc20LpAddStep2BatchItemResp = make([]*respond.Brc20LpAddStep2BatchItemResp, 0)
	)

	netParams = order_brc20_service.GetNetParams(req.Net)
	_, platformAddressLp = order_brc20_service.GetPlatformKeyAndAddressForLp(req.Net)
	//marketPrice = order_brc20_service.GetMarketPrice(orderEntity.Net, orderEntity.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(orderEntity.Tick)))
	//if marketPrice == 0 {
	//	return nil, errors.New("no marketPrice")
	//}
	//marketPrice = 75
	marketPrice = 280

	switch req.Ratio {
	case 12, 15, 18, 50, 100:
		break
	default:
		return nil, errors.New("ratio error")
	}

	for _, lpOrderId := range req.LpOrderIdList {
		orderEntity, _ := mongo_service.FindLpBrc20ModelByOrderId(lpOrderId)
		if orderEntity == nil {
			return nil, errors.New("no order")
		}
		btcAmount := int64(marketPrice) * orderEntity.Brc20CoinAmount * req.Ratio / 10
		orderEntity.BtcAmount = btcAmount
		orderEntityList = append(orderEntityList, orderEntity)
	}

	inputs := make([]*order_brc20_service.TxInputUtxo, 0)
	inputs = append(inputs, &order_brc20_service.TxInputUtxo{
		TxId:     req.TxId,
		TxIndex:  req.Index,
		PkScript: req.PkScript,
		Amount:   req.Amount,
		PriHex:   req.PriKeyHex,
	})

	outputs := make([]*order_brc20_service.TxOutput, 0)
	for _, orderEntity := range orderEntityList {
		outputs = append(outputs, &order_brc20_service.TxOutput{
			Address: platformAddressLp,
			Amount:  orderEntity.BtcAmount,
		})
	}

	if req.ChangeAddress == "" {
		req.ChangeAddress = req.Address
	}
	tx, err := order_brc20_service.BuildCommonTx(netParams, inputs, outputs, req.ChangeAddress, req.FeeRate)
	if err != nil {
		fmt.Printf("[LP]BuildCommonTx err:%s\n", err.Error())
		return nil, err
	}
	txRaw, err = order_brc20_service.ToRaw(tx)
	txResp, err := unisat_service.BroadcastTx(req.Net, txRaw)
	if err != nil {
		fmt.Printf("[LP][%s] [%s]-BroadcastTx err:%s\n", "step2 batch", req.Net, err.Error())
		return nil, err
	}
	txId := txResp.Result

	for i, orderEntity := range orderEntityList {
		btcAmountDe := decimal.NewFromInt(orderEntity.BtcAmount)
		coinAmountDe := decimal.NewFromInt(orderEntity.Brc20CoinAmount)
		poolCoinRatePrice := btcAmountDe.Div(coinAmountDe).IntPart()

		orderEntity.BtcUtxoId = fmt.Sprintf("%s_%d", txId, i)
		orderEntity.BtcOutValue = req.BtcOutValue
		orderEntity.BtcConfirmStatus = model.Unconfirmed
		orderEntity.PoolCoinRatePrice = uint64(poolCoinRatePrice)
		orderEntity.CoinRatePrice = marketPrice
		orderEntity.Ratio = req.Ratio
		_, err = mongo_service.SetLpBrc20Model(orderEntity)
		if err != nil {
			major.Println(fmt.Sprintf("SetLpBrc20Model err:%s", err.Error()))
			continue
		}

		list = append(list, &respond.Brc20LpAddStep2BatchItemResp{
			CoinPrice:     orderEntity.PoolCoinRatePrice,
			LpOrderId:     orderEntity.OrderId,
			BtcUtxoId:     orderEntity.BtcUtxoId,
			BtcAmount:     orderEntity.BtcAmount,
			CoinRatePrice: orderEntity.CoinRatePrice,
			Ratio:         orderEntity.Ratio,
		})
	}

	return &respond.Brc20LpAddStep2BatchResp{
		Fees: 0,
		TxId: txId,
		List: list,
	}, nil
}

func CancelOneLpBatch(req *request.LpCancelOneBatchRequest) (*respond.Brc20LpCancelBatchResp, error) {
	var (
		netParams                               *chaincfg.Params = order_brc20_service.GetNetParams(req.Net)
		platformPrivateKeyLp, platformAddressLp string           = order_brc20_service.GetPlatformKeyAndAddressForLp(req.Net)
		txRaw                                   string
		orderEntityList                         []*model.LpBrc20Model                 = make([]*model.LpBrc20Model, 0)
		list                                    []*respond.Brc20LpCancelBatchItemResp = make([]*respond.Brc20LpCancelBatchItemResp, 0)
		totalAmount                             int64                                 = 0
		txId                                    string                                = ""

		publicKeyLp string = tool.GetPublicKeyFromPrivateKey(platformPrivateKeyLp)
	)

	for _, lpOrderId := range req.LpOrderIdList {
		orderEntity, _ := mongo_service.FindLpBrc20ModelByOrderId(lpOrderId)
		if orderEntity == nil {
			return nil, errors.New("no order")
		}
		orderEntityList = append(orderEntityList, orderEntity)
	}

	addrPlatformLp, err := btcutil.DecodeAddress(platformAddressLp, netParams)

	if err != nil {
		return nil, err
	}
	pkScriptAddrPlatformLp, err := txscript.PayToAddrScript(addrPlatformLp)
	if err != nil {
		return nil, err
	}

	inputs := make([]*order_brc20_service.TxInputUtxo, 0)
	for _, orderEntity := range orderEntityList {
		btcUtxoId := orderEntity.BtcUtxoId
		btcUtxoIdStrs := strings.Split(btcUtxoId, "_")
		if len(btcUtxoIdStrs) != 2 {
			continue
		}
		btcUtxoTxId := btcUtxoIdStrs[0]
		btcUtxoTxIndex, _ := strconv.ParseInt(btcUtxoIdStrs[1], 10, 64)

		inputs = append(inputs, &order_brc20_service.TxInputUtxo{
			TxId:     btcUtxoTxId,
			TxIndex:  btcUtxoTxIndex,
			PkScript: hex.EncodeToString(pkScriptAddrPlatformLp),
			Amount:   uint64(orderEntity.BtcAmount),
			PriHex:   platformPrivateKeyLp,
		})
		totalAmount = totalAmount + orderEntity.BtcAmount

		list = append(list, &respond.Brc20LpCancelBatchItemResp{
			LpOrderId: orderEntity.OrderId,
			BtcUtxoId: orderEntity.BtcUtxoId,
			BtcAmount: orderEntity.BtcAmount,
		})
	}

	totalSize := int64(len(inputs))*order_brc20_service.SpendSize + 2*order_brc20_service.OutSize + order_brc20_service.OtherSize + 200
	totalFee := totalSize * int64(req.FeeRate)
	fmt.Printf("totalSize:%d, req.FeeRate:%d, totalFee:%d\n", totalSize, req.FeeRate, totalFee)
	outputs := make([]*order_brc20_service.TxOutput, 0)
	outputs = append(outputs, &order_brc20_service.TxOutput{
		Address: req.Address,
		Amount:  totalAmount - totalFee,
	})

	tx, err := order_brc20_service.BuildCommonTx(netParams, inputs, outputs, req.Address, req.FeeRate)
	if err != nil {
		fmt.Printf("[LP]BuildCommonTx err:%s\n", err.Error())
		return nil, err
	}
	tx.SerializeSize()
	txRaw, err = order_brc20_service.ToRaw(tx)
	if err != nil {
		fmt.Printf("[LP]ToRaw err:%s\n", err.Error())
		return nil, err
	}
	if !req.IsCalOnly {
		txResp, err := unisat_service.BroadcastTx(req.Net, txRaw)
		if err != nil {
			fmt.Printf("[LP][%s] [%s]-BroadcastTx err:%s\n", "cancel", req.Net, err.Error())
			return nil, err
		}
		txId = txResp.Result

		// remove lp
		for _, orderEntity := range orderEntityList {
			_, err := order_brc20_service.UpdatePoolOrder(&request.OrderPoolBrc20UpdateReq{
				Net:       req.Net,
				OrderId:   orderEntity.OrderId,
				PoolState: model.PoolStateRemove,
			}, publicKeyLp, "")
			if err != nil {
				fmt.Printf("[LP][cancel][%s] PushPoolOrder err:%s\n", orderEntity.OrderId, err.Error())
				continue
			}

			orderEntity.PoolOrderState = 0
			_, err = mongo_service.SetLpBrc20Model(orderEntity)
			if err != nil {
				fmt.Printf("[LP][cancel] [%s] SetLpBrc20Model err:%s\n", orderEntity.OrderId, err.Error())
				continue
			}
		}
	}

	fee := int64(tx.SerializeSize()) * int64(req.FeeRate)

	return &respond.Brc20LpCancelBatchResp{
		Fees:        fee,
		TxId:        txId,
		TotalAmount: totalAmount,
		List:        list,
	}, nil
}

func JobAddLp() {
	var (
		net                                     string = "livenet"
		jobName                                 string = "ADD-LP"
		entityList                              []*model.LpBrc20Model
		limit                                   int64            = 2000
		timestamp                               int64            = 0
		netParams                               *chaincfg.Params = order_brc20_service.GetNetParams(net)
		platformPrivateKeyLp, platformAddressLp string           = order_brc20_service.GetPlatformKeyAndAddressForLp(net)
		publicKeyLp                             string           = tool.GetPublicKeyFromPrivateKey(platformPrivateKeyLp)
	)

	entityList, _ = mongo_service.FindLpBrc20ModelList(limit, timestamp, 0)
	if entityList == nil || len(entityList) == 0 {
		return
	}
	for _, v := range entityList {
		brc20InscriptionId := v.Brc20InscriptionId
		brc20InscriptionTxId := ""
		brc20InscriptionTxIndex := int64(0)
		btcUtxoId := v.BtcUtxoId
		btcUtxoTxId := ""
		btcUtxoTxIndex := int64(0)
		isUpdate := false
		if brc20InscriptionId != "" {
			brc20InscriptionIdStrs := strings.Split(v.Brc20InscriptionId, "i")
			if len(brc20InscriptionIdStrs) != 2 {
				continue
			}
			brc20InscriptionTxId = brc20InscriptionIdStrs[0]
			brc20InscriptionTxIndex, _ = strconv.ParseInt(brc20InscriptionIdStrs[1], 10, 64)

			if v.Brc20ConfirmStatus == model.Unconfirmed {
				block := order_brc20_service.GetTxBlock(brc20InscriptionTxId)
				if block != 0 {
					v.Brc20ConfirmStatus = model.Confirmed
					isUpdate = true
				}
			}
		}

		if btcUtxoId != "" {
			btcUtxoIdStrs := strings.Split(v.BtcUtxoId, "_")
			if len(btcUtxoIdStrs) != 2 {
				continue
			}
			btcUtxoTxId = btcUtxoIdStrs[0]
			btcUtxoTxIndex, _ = strconv.ParseInt(btcUtxoIdStrs[1], 10, 64)

			if v.BtcConfirmStatus == model.Unconfirmed {
				block := order_brc20_service.GetTxBlock(btcUtxoTxId)
				if block != 0 {
					v.BtcConfirmStatus = model.Confirmed
					isUpdate = true
				}
			}
		}

		if isUpdate {
			_, err := mongo_service.SetLpBrc20Model(v)
			if err != nil {
				fmt.Printf("[LP][%s] [%s] SetLpBrc20Model err:%s\n", jobName, v.OrderId, err.Error())
				continue
			}
		}

		if !(v.Brc20ConfirmStatus == model.Confirmed && v.BtcConfirmStatus == model.Confirmed) {
			continue
		}

		// add pool order
		//1.Fetch pool platform PubKey and Address
		//2.generate multisig address
		//3.generate psbt
		//4.Push pool order to pool
		platformKeyResp, err := order_brc20_service.FetchPoolPlatformPublicKey(&request.PoolBrc20PushReq{
			Net: "livenet",
		})
		if err != nil {
			fmt.Printf("[LP][%s] [%s]-[%d] fetch pool platform publicKey err:%s\n", jobName, v.OrderId, v.Brc20CoinAmount, err.Error())
			continue
		}
		brc20PublicKey := platformKeyResp.PublicKey
		btcPublicKey := platformKeyResp.BtcPublicKey

		multiSigScript, multiSigAddress, multiSigSegWitAddress, err := order_brc20_service.CreateMultiSigAddress(netParams, []string{publicKeyLp, brc20PublicKey}...)
		if err != nil {
			fmt.Printf("[LP][%s] [%s] Create Brc20 MultiSigAddress err:%s\n", jobName, v.OrderId, err.Error())
			continue
		}
		_ = multiSigScript
		_ = multiSigSegWitAddress
		_ = multiSigAddress

		multiSigScriptBtc, multiSigAddressBtc, multiSigSegWitAddressBtc, err := order_brc20_service.CreateMultiSigAddress(netParams, []string{publicKeyLp, btcPublicKey}...)
		if err != nil {
			fmt.Printf("[LP][%s] [%s] Create btc MultiSigAddress err:%s\n", jobName, v.OrderId, err.Error())
			continue
		}
		_ = multiSigScriptBtc
		_ = multiSigSegWitAddressBtc
		_ = multiSigAddressBtc

		coinPsbtRaw, err := makeLpPsbt(netParams,
			platformPrivateKeyLp, platformAddressLp, brc20InscriptionTxId, brc20InscriptionTxIndex, v.Brc20InValue,
			multiSigSegWitAddress, v.BtcAmount)
		if err != nil {
			fmt.Printf("[LP][%s] [%s] Create Brc20 Psbt err:%s\n", jobName, v.OrderId, err.Error())
			continue
		}

		psbtRaw, err := makeLpPsbt(netParams,
			platformPrivateKeyLp, platformAddressLp, btcUtxoTxId, btcUtxoTxIndex, v.BtcAmount,
			multiSigSegWitAddressBtc, v.BtcOutValue)
		if err != nil {
			fmt.Printf("[LP][%s] [%s] Create Btc Psbt err:%s\n", jobName, v.OrderId, err.Error())
			continue
		}

		pushPoolResp, err := order_brc20_service.PushPoolOrder(&request.PoolBrc20PushReq{
			Net:         v.Net,
			Tick:        v.Tick,
			Pair:        fmt.Sprintf("%s-BTC", strings.ToUpper(v.Tick)),
			PoolType:    model.PoolTypeBoth,
			PoolState:   model.PoolStateAdd,
			Address:     v.Address,
			CoinPsbtRaw: coinPsbtRaw,
			CoinAmount:  uint64(v.Brc20CoinAmount),
			PsbtRaw:     psbtRaw,
			BtcPoolMode: model.PoolModePsbt,
			//BtcUtxoId:   "",
			Amount: uint64(v.BtcAmount),
			Ratio:  v.Ratio,
		}, publicKeyLp)
		if err != nil {
			fmt.Printf("[LP][%s] [%s] PushPoolOrder err:%s\n", jobName, v.OrderId, err.Error())
			continue
		}
		v.PoolOrderId = pushPoolResp
		poolOrder, _ := mongo_service.FindPoolBrc20ModelByOrderId(v.PoolOrderId)
		if poolOrder != nil {
			v.PoolOrderState = poolOrder.PoolState
			v.PoolOrderCoinState = poolOrder.PoolCoinState
			_, err = mongo_service.SetLpBrc20Model(v)
			if err != nil {
				fmt.Printf("[LP][%s] [%s] SetLpBrc20Model err:%s\n", jobName, v.OrderId, err.Error())
				continue
			}
		}
		major.Println(fmt.Sprintf("[LP][%s] [%s] add lp success", jobName, v.OrderId))
	}
}

// make lp psbt
func makeLpPsbt(netParams *chaincfg.Params, inPrivateKey, inAddress, inTxId string, inTxIndex, inValue int64,
	outAddress string, outValue int64) (string, error) {
	var (
		builder *order_brc20_service.PsbtBuilder
		psbtRaw string = ""
		err     error
	)
	addr, err := btcutil.DecodeAddress(inAddress, netParams)
	if err != nil {
		return "", err
	}
	pkScript, err := txscript.PayToAddrScript(addr)

	inputs := make([]order_brc20_service.Input, 0)
	inputs = append(inputs, order_brc20_service.Input{
		OutTxId:  inTxId,
		OutIndex: uint32(inTxIndex),
	})

	outputs := make([]order_brc20_service.Output, 0)
	outputs = append(outputs, order_brc20_service.Output{
		Address: outAddress,
		Amount:  uint64(outValue),
	})
	inputSigns := make([]*order_brc20_service.InputSign, 0)

	inputSigns = append(inputSigns, &order_brc20_service.InputSign{
		Index:       0,
		OutRaw:      "",
		PkScript:    hex.EncodeToString(pkScript),
		SighashType: txscript.SigHashSingle | txscript.SigHashAnyOneCanPay,
		PriHex:      inPrivateKey,
		UtxoType:    order_brc20_service.Witness,
		Amount:      uint64(inValue),
	})

	builder, err = order_brc20_service.CreatePsbtBuilder(netParams, inputs, outputs)
	if err != nil {
		return "", err
	}
	err = builder.UpdateAndSignInput(inputSigns)
	if err != nil {
		return "", err
	}
	psbtRaw, err = builder.ToString()
	if err != nil {
		return "", err
	}
	return psbtRaw, nil
}

func loopReleaseLp() {

}
