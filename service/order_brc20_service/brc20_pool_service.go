package order_brc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
)

func FetchPoolPairInfo(req *request.PoolPairFetchOneReq) (*respond.PoolInfoResponse, error) {
	var (
		total      int64 = 0
		entityList []*model.PoolInfoModel
		//entity *model.PoolInfoModel
		list []*respond.PoolInfoItem = make([]*respond.PoolInfoItem, 0)
	)

	total, _ = mongo_service.CountPoolInfoModelList(req.Net, req.Tick, req.Pair)
	entityList, _ = mongo_service.FindPoolInfoModelList(req.Net, req.Tick, req.Pair)
	for _, v := range entityList {
		item := &respond.PoolInfoItem{
			Net:            v.Net,
			Tick:           v.Tick,
			Pair:           v.Pair,
			CoinAmount:     v.CoinAmount,
			CoinDecimalNum: v.CoinDecimalNum,
			Amount:         v.Amount,
			DecimalNum:     v.DecimalNum,
		}
		list = append(list, item)
	}
	return &respond.PoolInfoResponse{
		Total:   total,
		Results: list,
		Flag:    0,
	}, nil
}

func FetchOnePoolPairInfo(req *request.PoolPairFetchOneReq) (*respond.PoolInfoItem, error) {
	var (
		entity          *model.PoolInfoModel
		coinAmountTotal uint64 = 0
		amountTotal     uint64 = 0
		count           uint64 = 0
	)
	entity, _ = mongo_service.FindPoolInfoModelByPair(req.Net, strings.ToUpper(req.Pair))
	if entity == nil || entity.Id == 0 {
		return nil, errors.New("pool info ie empty")
	}

	coinAmountTotal, amountTotal, count, _ = getOwnPoolInfo(req.Net, req.Tick, strings.ToUpper(req.Pair), req.Address)
	return &respond.PoolInfoItem{
		Net:            entity.Net,
		Tick:           entity.Tick,
		Pair:           entity.Pair,
		CoinAmount:     entity.CoinAmount,
		CoinDecimalNum: entity.CoinDecimalNum,
		Amount:         entity.Amount,
		DecimalNum:     entity.DecimalNum,
		OwnCoinAmount:  coinAmountTotal,
		OwnAmount:      amountTotal,
		OwnCount:       count,
	}, nil

}

func FetchPoolOrders(req *request.PoolBrc20FetchReq) (*respond.PoolResponse, error) {
	var (
		entityList []*model.PoolBrc20Model
		total      int64                    = 0
		list       []*respond.PoolBrc20Item = make([]*respond.PoolBrc20Item, 0)
		flag       int64                    = 0
	)
	total, _ = mongo_service.CountPoolBrc20ModelList(req.Net, req.Tick, req.Pair, req.Address, req.PoolType, req.PoolState)
	entityList, _ = mongo_service.FindPoolBrc20ModelList(req.Net, req.Tick, req.Pair, req.Address, req.PoolType, req.PoolState,
		req.Limit, req.Flag, req.Page, req.SortKey, req.SortType)
	for _, v := range entityList {
		item := &respond.PoolBrc20Item{
			Net:            v.Net,
			OrderId:        v.OrderId,
			Tick:           v.Tick,
			Pair:           v.Pair,
			CoinAmount:     v.CoinAmount,
			CoinDecimalNum: v.CoinDecimalNum,
			CoinAddress:    v.CoinAddress,
			Amount:         v.Amount,
			DecimalNum:     v.DecimalNum,
			Address:        v.Address,
			PoolType:       v.PoolType,
			PoolState:      v.PoolState,
			InscriptionId:  v.InscriptionId,
			//PsbtRaw:       v.PsbtRaw,
			Timestamp: v.Timestamp,
		}
		if req.SortKey == "todo" {
			//flag = int64(v.CoinRatePrice)
		} else {
			flag = v.Timestamp
		}
		list = append(list, item)
	}
	return &respond.PoolResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}

func FetchOnePoolOrder(req *request.PoolBrc20FetchOneReq, publicKey, ip string) (*respond.PoolBrc20Item, error) {
	var (
		entity    *model.PoolBrc20Model
		netParams *chaincfg.Params = GetNetParams(req.Net)
	)
	entity, _ = mongo_service.FindPoolBrc20ModelByOrderId(req.OrderId)
	if entity == nil {
		return nil, errors.New("Order is empty. ")
	}
	netParams = GetNetParams(entity.Net)

	verified, err := CheckPublicKeyAddress(netParams, publicKey, req.Address)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
	}
	if !verified {
		return nil, errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
	}

	item := &respond.PoolBrc20Item{
		Net:            entity.Net,
		OrderId:        entity.OrderId,
		Tick:           entity.Tick,
		CoinAmount:     entity.CoinAmount,
		CoinDecimalNum: entity.CoinDecimalNum,
		CoinAddress:    entity.CoinAddress,
		Amount:         entity.Amount,
		DecimalNum:     entity.DecimalNum,
		Address:        entity.Address,
		PoolType:       entity.PoolType,
		PoolState:      entity.PoolState,
		InscriptionId:  entity.InscriptionId,
		CoinPsbtRaw:    entity.CoinPsbtRaw,
		PsbtRaw:        entity.PsbtRaw,
		Timestamp:      entity.Timestamp,
	}
	return item, nil
}

func FetchPoolPlatformPublicKey(req *request.PoolBrc20PushReq) (*respond.PoolKeyInfoResp, error) {
	var (
		_, platformPublicKeyMultiSig = GetPlatformKeyMultiSig(req.Net)
	)
	return &respond.PoolKeyInfoResp{
		Net:       req.Net,
		PublicKey: platformPublicKeyMultiSig,
	}, nil
}

func PushPoolOrder(req *request.PoolBrc20PushReq, publicKey string) (string, error) {
	var (
		netParams         *chaincfg.Params = GetNetParams(req.Net)
		entity            *model.PoolBrc20Model
		err               error
		orderId           string = ""
		psbtBuilder       *PsbtBuilder
		coinAddress       string = ""
		coinAmount        uint64 = 0
		coinDec           int    = 18
		outAmount         uint64 = 0
		amountDec         int    = 8
		coinRatePrice     uint64 = 0
		inscriptionId     string = ""
		inscriptionNumber string = ""

		_, platformPublicKeyMultiSig             = GetPlatformKeyMultiSig(req.Net)
		_, platformAddressReceiveBidValue string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)

		multiSigScript        string = ""
		multiSigAddress       string = ""
		multiSigSegWitAddress string = ""
		inValue               uint64 = 0

		address string = "" //btc pair
		utxoId  string = ""
		amount  uint64 = 0
	)

	multiSigScript, multiSigAddress, multiSigSegWitAddress, err = createMultiSigAddress(netParams, []string{publicKey, platformPublicKeyMultiSig}...)
	if err != nil {
		return "", err
	}
	_ = multiSigScript
	_ = multiSigSegWitAddress
	_ = multiSigAddress
	//fmt.Printf("PublicKeyList:%+v\n", []string{publicKey, platformPublicKeyMultiSig})
	//fmt.Printf("multiSigScript:%+v\n", multiSigScript)
	//fmt.Printf("multiSigSegWitAddress:%+v\n", multiSigSegWitAddress)

	psbtBuilder, err = NewPsbtBuilder(netParams, req.CoinPsbtRaw)
	if err != nil {
		return "", err
	}
	switch req.PoolType {
	case model.PoolTypeTick:
		var (
			inscriptionBrc20BalanceItem *oklink_service.BalanceItem
			has                         = false
		)
		coinAddress = req.Address
		coinAmount = req.CoinAmount
		address = platformAddressReceiveBidValue

		preOutList := psbtBuilder.GetInputs()
		if preOutList == nil || len(preOutList) == 0 {
			return "", errors.New("Wrong Psbt: empty inputs. ")
		}
		for _, v := range preOutList {
			inscriptionId = fmt.Sprintf("%s:%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
			inscriptionBrc20BalanceItem, err = CheckBrc20Ordinals(v, req.Tick, coinAddress)
			if err != nil {
				continue
			}
			has = true

			preBrc20Tx, err := oklink_service.GetTxDetail(preOutList[0].PreviousOutPoint.Hash.String())
			if err != nil {
				return "", errors.New("Wrong Psbt: brc20 input is empty preTx. ")
			}
			inValueDe, err := decimal.NewFromString(preBrc20Tx.OutputDetails[preOutList[0].PreviousOutPoint.Index].Amount)
			if err != nil {
				return "", errors.New("Wrong Psbt: The value of brc20 input decimal parse err. ")
			}
			inValue = uint64(inValueDe.Mul(decimal.New(1, 8)).IntPart())
			if inValue == 0 {
				return "", errors.New("Wrong Psbt: brc20 out of preTx is empty amount. ")
			}
		}

		if req.Net == "mainnet" || req.Net == "livenet" {
			if !has || inscriptionBrc20BalanceItem == nil {
				return "", errors.New("Wrong Psbt: Empty inscription. ")
			}
			inscriptionNumber = inscriptionBrc20BalanceItem.InscriptionNumber
			coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
		}
		verified, err := CheckPublicKeyAddress(netParams, publicKey, coinAddress)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
		}
		if !verified {
			return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
		}

		outList := psbtBuilder.GetOutputs()
		if outList == nil || len(outList) == 0 {
			return "", errors.New("Wrong Psbt: empty outputs. ")
		}
		for _, v := range outList {
			addr, err := btcutil.DecodeAddress(multiSigSegWitAddress, netParams)
			if err != nil {
				return "", err
			}
			pkScript, err := txscript.PayToAddrScript(addr)
			if err != nil {
				return "", err
			}

			if hex.EncodeToString(v.PkScript) != hex.EncodeToString(pkScript) {
				return "", errors.New("Wrong Psbt: wrong multiSigScript of out in pool psbt. ")
			}
			outAmount = uint64(v.Value)
			//_, addrs, _, err := txscript.ExtractPkScriptAddrs(v.PkScript, netParams)
			//if err != nil {
			//	return "", err
			//}
			//fmt.Printf("multiSigSegWitAddress parse:%s\n", addrs[0].EncodeAddress())
		}

		outAmountDe := decimal.NewFromInt(int64(outAmount))
		coinAmountDe := decimal.NewFromInt(int64(coinAmount))
		coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
		coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)

		orderId = fmt.Sprintf("%s_%s_%s_%s_%d_%d", req.Net, req.Tick, inscriptionId, coinAddress, outAmount, coinAmount)
		orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
		break
	case model.PoolTypeBtc:
		return "", errors.New("Not yet. ")
	case model.PoolTypeBoth:
		var (
			inscriptionBrc20BalanceItem *oklink_service.BalanceItem
			has                         = false
		)
		coinAddress = req.Address
		coinAmount = req.CoinAmount
		address = platformAddressReceiveBidValue

		preOutList := psbtBuilder.GetInputs()
		if preOutList == nil || len(preOutList) == 0 {
			return "", errors.New("Wrong Psbt: empty inputs. ")
		}
		for _, v := range preOutList {
			inscriptionId = fmt.Sprintf("%s:%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
			inscriptionBrc20BalanceItem, err = CheckBrc20Ordinals(v, req.Tick, coinAddress)
			if err != nil {
				continue
			}
			has = true

			preBrc20Tx, err := oklink_service.GetTxDetail(preOutList[0].PreviousOutPoint.Hash.String())
			if err != nil {
				return "", errors.New("Wrong Psbt: brc20 input is empty preTx. ")
			}
			inValueDe, err := decimal.NewFromString(preBrc20Tx.OutputDetails[preOutList[0].PreviousOutPoint.Index].Amount)
			if err != nil {
				return "", errors.New("Wrong Psbt: The value of brc20 input decimal parse err. ")
			}
			inValue = uint64(inValueDe.Mul(decimal.New(1, 8)).IntPart())
			if inValue == 0 {
				return "", errors.New("Wrong Psbt: brc20 out of preTx is empty amount. ")
			}
		}

		if req.Net == "mainnet" || req.Net == "livenet" {
			if !has || inscriptionBrc20BalanceItem == nil {
				return "", errors.New("Wrong Psbt: Empty inscription. ")
			}
			inscriptionNumber = inscriptionBrc20BalanceItem.InscriptionNumber
			coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
		}
		verified, err := CheckPublicKeyAddress(netParams, publicKey, coinAddress)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
		}
		if !verified {
			return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
		}

		outList := psbtBuilder.GetOutputs()
		if outList == nil || len(outList) == 0 {
			return "", errors.New("Wrong Psbt: empty outputs. ")
		}
		for _, v := range outList {
			if hex.EncodeToString(v.PkScript) != multiSigScript {
				return "", errors.New("Wrong Psbt: wrong multiSigScript of out in pool psbt. ")
			}
			outAmount = uint64(v.Value)

		}

		outAmountDe := decimal.NewFromInt(int64(outAmount))
		coinAmountDe := decimal.NewFromInt(int64(coinAmount))
		coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
		coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)

		orderId = fmt.Sprintf("%s_%s_%s_%s_%d_%d_%s_%s_%d", req.Net, req.Tick, inscriptionId, coinAddress, outAmount, coinAmount, utxoId, address, amount)
		orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))

		return "", errors.New("Not yet. ")
	default:
		return "", errors.New("Wrong OrderState. ")
	}

	//todo Fix pool
	entity = &model.PoolBrc20Model{
		Net:            req.Net,
		OrderId:        orderId,
		Tick:           req.Tick,
		Pair:           fmt.Sprintf("%s_BTC", strings.ToUpper(req.Tick)),
		CoinAmount:     coinAmount,
		CoinDecimalNum: coinDec,
		CoinRatePrice:  coinRatePrice,
		CoinAddress:    coinAddress,
		CoinPublicKey:  publicKey,
		CoinInputValue: inValue,

		Amount:     outAmount,
		DecimalNum: amountDec,
		Address:    address,

		MultiSigScript:        multiSigScript,
		MultiSigScriptAddress: multiSigSegWitAddress,
		CoinPsbtRaw:           req.CoinPsbtRaw,
		PsbtRaw:               req.PsbtRaw,
		InscriptionId:         inscriptionId,
		InscriptionNumber:     inscriptionNumber,
		UtxoId:                utxoId,
		PoolType:              req.PoolType,
		PoolState:             req.PoolState,
		Timestamp:             tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetPoolBrc20Model(entity)
	if err != nil {
		return "", err
	}

	updatePoolInfo(entity)

	return "success", nil
}

func UpdatePoolOrder(req *request.OrderPoolBrc20UpdateReq, publicKey, ip string) (string, error) {
	var (
		netParams   *chaincfg.Params = GetNetParams(req.Net)
		entityOrder *model.PoolBrc20Model
	)
	entityOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(req.OrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("Order is empty. ")
	}

	if req.PoolState == model.PoolStateRemove {
		verified, err := CheckPublicKeyAddress(netParams, publicKey, entityOrder.CoinAddress)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
		}
		if !verified {
			return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
		}

		entityOrder.PoolState = req.PoolState
		_, err = mongo_service.SetPoolBrc20Model(entityOrder)
		if err != nil {
			return "", err
		}
	} else {
		return "", errors.New("Wrong state. ")
	}

	return req.OrderId, nil
}

func FetchPoolInscription(req *request.PoolBrc20FetchInscriptionReq, publicKey, ip string) (*respond.PoolInscriptionResp, error) {
	var (
		total      int64 = 0
		entityList []*model.PoolBrc20Model
		list       []*respond.PoolInscriptionItem = make([]*respond.PoolInscriptionItem, 0)
	)
	entityList, total = getMyPoolInscription(req.Net, req.Tick, req.Address)
	for _, v := range entityList {
		coinAmountStr := strconv.FormatUint(v.CoinAmount, 10)
		list = append(list, &respond.PoolInscriptionItem{
			InscriptionId:     v.InscriptionId,
			InscriptionNumber: v.InscriptionNumber,
			CoinAmount:        coinAmountStr,
		})
	}
	return &respond.PoolInscriptionResp{
		Net:   req.Net,
		Tick:  req.Tick,
		List:  list,
		Total: total,
	}, nil
}

func ClaimPool(req *request.PoolBrc20ClaimReq, publicKey, ip string) (*respond.PoolBrc20ClaimResp, error) {
	var (
		netParams                             *chaincfg.Params = GetNetParams(req.Net)
		entityOrder                           *model.PoolBrc20Model
		preSigScriptByte                      []byte
		err                                   error
		tx                                    *wire.MsgTx
		coinTx                                *wire.MsgTx
		coinTransferTx                        *wire.MsgTx
		psbtRaw                               string
		coinPsbtRaw                           string
		coinTransferPsbtRaw                   string
		_, platformAddressReceiveBidValue     string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		_, platformAddressMultiSigInscription string = GetPlatformKeyAndAddressForMultiSigInscription(req.Net)
		//_, platformAddressMultiSigInscriptionForReceiveValue string = GetPlatformKeyAndAddressForMultiSigInscriptionAndReceiveValue(req.Net)
		rewardPsbtRaw string = ""
	)

	entityOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(req.PoolOrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return nil, errors.New("Order is empty. ")
	}

	verified, err := CheckPublicKeyAddress(netParams, publicKey, entityOrder.CoinAddress)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
	}
	if !verified {
		return nil, errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
	}

	netParams = GetNetParams(entityOrder.Net)
	_ = netParams
	preSigScriptByte, err = hex.DecodeString(req.PreSigScript)
	if err != nil {
		return nil, err
	}

	//ordinals
	coinTx, coinPsbtRaw, err = claimPoolBrc20Order(req.PoolOrderId, platformAddressMultiSigInscription, 0, preSigScriptByte)
	if err != nil {
		return nil, err
	}
	_ = coinTx
	_ = psbtRaw

	//brc20
	coinTransferTx, coinTransferPsbtRaw, err = claimPoolBrc20Order(req.PoolOrderId, req.Address, model.PoolTypeMultiSigInscription, preSigScriptByte)
	if err != nil {
		return nil, err
	}
	_ = coinTransferTx
	_ = coinTransferPsbtRaw

	//btc
	tx, psbtRaw, err = claimPoolBrc20Order(req.PoolOrderId, platformAddressReceiveBidValue, model.PoolTypeBtc, preSigScriptByte)
	if err != nil {
		return nil, err
	}
	_ = tx

	//rewardPsbtRaw, err = makePoolRewardPsbt(entityOrder.Net, req.Address)
	//if err != nil {
	//	major.Println(fmt.Sprintf("[POOL-CLAIM] makePoolRewardPsbt err:%s\n", err.Error()))
	//}

	return &respond.PoolBrc20ClaimResp{
		Net:     entityOrder.Net,
		OrderId: entityOrder.OrderId,
		Tick:    entityOrder.Tick,
		//Fee:           0,
		CoinAmount:          entityOrder.CoinAmount,
		InscriptionId:       entityOrder.DealInscriptionId,
		CoinPsbtRaw:         coinPsbtRaw,
		PsbtRaw:             psbtRaw,
		CoinTransferPsbtRaw: coinTransferPsbtRaw,
		RewardPsbtRaw:       rewardPsbtRaw,
		RewardCoinAmount:    1500,
	}, nil
}

func UpdateClaim(req *request.PoolBrc20ClaimUpdateReq, publicKey, ip string) (string, error) {
	var (
		netParams                                                *chaincfg.Params
		entityOrder                                              *model.PoolBrc20Model
		err                                                      error
		txRaw                                                    string = ""
		finalClaimPsbtBuilder                                    *PsbtBuilder
		platformAddressReceiveBidValue                           string = ""
		platformAddressMultiSigInscription                       string = ""
		hasAddressMultiSigInscription, hasAddressReceiveBidValue bool   = false, false
		multiSigInscriptionTxIndex, multiSigInscriptionTxAmount  int64  = 0, 0
	)
	entityOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(req.PoolOrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("Order is empty. ")
	}

	_, platformAddressReceiveBidValue = GetPlatformKeyAndAddressReceiveBidValue(entityOrder.Net)
	_, platformAddressMultiSigInscription = GetPlatformKeyAndAddressForMultiSigInscription(entityOrder.Net)
	if req.RewardIndex == 1 {
		//finalClaimPsbtBuilder, err = addPoolRewardPsbt(entityOrder.Net, req.Address, req.PsbtRaw)
		//txRaw, err = finalClaimPsbtBuilder.ExtractPsbtTransaction()
		//if err != nil {
		//	return "", errors.New(fmt.Sprintf("PSBT: ExtractPsbtTransaction err:%s", err.Error()))
		//}
	} else {
		netParams = GetNetParams(entityOrder.Net)
		finalClaimPsbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Wrong PSBT: NewPsbtBuilder err:%s", err.Error()))
		}

		if finalClaimPsbtBuilder.GetOutputs() == nil || len(finalClaimPsbtBuilder.GetOutputs()) == 0 {
			return "", errors.New(fmt.Sprintf("Wrong PSBT: outputs are empty"))
		}
		if len(finalClaimPsbtBuilder.GetOutputs()) < 3 {
			return "", errors.New(fmt.Sprintf("Wrong PSBT: The length of the outputs is less than 3 "))
		}
		for k, v := range finalClaimPsbtBuilder.GetOutputs() {
			_, addrs, _, err := txscript.ExtractPkScriptAddrs(v.PkScript, netParams)
			if err != nil {
				return "", errors.New("Wrong Psbt: Extract address from out err. ")
			}
			if addrs[0].EncodeAddress() == platformAddressMultiSigInscription {
				multiSigInscriptionTxIndex = int64(k)
				hasAddressMultiSigInscription = true
				multiSigInscriptionTxAmount = v.Value
				if v.Value != 4000 {
					return "", errors.New(fmt.Sprintf("Wrong Psbt: value of output[%d] is not 4000", k))
				}
			} else if addrs[0].EncodeAddress() == platformAddressReceiveBidValue {
				hasAddressReceiveBidValue = true
			}
		}
		if !(hasAddressMultiSigInscription && hasAddressReceiveBidValue) {
			return "", errors.New(fmt.Sprintf("Wrong PSBT: outputs missing"))
		}

		txRaw, err = finalClaimPsbtBuilder.ExtractPsbtTransaction()
		if err != nil {
			return "", errors.New(fmt.Sprintf("PSBT: ExtractPsbtTransaction err:%s", err.Error()))
		}
	}

	err = updateClaim(entityOrder, txRaw)
	if err != nil {
		return "", err
	}
	saveNewMultiSigInscriptionUtxo(entityOrder.Net, entityOrder.ClaimTx, multiSigInscriptionTxIndex, uint64(multiSigInscriptionTxAmount))
	return "success", err
}
