package order_brc20_service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/node"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
	"time"
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

	if strings.Contains(req.Pair, "_") {
		req.Pair = strings.ReplaceAll(req.Pair, "_", "-")
	}

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
		multiSigScriptAddressTickAvailableState := v.MultiSigScriptAddressTickAvailableState
		decreasing := v.Decreasing
		if v.PoolState == model.PoolStateUsed {
			if v.MultiSigScriptAddressTickAvailableState == model.MultiSigScriptAddressTickAvailableStateNo {
				brc20TxResp, _ := oklink_service.GetAddressBrc20BalanceTransactionList(v.MultiSigScriptAddress, v.Tick, 0, 100)
				if brc20TxResp != nil && brc20TxResp.InscriptionsList != nil {
					for _, tx := range brc20TxResp.InscriptionsList {
						if tx.TxId == v.DealCoinTx && tx.State == "success" {
							multiSigScriptAddressTickAvailableState = model.MultiSigScriptAddressTickAvailableStateYes
							v.MultiSigScriptAddressTickAvailableState = multiSigScriptAddressTickAvailableState
							err := mongo_service.SetPoolBrc20ModelForMultiSigScriptAddressTickAvailableState(v)
							if err != nil {
								major.Println(fmt.Sprintf("SetPoolBrc20ModelForMultiSigScriptAddressTickAvailableState err: %s", err.Error()))
							}
							break
						}
					}
				}
			}
			decreasing = calculateDecrementFoNoReleasePool(v)
			time.Sleep(50 * time.Millisecond)
		}

		//rewardNowAmount := getRealNowReward(v)
		//bidCount := checkPoolBidCount(v)

		item := &respond.PoolBrc20Item{
			Net:                                     v.Net,
			OrderId:                                 v.OrderId,
			Tick:                                    v.Tick,
			Pair:                                    v.Pair,
			CoinAmount:                              v.CoinAmount,
			CoinDecimalNum:                          v.CoinDecimalNum,
			CoinRatePrice:                           v.CoinRatePrice,
			CoinPrice:                               v.CoinPrice,
			CoinPriceDecimalNum:                     v.CoinPriceDecimalNum,
			CoinAddress:                             v.CoinAddress,
			Amount:                                  v.Amount,
			DecimalNum:                              v.DecimalNum,
			Address:                                 v.Address,
			PoolType:                                v.PoolType,
			PoolState:                               v.PoolState,
			DealTx:                                  v.DealTx,
			DealCoinTx:                              v.DealCoinTx,
			MultiSigScriptAddress:                   v.MultiSigScriptAddress,
			DealInscriptionId:                       v.DealInscriptionId,
			DealInscriptionTx:                       v.DealInscriptionTx,
			DealInscriptionTxIndex:                  v.DealInscriptionTxIndex,
			DealInscriptionTxOutValue:               v.DealInscriptionTxOutValue,
			DealInscriptionTime:                     v.DealInscriptionTime,
			InscriptionId:                           v.InscriptionId,
			MultiSigScriptAddressTickAvailableState: multiSigScriptAddressTickAvailableState,
			UtxoId:                                  v.UtxoId,
			//PsbtRaw:       v.PsbtRaw,
			Timestamp: v.Timestamp,
			//RewardCoinAmount: rewardNowAmount,

			Percentage:           v.Percentage,
			RewardAmount:         v.RewardAmount,
			RewardRealAmount:     v.RewardRealAmount,
			PercentageExtra:      v.PercentageExtra,
			RewardExtraAmount:    v.RewardExtraAmount,
			DealCoinTxBlockState: v.DealCoinTxBlockState,
			DealCoinTxBlock:      v.DealCoinTxBlock,
			CalStartBlock:        v.CalStartBlock,
			CalEndBlock:          v.CalEndBlock,

			ReleaseTx:      v.ClaimTx,
			ReleaseTime:    v.ClaimTime,
			ReleaseTxBlock: v.ClaimTxBlock,
			DealTime:       v.DealTime,
			Decreasing:     decreasing,
			//BidCount:         bidCount,
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
		_, platformPublicKeyMultiSig                       = GetPlatformKeyMultiSig(req.Net)
		_, platformPublicKeyMultiSigBtc                    = GetPlatformKeyMultiSigForBtc(req.Net)
		_, platformAddressReceiveBidValueForPoolBtc string = GetPlatformKeyAndAddressReceiveValueForPoolBtc(req.Net)
	)
	return &respond.PoolKeyInfoResp{
		Net:               req.Net,
		PublicKey:         platformPublicKeyMultiSig,
		BtcPublicKey:      platformPublicKeyMultiSigBtc,
		BtcReceiveAddress: platformAddressReceiveBidValueForPoolBtc,
	}, nil
}

func PushPoolOrder(req *request.PoolBrc20PushReq, publicKey string) (string, error) {
	var (
		netParams         *chaincfg.Params = GetNetParams(req.Net)
		entity            *model.PoolBrc20Model
		err               error
		orderId           string = ""
		psbtBuilder       *PsbtBuilder
		btcPsbtBuilder    *PsbtBuilder
		coinAddress       string = ""
		coinAmount        uint64 = 0
		coinDec           int    = 18
		outAmount         uint64 = 0
		btcOutAmount      uint64 = 0
		amountDec         int    = 8
		coinRatePrice     uint64 = 0
		inscriptionId     string = ""
		inscriptionNumber string = ""

		_, platformPublicKeyMultiSig                       = GetPlatformKeyMultiSig(req.Net)
		_, platformAddressReceiveBidValue           string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		_, platformPublicKeyMultiSigBtc                    = GetPlatformKeyMultiSigForBtc(req.Net)
		_, platformAddressReceiveBidValueForPoolBtc string = GetPlatformKeyAndAddressReceiveValueForPoolBtc(req.Net)
		platformPrivateKeyLp, _                     string = GetPlatformKeyAndAddressForLp(req.Net)
		publicKeyLp                                 string = tool.GetPublicKeyFromPrivateKey(platformPrivateKeyLp)

		multiSigScript           string = ""
		multiSigAddress          string = ""
		multiSigSegWitAddress    string = ""
		multiSigScriptBtc        string = ""
		multiSigAddressBtc       string = ""
		multiSigSegWitAddressBtc string = ""
		inValue                  uint64 = 0

		address            string = "" //btc pair
		utxoId             string = req.BtcUtxoId
		amount             uint64 = 0
		btcPsbtPreOutTx    string = ""
		btcPsbtPreOutIndex uint32 = 0

		marketPrice         uint64 = 0
		marketCoinPrice     uint64 = 0
		coinPrice           int64  = 0
		coinPriceDecimalNum int32  = 0
	)
	_ = marketPrice

	if publicKey != publicKeyLp {
		if !CheckLpWhiteList(req.Address) {
			currentBlockHeight, err := node.CurrentBlockHeight(req.Net)
			if err != nil {
				return "", errors.New(fmt.Sprintf("CurrentBlockHeight err: %s. ", err.Error()))
			}
			if int64(currentBlockHeight) < config.EventOneStartBlock || int64(currentBlockHeight) > config.EventOneEndBlock {
				if req.Tick == "rdex" {
					return "", errors.New("rdex not in pool")
				}
			}
			//if tool.MakeTimestamp() < config.EventOneStartTime || tool.MakeTimestamp() > config.EventOneEndTime {
			//	if req.Tick == "rdex" {
			//		return "", errors.New("rdex not in pool")
			//	}
			//}
		}
	}

	if req.Ratio != 0 {
		if req.Ratio < 10 || req.Ratio > 10000 {
			return "", errors.New("ratio invalid")
		}
	}

	multiSigScript, multiSigAddress, multiSigSegWitAddress, err = CreateMultiSigAddress(netParams, []string{publicKey, platformPublicKeyMultiSig}...)
	if err != nil {
		return "", err
	}
	_ = multiSigScript
	_ = multiSigSegWitAddress
	_ = multiSigAddress

	multiSigScriptBtc, multiSigAddressBtc, multiSigSegWitAddressBtc, err = CreateMultiSigAddress(netParams, []string{publicKey, platformPublicKeyMultiSigBtc}...)
	if err != nil {
		return "", err
	}
	_ = multiSigScriptBtc
	_ = multiSigSegWitAddressBtc
	_ = multiSigAddressBtc
	//fmt.Printf("PublicKeyList:%+v\n", []string{publicKey, platformPublicKeyMultiSig})
	//fmt.Printf("multiSigScript:%+v\n", multiSigScript)
	//fmt.Printf("multiSigSegWitAddress:%+v\n", multiSigSegWitAddress)

	psbtBuilder, err = NewPsbtBuilder(netParams, req.CoinPsbtRaw)
	if err != nil {
		return "", err
	}
	if req.PsbtRaw != "" {
		btcPsbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
		if err != nil {
			return "", err
		}
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
		}

		outAmountDe := decimal.NewFromInt(int64(outAmount))
		coinAmountDe := decimal.NewFromInt(int64(coinAmount))
		coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
		coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)
		coinPrice, coinPriceDecimalNum = MakePrice(int64(coinAmount), int64(outAmount))

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
		address = req.Address

		preOutList := psbtBuilder.GetInputs()
		if preOutList == nil || len(preOutList) == 0 {
			return "", errors.New("Wrong Psbt: empty inputs. ")
		}
		for _, v := range preOutList {
			inscriptionId = fmt.Sprintf("%s:%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
			inscriptionBrc20BalanceItem, err = CheckBrc20Ordinals(v, req.Tick, coinAddress)
			if err != nil {
				fmt.Printf("CheckBrc20Ordinals err:%s\n", err.Error())
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
		}

		outAmountDe := decimal.NewFromInt(int64(outAmount))
		coinAmountDe := decimal.NewFromInt(int64(coinAmount))
		coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
		coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)
		coinPrice, coinPriceDecimalNum = MakePrice(int64(coinAmount), int64(outAmount))

		orderId = fmt.Sprintf("%s_%s_%s_%s_%d_%d_%s_%s_%d", req.Net, req.Tick, inscriptionId, coinAddress, outAmount, coinAmount, utxoId, address, amount)
		orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))

		if btcPsbtBuilder == nil {
			return "", errors.New("PsbtRaw is empty")
		}

		switch req.BtcPoolMode {
		case model.PoolModePsbt:
			btcOutList := btcPsbtBuilder.GetOutputs()
			if btcOutList == nil || len(btcOutList) == 0 {
				return "", errors.New("Wrong Psbt(btc): empty outputs. ")
			}
			for _, v := range btcOutList {
				addr, err := btcutil.DecodeAddress(multiSigSegWitAddressBtc, netParams)
				if err != nil {
					return "", err
				}
				pkScript, err := txscript.PayToAddrScript(addr)
				if err != nil {
					return "", err
				}
				if hex.EncodeToString(v.PkScript) != hex.EncodeToString(pkScript) {
					return "", errors.New("Wrong Psbt(btc): wrong multiSigScript of out in pool psbt. ")
				}
				btcOutAmount = uint64(v.Value)
			}
			//todo check btcOutAmount

			btcInList := btcPsbtBuilder.GetInputs()
			if btcInList == nil || len(btcInList) == 0 {
				return "", errors.New("Wrong Psbt(btc): empty inputs. ")
			}
			for _, v := range btcInList {
				btcPsbtPreOutTx = v.PreviousOutPoint.Hash.String()
				btcPsbtPreOutIndex = v.PreviousOutPoint.Index
			}
			utxoId = fmt.Sprintf("%s_%d", btcPsbtPreOutTx, btcPsbtPreOutIndex)

			break
		case model.PoolModeCustody, model.PoolModeNone:
			addr, err := btcutil.DecodeAddress(platformAddressReceiveBidValueForPoolBtc, netParams)
			if err != nil {
				return "", nil
			}
			pkScriptBtc, err := txscript.PayToAddrScript(addr)
			if err != nil {
				return "", nil
			}

			btcOutList := btcPsbtBuilder.GetOutputs()
			if btcOutList == nil || len(btcOutList) == 0 {
				return "", errors.New("Wrong Psbt(btc): empty outputs. ")
			}
			btcOutIndex := int64(0)
			for i, v := range btcOutList {
				if hex.EncodeToString(v.PkScript) == hex.EncodeToString(pkScriptBtc) {
					btcOutIndex = int64(i)
					btcOutAmount = uint64(v.Value)
				}
			}
			if btcOutAmount == 0 {
				return "", errors.New("Wrong Psbt(btc): empty value of out in pool psbt. ")
			}

			txRawPsbtBtc, err := btcPsbtBuilder.ExtractPsbtTransaction()
			if err != nil {
				return "", errors.New(fmt.Sprintf("Wrong Psbt(btc): ExtractPsbtTransaction err:%s", err.Error()))
			}
			txRawPsbtBtcByte, _ := hex.DecodeString(txRawPsbtBtc)

			txPsbtBtc := wire.NewMsgTx(2)
			err = txPsbtBtc.Deserialize(bytes.NewReader(txRawPsbtBtcByte))
			if err != nil {
				return "", errors.New(fmt.Sprintf("Wrong Psbt(btc): txRawPsbt Deserialize err:%s", err.Error()))
			}
			psbtBtcTxId := txPsbtBtc.TxHash().String()
			if utxoId != fmt.Sprintf("%s_%d", psbtBtcTxId, btcOutIndex) {
				return "", errors.New(fmt.Sprintf("Wrong Psbt(btc): wrong utxoId"))
			}

			_, err = unisat_service.BroadcastTx(req.Net, txRawPsbtBtc)
			if err != nil {
				return "", errors.New(fmt.Sprintf("Broadcast Psbt(btc) %s err:%s", req.Net, err.Error()))
			}
			break
		default:
			return "", errors.New("invalid BtcPoolMode")
		}
	default:
		return "", errors.New("Wrong OrderState. ")
	}
	marketPrice = GetMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	marketCoinPrice = GetMarketPriceV2(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

	if publicKey != publicKeyLp {
		//if coinRatePrice < marketPrice {
		//	fmt.Printf("coinRatePrice:%d, marketPrice:%d\n", coinRatePrice, marketPrice)
		//	return "", errors.New("The liquidity price must not be lower than the market price. ")
		//}
		if uint64(coinPrice) < marketCoinPrice {
			fmt.Printf("coinPrice:%d, marketCoinPrice:%d\n", coinPrice, marketCoinPrice)
			return "", errors.New("The liquidity price must not be lower than the market price. ")
		}
	}

	//check inscriptionId and utxoId which has been add
	inscriptionIdUsedCount, _ := mongo_service.FindUsedInscriptionPool(inscriptionId)
	if inscriptionIdUsedCount != 0 {
		return "", errors.New("Inscription has been Add/Used/Release in pool. ")
	}
	if btcPsbtPreOutTx != "" {
		btcUtxoIdUsedCount, _ := mongo_service.FindUsedBtcUtxoPool(inscriptionId)
		if btcUtxoIdUsedCount != 0 {
			fmt.Printf("Used InscriptionPool: [%s]\n", inscriptionId)
			return "", errors.New("Btc Utxo has been Add/Used/Release in pool. ")
		}
	}

	entity = &model.PoolBrc20Model{
		Net:                 req.Net,
		OrderId:             orderId,
		Tick:                req.Tick,
		Pair:                fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)),
		CoinAmount:          coinAmount,
		CoinDecimalNum:      coinDec,
		CoinRatePrice:       coinRatePrice,
		CoinPrice:           coinPrice,
		CoinPriceDecimalNum: coinPriceDecimalNum,
		CoinAddress:         coinAddress,
		CoinPublicKey:       publicKey,
		CoinInputValue:      inValue,

		Amount:     outAmount,
		DecimalNum: amountDec,
		Address:    address,

		MultiSigScript:           multiSigScript,
		MultiSigScriptAddress:    multiSigSegWitAddress,
		CoinPsbtRaw:              req.CoinPsbtRaw,
		MultiSigScriptBtc:        multiSigScriptBtc,
		MultiSigScriptAddressBtc: multiSigSegWitAddressBtc,
		PsbtRaw:                  req.PsbtRaw,
		InscriptionId:            inscriptionId,
		InscriptionNumber:        inscriptionNumber,
		BtcPoolMode:              req.BtcPoolMode,
		UtxoId:                   utxoId,
		PoolType:                 req.PoolType,
		PoolState:                req.PoolState,
		Ratio:                    req.Ratio,
		RewardRatio:              getRewardRatio(req.Ratio),
		Timestamp:                tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetPoolBrc20Model(entity)
	if err != nil {
		return "", err
	}

	if btcPsbtPreOutTx != "" {
		SaveForUserLpUtxo(entity.Net, entity.Tick, entity.Address, entity.OrderId, btcPsbtPreOutTx, int64(btcPsbtPreOutIndex), model.DummyStateLive)
	}

	updatePoolInfo(entity)

	return entity.OrderId, nil
}

func UpdatePoolOrder(req *request.OrderPoolBrc20UpdateReq, publicKey, ip string) (string, error) {
	var (
		netParams                                                                             *chaincfg.Params = GetNetParams(req.Net)
		entityOrder                                                                           *model.PoolBrc20Model
		limitTime                                                                             int64                   = 1000 * 60 * 60 * 24 * 15
		_, platformAddressReceiveBidValue                                                     string                  = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		platformPrivateKeyReceiveBidValueForPoolBtc, platformAddressReceiveBidValueForPoolBtc string                  = GetPlatformKeyAndAddressReceiveValueForPoolBtc(req.Net)
		refundUtxoList                                                                        []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		utxoListForRefundFee                                                                  []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		refundTx                                                                              string                  = ""
		txRaw                                                                                 string                  = ""
	)
	entityOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(req.OrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("Order is empty. ")
	}

	if req.PoolState == model.PoolStateRemove {
		if entityOrder.PoolState != model.PoolStateAdd {
			return "", errors.New("Order not in add state. ")
		}

		if publicKey != "" {
			verified, err := CheckPublicKeyAddress(netParams, publicKey, entityOrder.CoinAddress)
			if err != nil {
				return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
			}
			if !verified {
				return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
			}
		}
		//verified, err := CheckPublicKeyAddress(netParams, publicKey, entityOrder.CoinAddress)
		//if err != nil {
		//	return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
		//}
		//if !verified {
		//	return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
		//}

		// refund btc
		switch entityOrder.PoolType {
		case model.PoolTypeBoth, model.PoolTypeBtc:
			if entityOrder.BtcPoolMode == model.PoolModeCustody {
				addr, err := btcutil.DecodeAddress(platformAddressReceiveBidValueForPoolBtc, netParams)
				if err != nil {
					return "", nil
				}
				pkScriptBtc, err := txscript.PayToAddrScript(addr)
				if err != nil {
					return "", nil
				}
				if entityOrder.UtxoId == "" {
					return "", errors.New("utxoId is empty")
				}
				btcUtxoIdStrs := strings.Split(entityOrder.UtxoId, "_")
				if len(btcUtxoIdStrs) != 2 {
					return "", errors.New("utxoId format error")
				}
				btcTxId := btcUtxoIdStrs[0]
				btcTxOutIndex, _ := strconv.ParseInt(btcUtxoIdStrs[1], 10, 64)
				//liveUtxoList := make([]*oklink_service.UtxoItem, 0)
				//utxoResp, err := oklink_service.GetAddressUtxo(platformAddressReceiveBidValueForPoolBtc, 1, 50)
				//if err != nil {
				//	return "", errors.New(fmt.Sprintf("Recheck address utxo list err:%s", err.Error()))
				//}
				//if utxoResp.UtxoList != nil && len(utxoResp.UtxoList) != 0 {
				//	liveUtxoList = append(liveUtxoList, utxoResp.UtxoList...)
				//}

				//utxoList, err := unisat_service.GetAddressUtxo(platformAddressReceiveBidValueForPoolBtc)
				refundUtxoList = append(refundUtxoList, &model.OrderUtxoModel{
					TxId:          btcTxId,
					Index:         btcTxOutIndex,
					Amount:        entityOrder.Amount,
					PrivateKeyHex: platformPrivateKeyReceiveBidValueForPoolBtc,
					PkScript:      hex.EncodeToString(pkScriptBtc),
				})
				refundAmount := entityOrder.Amount

				utxoListForRefundFee, err = GetUnoccupiedUtxoList(req.Net, 1, 0, model.UtxoTypeBidY, "", 0)
				defer ReleaseUtxoList(utxoListForRefundFee)
				refundUtxoList = append(refundUtxoList, utxoListForRefundFee...)
				//todo
				if tool.MakeTimestamp()-entityOrder.CreateTime >= limitTime {

				} else {

				}

				tx, err := makeBtcRefundTx(netParams, refundUtxoList, refundAmount, entityOrder.Address, platformAddressReceiveBidValue, 14)
				if err != nil {
					fmt.Printf("BuildCommonTx err:%s\n", err.Error())
					return "", err
				}
				txRaw, err = ToRaw(tx)
				if err != nil {
					return "", err
				}
				txResp, err := unisat_service.BroadcastTx(req.Net, txRaw)
				if err != nil {
					return "", err
				}
				setUsedBidYUtxo(utxoListForRefundFee, txResp.Result)

				refundTx = txResp.Result
				entityOrder.RefundTx = refundTx
			}
			break
		}

		entityOrder.PoolState = req.PoolState
		removeInvalidBidByPoolOrderId(entityOrder.OrderId)

		_, err := mongo_service.SetPoolBrc20Model(entityOrder)
		if err != nil {
			return "", err
		}

		UpdateForOrderLiveUtxo(entityOrder.OrderId, model.DummyStateCancel)

		updatePoolInfo(entityOrder)
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
		rewardPsbtRaw     string = ""
		rewardNowAmount   int64  = 0
		btcReceiveAddress string = platformAddressReceiveBidValue
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
	if entityOrder.PoolType == model.PoolTypeBoth {
		btcReceiveAddress = entityOrder.Address
	}
	tx, psbtRaw, err = claimPoolBrc20Order(req.PoolOrderId, btcReceiveAddress, model.PoolTypeBtc, preSigScriptByte)
	if err != nil {
		return nil, err
	}
	_ = tx

	//rewardPsbtRaw, err = makePoolRewardPsbt(entityOrder.Net, req.Address)
	//if err != nil {
	//	major.Println(fmt.Sprintf("[POOL-CLAIM] makePoolRewardPsbt err:%s\n", err.Error()))
	//}

	//rewardNowAmount = getRealNowReward(entityOrder)

	decreasing := calculateDecrementFoNoReleasePool(entityOrder)
	rewardNowAmount = getRealNowRewardByDecreasing(entityOrder.RewardAmount, decreasing)

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
		RewardCoinAmount:    rewardNowAmount,
	}, nil
}

func UpdateClaim(req *request.PoolBrc20ClaimUpdateReq, publicKey, ip string) (string, error) {
	var (
		netParams             *chaincfg.Params
		entityOrder           *model.PoolBrc20Model
		err                   error
		txRaw                 string = ""
		finalClaimPsbtBuilder *PsbtBuilder
		//platformAddressReceiveBidValue                           string = ""
		platformAddressMultiSigInscription                                          string = ""
		hasAddressMultiSigInscription, hasAddressReceiveBtc, hasAddressReceiveBrc20 bool   = false, false, false
		multiSigInscriptionTxIndex, multiSigInscriptionTxAmount                     int64  = 0, 0

		rewardEntity *model.OrderBrc20Model
	)
	entityOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(req.PoolOrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("Order is empty. ")
	}

	//_, platformAddressReceiveBidValue = GetPlatformKeyAndAddressReceiveBidValue(entityOrder.Net)
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
				if v.Value != 4000 && v.Value != 5000 {
					return "", errors.New(fmt.Sprintf("Wrong Psbt: value of output[%d] is not 4000 or 5000", k))
				}
				//} else if addrs[0].EncodeAddress() == platformAddressReceiveBidValue {
			} else {
				if addrs[0].EncodeAddress() == entityOrder.Address {
					hasAddressReceiveBtc = true
				}
				if addrs[0].EncodeAddress() == entityOrder.CoinAddress {
					hasAddressReceiveBrc20 = true
				}
			}
		}
		if !(hasAddressMultiSigInscription && hasAddressReceiveBtc && hasAddressReceiveBrc20) {
			//fmt.Printf("[Release]hasAddressMultiSigInscription:%v, hasAddressReceiveBtc:%v, hasAddressReceiveBrc20:%v\n", hasAddressMultiSigInscription, hasAddressReceiveBtc, hasAddressReceiveBrc20)
			return "", errors.New(fmt.Sprintf("Wrong PSBT: outputs missing"))
		}

		if finalClaimPsbtBuilder.GetInputs() == nil || len(finalClaimPsbtBuilder.GetInputs()) == 0 {
			return "", errors.New(fmt.Sprintf("Wrong PSBT: inputs are empty"))
		}

		//get reward input
		for _, v := range finalClaimPsbtBuilder.GetInputs() {
			inscriptionId := fmt.Sprintf("%si%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
			rewardEntity, _ = mongo_service.FindOrderBrc20ModelByInscriptionId(inscriptionId, model.OrderStatePoolPreClaim)
			if rewardEntity != nil {
				break
			}
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

	if rewardEntity != nil {
		rewardEntity.OrderState = model.OrderStatePoolFinishClaim
		rewardEntity.PsbtRawFinalAsk = req.PsbtRaw
		rewardEntity.PsbtAskTxId = entityOrder.ClaimTx
		rewardEntity.DealTime = tool.MakeTimestamp()
		_, err = mongo_service.SetOrderBrc20Model(rewardEntity)
		if err != nil {
			//return "", err
		}
	}

	return "success", err
}

func FetchOwnerReward(req *request.PoolBrc20RewardReq) (*respond.PoolBrc20RewardResp, error) {
	var (
		entityBlockReward      *model.PoolRewardBlockUserCount
		entityReward           *model.PoolRewardCount
		entityRewardOrderCount *model.PoolRewardOrderCount
		totalRewardAmount      uint64 = 0
		totalRewardExtraAmount uint64 = 0
		//claimedOwnCoinAmount   uint64 = 0
		//claimedOwnAmount       uint64 = 0
		//claimedOwnCount        uint64 = 0
		hadClaimRewardAmount uint64 = 0
		//hadClaimRewardOrderCount uint64 = 0
		hasReleasePoolOrderCount int64            = 0
		rewardType               model.RewardType = model.RewardTypeNormal
		tick                     string           = ""
		rewardTick               string           = config.PlatformRewardTick
		startBlock               int64            = config.PlatformRewardCalStartBlock
		//entityExtraReward *model.PoolExtraRewardCount
	)

	//if req.Tick != config.PlatformRewardTick {
	//	return nil, errors.New(fmt.Sprintf("tick wrong:%s", config.PlatformRewardTick))
	//}

	_ = entityBlockReward
	//entityBlockReward, _ = mongo_service.CountPoolRewardBlockUser(req.Net, req.Address)
	//if entityBlockReward != nil {
	//	totalRewardAmount = uint64(entityBlockReward.RewardCoinAmountTotal)
	//	entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, "", "", req.Address)
	//	if entityRewardOrderCount != nil {
	//		hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)
	//	}
	//}
	//
	//_ = entityReward
	//
	hasReleasePoolOrderCount, _ = mongo_service.CountPoolBrc20ModelList(req.Net, req.Tick, "", req.Address, model.PoolTypeAll, model.PoolStateUsed)

	if req.Tick == "rdex" {
		tick = "rdex"
		rewardType = model.RewardTypeEventOneLp
		rewardTick = config.EventOneRewardTick
		startBlock = config.EventOneStartBlock
	}

	//entityExtraReward, _ = mongo_service.CountPoolExtraRewardRecord(req.Net, tick, "", "", req.Address, rewardType)
	//if entityExtraReward != nil {
	//	totalRewardExtraAmount = uint64(entityExtraReward.RewardExtraAmountTotal)
	//}

	entityReward, _ = mongo_service.CountOwnPoolReward(req.Net, tick, "", req.Address, startBlock)
	if entityReward != nil {
		totalRewardAmount = uint64(entityReward.RewardAmountTotal)
		//totalRewardExtraAmount = uint64(entityReward.RewardExtraAmountTotal)
		//claimedOwnCoinAmount = uint64(entityReward.CoinAmountTotal)
		//claimedOwnAmount = uint64(entityReward.AmountTotal)
		//claimedOwnCount = uint64(entityReward.OrderCounts)

		entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, rewardTick, "", req.Address, rewardType)
		fmt.Printf("[CLAIM-REWARD]entityRewardOrderCount:%v\n", entityRewardOrderCount)
		if entityRewardOrderCount != nil {
			hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)
			//hadClaimRewardOrderCount = uint64(entityRewardOrderCount.RewardCoinOrderCount)
		}
	}
	return &respond.PoolBrc20RewardResp{
		Net:                    req.Net,
		Tick:                   req.Tick,
		RewardTick:             rewardTick,
		TotalRewardAmount:      totalRewardAmount,
		TotalRewardExtraAmount: totalRewardExtraAmount,
		//ClaimedOwnCoinAmount:   claimedOwnCoinAmount,
		//ClaimedOwnAmount:       claimedOwnAmount,
		//ClaimedOwnCount:        claimedOwnCount,
		HadClaimRewardAmount: hadClaimRewardAmount,
		//HadClaimRewardOrderCount: hadClaimCoinOrderCount,
		HasReleasePoolOrderCount: hasReleasePoolOrderCount,
	}, nil
}

func ClaimReward(req *request.PoolBrc20ClaimRewardReq, publicKey, ip string) (string, error) {
	var (
		netParams              *chaincfg.Params = GetNetParams(req.Net)
		orderId                string           = ""
		entityOrder            *model.PoolRewardOrderModel
		nowTime                int64 = tool.MakeTimestamp()
		entityReward           *model.PoolRewardCount
		entityRewardOrderCount *model.PoolRewardOrderCount
		entityBlockReward      *model.PoolRewardBlockUserCount
		totalRewardAmount      uint64           = 0
		totalRewardExtraAmount uint64           = 0
		hadClaimRewardAmount   uint64           = 0
		remainingRewardAmount  int64            = 0
		rewardType             model.RewardType = model.RewardTypeNormal
		tick                   string           = ""
		rewardTick             string           = config.PlatformRewardTick
		startBlock             int64            = config.PlatformRewardCalStartBlock
		//entityExtraReward      *model.PoolExtraRewardCount
	)
	//if req.Tick != config.PlatformRewardTick {
	//	return "", errors.New(fmt.Sprintf("tick wrong:%s", config.PlatformRewardTick))
	//}

	verified, err := CheckPublicKeyAddress(netParams, publicKey, req.Address)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
	}
	if !verified {
		return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
	}

	_ = entityBlockReward
	//entityBlockReward, _ = mongo_service.CountPoolRewardBlockUser(req.Net, req.Address)
	//if entityBlockReward != nil {
	//	totalRewardAmount = uint64(entityBlockReward.RewardCoinAmountTotal)
	//	entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, "", "", req.Address)
	//	if entityRewardOrderCount != nil {
	//		hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)
	//		remainingRewardAmount = int64(totalRewardAmount) - int64(hadClaimRewardAmount)
	//	}
	//}

	if req.Tick == "rdex" {
		tick = "rdex"
		rewardType = model.RewardTypeEventOneLp
		rewardTick = config.EventOneRewardTick
		startBlock = config.EventOneStartBlock
	}

	//entityExtraReward, _ = mongo_service.CountPoolExtraRewardRecord(req.Net, tick, "", "", req.Address, rewardType)
	//if entityExtraReward != nil {
	//	totalRewardExtraAmount = uint64(entityExtraReward.RewardExtraAmountTotal)
	//}

	entityReward, _ = mongo_service.CountOwnPoolReward(req.Net, tick, "", req.Address, startBlock)
	if entityReward != nil {
		totalRewardAmount = uint64(entityReward.RewardAmountTotal)
		//totalRewardExtraAmount = uint64(entityReward.RewardExtraAmountTotal)

		entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, rewardTick, "", req.Address, rewardType)
		if entityRewardOrderCount != nil {
			hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)

			remainingRewardAmount = int64(totalRewardAmount+totalRewardExtraAmount) - int64(hadClaimRewardAmount)
		}
	}

	if remainingRewardAmount < 0 {
		remainingRewardAmount = 0
	}
	if remainingRewardAmount < req.RewardAmount || req.RewardAmount <= 0 {
		return "", errors.New(fmt.Sprintf("You only have %d rdex to claim.", remainingRewardAmount))
	}

	orderId = fmt.Sprintf("%s_%s_%s_%d_%d", req.Net, rewardTick, req.Address, req.RewardAmount, nowTime)
	orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
	entityOrder, _ = mongo_service.FindPoolRewardOrderModelByOrderId(orderId)
	if entityOrder != nil {
		return "", errors.New("already exist")
	}

	entityOrder = &model.PoolRewardOrderModel{
		Net:              req.Net,
		Tick:             rewardTick,
		OrderId:          orderId,
		Pair:             fmt.Sprintf("%s-BTC", strings.ToUpper(rewardTick)),
		RewardCoinAmount: req.RewardAmount,
		Address:          req.Address,
		RewardState:      model.RewardStateCreate,
		Timestamp:        nowTime,
		RewardType:       rewardType,
	}

	_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
	if err != nil {
		return "", errors.New("create order err")
	}

	return "success", nil
}

func FetchPoolRewardOrders(req *request.PoolRewardOrderFetchReq) (*respond.PoolRewardOrderResponse, error) {
	var (
		entityList []*model.PoolRewardOrderModel
		total      int64                          = 0
		list       []*respond.PoolRewardOrderItem = make([]*respond.PoolRewardOrderItem, 0)
		flag       int64                          = 0

		rewardType model.RewardType = model.RewardTypeNormal
		rewardTick string           = config.PlatformRewardTick
	)
	//if req.Tick != config.PlatformRewardTick {
	//	return nil, errors.New(fmt.Sprintf("tick wrong:%s", config.PlatformRewardTick))
	//}
	if req.RewardType != 0 {
		rewardType = req.RewardType
		if req.RewardType == model.RewardTypeEventOneLp || req.RewardType == model.RewardTypeEventOneBid {
			rewardTick = config.EventOneRewardTick
		}
	}

	total, _ = mongo_service.CountPoolRewardOrderModelList(req.Net, rewardTick, req.Pair, req.Address, req.RewardState, rewardType)
	entityList, _ = mongo_service.FindPoolRewardOrderModelList(req.Net, rewardTick, req.Pair, req.Address, req.RewardState,
		req.Limit, req.Flag, req.Page, req.SortKey, req.SortType, rewardType)
	for _, v := range entityList {
		item := &respond.PoolRewardOrderItem{
			Net:              v.Net,
			Tick:             v.Tick,
			OrderId:          v.OrderId,
			Pair:             v.Pair,
			RewardCoinAmount: v.RewardCoinAmount,
			Address:          v.Address,
			RewardState:      v.RewardState,
			Timestamp:        v.Timestamp,
			InscriptionId:    v.InscriptionId,
			SendId:           v.SendId,
		}
		if req.SortKey == "todo" {
			//flag = int64(v.CoinRatePrice)
		} else {
			flag = v.Timestamp
		}
		list = append(list, item)
	}
	return &respond.PoolRewardOrderResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}

func RefundPool() (string, error) {
	var (
		net                                       string           = "livenet"
		netParams                                 *chaincfg.Params = GetNetParams(net)
		orderId                                   string           = "e1b3e1d6c64478eb5cc98dd1504266051218ad55229a6e954699c50619962686"
		poolOrder                                 *model.PoolBrc20Model
		entity                                    *model.OrderBrc20Model
		platformPrivateKeyToX, platformAddressToX                         = GetPlatformKeyAndAddressReceiveBidValueToX(net)
		refundUtxoList                            []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		txRaw                                     string                  = ""
		refundAddress                             string                  = "bc1pkyn26w9sc654a2acc473twtfpa6yc8c0tu89t3gc273htk364j3qhy2xve"
	)
	entity, _ = mongo_service.FindOrderBrc20ModelByOrderId(orderId)
	if entity == nil {
		return "", errors.New("order is empty")
	}
	poolOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(entity.PoolOrderId)
	if poolOrder == nil {
		return "", errors.New("pool order is empty")
	}
	if entity.BidValueToXUtxoId == "" {
		return "", errors.New("BidValueToXUtxoId is empty")
	}

	addr, err := btcutil.DecodeAddress(platformAddressToX, netParams)
	if err != nil {
		return "", nil
	}
	pkScriptBtc, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", nil
	}

	btcUtxoIdStrs := strings.Split(entity.BidValueToXUtxoId, "_")
	if len(btcUtxoIdStrs) != 2 {
		return "", errors.New("utxoId format error")
	}
	btcTxId := btcUtxoIdStrs[0]
	btcTxOutIndex, _ := strconv.ParseInt(btcUtxoIdStrs[1], 10, 64)
	refundUtxoList = append(refundUtxoList, &model.OrderUtxoModel{
		TxId:          btcTxId,
		Index:         btcTxOutIndex,
		Amount:        67307500,
		PrivateKeyHex: platformPrivateKeyToX,
		PkScript:      hex.EncodeToString(pkScriptBtc),
	})
	refundAmount := uint64(67307500)

	utxoListForRefundFee, err := GetUnoccupiedUtxoList(net, 1, 20000, model.UtxoTypeBidY, "", 0)
	defer ReleaseUtxoList(utxoListForRefundFee)
	refundUtxoList = append(refundUtxoList, utxoListForRefundFee...)

	tx, err := makeBtcRefundTx(netParams, refundUtxoList, refundAmount, refundAddress, refundAddress, 45)
	if err != nil {
		fmt.Printf("BuildCommonTx err:%s\n", err.Error())
		return "", err
	}
	txRaw, err = ToRaw(tx)
	if err != nil {
		return "", err
	}
	txResp, err := unisat_service.BroadcastTx(net, txRaw)
	if err != nil {
		return "", err
	}

	setUsedBidYUtxo(utxoListForRefundFee, txResp.Result)

	return txResp.Result, nil
}

func FetchRewardRecord(req *request.PoolRewardRecordFetchReq) (*respond.PoolRewardRecordResponse, error) {
	var (
		entityList []*model.RewardRecordModel
		total      int64                           = 0
		list       []*respond.PoolRewardRecordItem = make([]*respond.PoolRewardRecordItem, 0)
		flag       int64                           = 0

		rewardType model.RewardType = model.RewardTypeNormal
		rewardTick string           = config.PlatformRewardTick
	)
	_ = rewardTick
	if req.Tick == "" {
		req.Tick = config.PlatformRewardTick
	}
	if req.RewardType == model.RewardTypeEventOneBid || req.RewardType == model.RewardTypeEventOneLpUnusedV2 {
		rewardType = req.RewardType
		rewardTick = config.EventOneRewardTick
	} else if req.RewardType == model.RewardTypeExtra {
		rewardType = req.RewardType
	}

	total, _ = mongo_service.CountRewardRecordModelList(req.Net, req.Tick, req.Address, rewardType)
	entityList, _ = mongo_service.FindRewardRecordModelList(req.Net, req.Tick, req.Address,
		req.Limit, req.Flag, req.Page, req.SortKey, req.SortType, rewardType)
	for _, v := range entityList {
		fromOrderTick := ""
		fromOrderCoinAmount := uint64(0)
		fromOrderAmount := uint64(0)
		fromOrderReward := int64(0)
		fromOrderPercentage := int64(0)
		fromOrderDealBlock := int64(0)
		fromOrderDealTime := int64(0)
		if (v.RewardType == model.RewardTypeExtra || v.RewardType == model.RewardTypeEventOneLpUnusedV2) && v.FromOrderId != "" {
			fromOrderTick, fromOrderCoinAmount, fromOrderAmount = mongo_service.FindPoolBrc20ModelTickAndAmountByOrderId(v.FromOrderId)
		} else if v.RewardType == model.RewardTypeEventOneBid && v.FromOrderId != "" {
			fromOrderTick, fromOrderCoinAmount, fromOrderAmount, fromOrderReward, fromOrderPercentage, fromOrderDealBlock, fromOrderDealTime = mongo_service.FindOrderBrc20ModelTickAndAmountByOrderId(v.FromOrderId)
		}

		item := &respond.PoolRewardRecordItem{
			Net:                 v.Net,
			Tick:                v.Tick,
			OrderId:             v.OrderId,
			FromOrderId:         v.FromOrderId,
			FromOrderTick:       fromOrderTick,
			FromOrderCoinAmount: int64(fromOrderCoinAmount),
			FromOrderAmount:     int64(fromOrderAmount),
			FromOrderReward:     fromOrderReward,
			FromOrderPercentage: fromOrderPercentage,
			FromOrderDealBlock:  fromOrderDealBlock,
			FromOrderDealTime:   fromOrderDealTime,
			Address:             v.Address,
			Percentage:          v.Percentage,
			RewardAmount:        v.RewardAmount,
			RewardType:          v.RewardType,
			CalBigBlock:         v.CalBigBlock,
			CalStartBlock:       v.CalStartBlock,
			CalEndBlock:         v.CalEndBlock,
		}
		if req.SortKey == "todo" {
			//flag = int64(v.CoinRatePrice)
		} else {
			flag = v.Timestamp
		}
		list = append(list, item)
	}
	return &respond.PoolRewardRecordResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}

func FetchErrPoolOrders(req *request.PoolBrc20ErrFetchReq) (*respond.PoolResponse, error) {
	var (
		entityList []*model.PoolBrc20Model
		total      int64                    = 0
		list       []*respond.PoolBrc20Item = make([]*respond.PoolBrc20Item, 0)
		flag       int64                    = 0
	)
	if req.Address == "" {
		return nil, errors.New("address is empty")
	}
	total, _ = mongo_service.CountPoolBrc20ModelErrList(req.Net, req.Tick, req.Pair, req.Address)
	entityList, _ = mongo_service.FindPoolBrc20ModelErrList(req.Net, req.Tick, req.Pair, req.Address,
		req.Limit, req.Flag, req.Page, req.SortKey, req.SortType)
	for _, v := range entityList {
		multiSigScriptAddressTickAvailableState := v.MultiSigScriptAddressTickAvailableState
		decreasing := v.Decreasing
		if v.PoolCoinState == model.PoolStateUsed {
			if v.MultiSigScriptAddressTickAvailableState == model.MultiSigScriptAddressTickAvailableStateNo {
				brc20TxResp, _ := oklink_service.GetAddressBrc20BalanceTransactionList(v.MultiSigScriptAddress, v.Tick, 0, 100)
				if brc20TxResp != nil && brc20TxResp.InscriptionsList != nil {
					for _, tx := range brc20TxResp.InscriptionsList {
						if tx.TxId == v.DealCoinTx && tx.State == "success" {
							multiSigScriptAddressTickAvailableState = model.MultiSigScriptAddressTickAvailableStateYes
							v.MultiSigScriptAddressTickAvailableState = multiSigScriptAddressTickAvailableState
							err := mongo_service.SetPoolBrc20ModelForMultiSigScriptAddressTickAvailableState(v)
							if err != nil {
								major.Println(fmt.Sprintf("SetPoolBrc20ModelForMultiSigScriptAddressTickAvailableState err: %s", err.Error()))
							}
							break
						}
					}
				}
			}
			decreasing = calculateDecrementFoNoReleasePool(v)
			time.Sleep(50 * time.Millisecond)
		}

		//rewardNowAmount := getRealNowReward(v)
		//bidCount := checkPoolBidCount(v)

		item := &respond.PoolBrc20Item{
			Net:                                     v.Net,
			OrderId:                                 v.OrderId,
			Tick:                                    v.Tick,
			Pair:                                    v.Pair,
			CoinAmount:                              v.CoinAmount,
			CoinDecimalNum:                          v.CoinDecimalNum,
			CoinRatePrice:                           v.CoinRatePrice,
			CoinPrice:                               v.CoinPrice,
			CoinPriceDecimalNum:                     v.CoinPriceDecimalNum,
			CoinAddress:                             v.CoinAddress,
			Amount:                                  v.Amount,
			DecimalNum:                              v.DecimalNum,
			Address:                                 v.Address,
			PoolType:                                v.PoolType,
			PoolState:                               v.PoolState,
			DealTx:                                  v.DealTx,
			DealCoinTx:                              v.DealCoinTx,
			MultiSigScriptAddress:                   v.MultiSigScriptAddress,
			DealInscriptionId:                       v.DealInscriptionId,
			DealInscriptionTx:                       v.DealInscriptionTx,
			DealInscriptionTxIndex:                  v.DealInscriptionTxIndex,
			DealInscriptionTxOutValue:               v.DealInscriptionTxOutValue,
			DealInscriptionTime:                     v.DealInscriptionTime,
			InscriptionId:                           v.InscriptionId,
			MultiSigScriptAddressTickAvailableState: multiSigScriptAddressTickAvailableState,
			UtxoId:                                  v.UtxoId,
			//PsbtRaw:       v.PsbtRaw,
			Timestamp: v.Timestamp,
			//RewardCoinAmount: rewardNowAmount,

			Percentage:           v.Percentage,
			RewardAmount:         v.RewardAmount,
			RewardRealAmount:     v.RewardRealAmount,
			PercentageExtra:      v.PercentageExtra,
			RewardExtraAmount:    v.RewardExtraAmount,
			DealCoinTxBlockState: v.DealCoinTxBlockState,
			DealCoinTxBlock:      v.DealCoinTxBlock,
			CalStartBlock:        v.CalStartBlock,
			CalEndBlock:          v.CalEndBlock,

			ReleaseTx:      v.ClaimTx,
			ReleaseTime:    v.ClaimTime,
			ReleaseTxBlock: v.ClaimTxBlock,
			DealTime:       v.DealTime,
			Decreasing:     decreasing,
			//BidCount:         bidCount,
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

func ReleaseErrPool(req *request.PoolBrc20ClaimReq, publicKey, ip string) (*respond.PoolBrc20ClaimResp, error) {
	var (
		netParams        *chaincfg.Params = GetNetParams(req.Net)
		entityOrder      *model.PoolBrc20Model
		preSigScriptByte []byte
		err              error
		//tx                                    *wire.MsgTx
		coinTx              *wire.MsgTx
		coinTransferTx      *wire.MsgTx
		psbtRaw             string
		coinPsbtRaw         string
		coinTransferPsbtRaw string
		//_, platformAddressReceiveBidValue     string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		_, platformAddressMultiSigInscription string = GetPlatformKeyAndAddressForMultiSigInscription(req.Net)
		//_, platformAddressMultiSigInscriptionForReceiveValue string = GetPlatformKeyAndAddressForMultiSigInscriptionAndReceiveValue(req.Net)
		rewardPsbtRaw   string = ""
		rewardNowAmount int64  = 0
	)

	entityOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(req.PoolOrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return nil, errors.New("Order is empty. ")
	}

	if !(entityOrder.PoolCoinState == model.PoolStateUsed && entityOrder.PoolState != model.PoolStateUsed) {
		return nil, errors.New("Pool state is not err. ")
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

	////btc
	//if entityOrder.PoolType == model.PoolTypeBoth {
	//	btcReceiveAddress = entityOrder.Address
	//}
	//tx, psbtRaw, err = claimPoolBrc20Order(req.PoolOrderId, btcReceiveAddress, model.PoolTypeBtc, preSigScriptByte)
	//if err != nil {
	//	return nil, err
	//}
	//_ = tx

	//decreasing := calculateDecrementFoNoReleasePool(entityOrder)
	//rewardNowAmount = getRealNowRewardByDecreasing(entityOrder.RewardAmount, decreasing)

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
		RewardCoinAmount:    rewardNowAmount,
	}, nil
}

func UpdateErrRelease(req *request.PoolBrc20ClaimUpdateReq, publicKey, ip string) (string, error) {
	var (
		netParams             *chaincfg.Params
		entityOrder           *model.PoolBrc20Model
		err                   error
		txRaw                 string = ""
		finalClaimPsbtBuilder *PsbtBuilder
		//platformAddressReceiveBidValue                           string = ""
		platformAddressMultiSigInscription                       string = ""
		hasAddressMultiSigInscription, _, hasAddressReceiveBrc20 bool   = false, false, false
		multiSigInscriptionTxIndex, multiSigInscriptionTxAmount  int64  = 0, 0
	)
	entityOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(req.PoolOrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("Order is empty. ")
	}

	//_, platformAddressReceiveBidValue = GetPlatformKeyAndAddressReceiveBidValue(entityOrder.Net)
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
		if len(finalClaimPsbtBuilder.GetOutputs()) < 2 {
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
				if v.Value != 4000 && v.Value != 5000 {
					return "", errors.New(fmt.Sprintf("Wrong Psbt: value of output[%d] is not 4000 or 5000", k))
				}
				//} else if addrs[0].EncodeAddress() == platformAddressReceiveBidValue {
			} else {

				if addrs[0].EncodeAddress() == entityOrder.CoinAddress {
					hasAddressReceiveBrc20 = true
				}
			}
		}
		if !(hasAddressMultiSigInscription && hasAddressReceiveBrc20) {
			//fmt.Printf("[Release]hasAddressMultiSigInscription:%v, hasAddressReceiveBtc:%v, hasAddressReceiveBrc20:%v\n", hasAddressMultiSigInscription, hasAddressReceiveBtc, hasAddressReceiveBrc20)
			return "", errors.New(fmt.Sprintf("Wrong PSBT: outputs missing"))
		}

		if finalClaimPsbtBuilder.GetInputs() == nil || len(finalClaimPsbtBuilder.GetInputs()) == 0 {
			return "", errors.New(fmt.Sprintf("Wrong PSBT: inputs are empty"))
		}

		txRaw, err = finalClaimPsbtBuilder.ExtractPsbtTransaction()
		if err != nil {
			return "", errors.New(fmt.Sprintf("PSBT: ExtractPsbtTransaction err:%s", err.Error()))
		}
	}

	err = updateErrClaim(entityOrder, txRaw)
	if err != nil {
		return "", err
	}
	saveNewMultiSigInscriptionUtxo(entityOrder.Net, entityOrder.ClaimTx, multiSigInscriptionTxIndex, uint64(multiSigInscriptionTxAmount))

	return "success", err
}

func FetchStandbyOwnerReward(req *request.PoolBrc20RewardReq) (*respond.PoolBrc20RewardResp, error) {
	var (
		//entityRewardSeller       *model.EventRewardCount
		//entityRewardBuyer        *model.EventRewardCount
		entityRewardOrderCount   *model.PoolRewardOrderCount
		totalRewardAmount        uint64 = 0
		totalRewardExtraAmount   uint64 = 0
		hadClaimRewardAmount     uint64 = 0
		hasReleasePoolOrderCount int64  = 0
		tick                     string = "rdex"
		rewardTick               string = config.EventOneRewardTick
		entityExtraReward        *model.RewardCount

		extraRewardType model.RewardType = model.RewardTypeEventOneLpUnusedV2
		eventRewardType model.RewardType = model.RewardTypeEventOneBid
	)

	if req.RewardType == model.RewardTypeExtra {
		tick = req.Tick
		rewardTick = config.PlatformRewardTick
		extraRewardType = model.RewardTypeExtra
		eventRewardType = model.RewardTypeExtra
	} else {
		return nil, errors.New(fmt.Sprintf("rewardType wrong:%d", req.RewardType))
	}

	entityExtraReward, _ = mongo_service.CountRewardRecord(req.Net, tick, "", "", req.Address, extraRewardType)
	if entityExtraReward != nil {
		totalRewardExtraAmount = uint64(entityExtraReward.RewardAmountTotal)
	}

	entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, rewardTick, "", req.Address, eventRewardType)
	if entityRewardOrderCount != nil {
		hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)
	}

	if hadClaimRewardAmount > totalRewardAmount+totalRewardExtraAmount {
		hadClaimRewardAmount = totalRewardAmount + totalRewardExtraAmount
	}

	return &respond.PoolBrc20RewardResp{
		Net:                      req.Net,
		Tick:                     req.Tick,
		RewardTick:               rewardTick,
		TotalRewardAmount:        totalRewardAmount,
		TotalRewardExtraAmount:   totalRewardExtraAmount,
		HadClaimRewardAmount:     hadClaimRewardAmount,
		HasReleasePoolOrderCount: hasReleasePoolOrderCount,
	}, nil
}

func CalStandbyClaimFee(req *request.OrderBrc20CalFeeReq) (*respond.CalFeeResp, error) {
	var (
		feeAmountForRewardInscription         int64  = 4000
		feeAmountForRewardSend                int64  = 4000
		_, platformAddressRewardBrc20FeeUtxos string = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(req.Net)
	)
	if req.Version == 2 {
		_, feeAmountForRewardInscription, feeAmountForRewardSend = GenerateBidTakerFee(req.NetworkFeeRate)
	}
	return &respond.CalFeeResp{
		RewardInscriptionFee: feeAmountForRewardInscription,
		RewardSendFee:        feeAmountForRewardSend,
		FeeAddress:           platformAddressRewardBrc20FeeUtxos,
	}, nil
}

func ClaimStandbyReward(req *request.PoolBrc20ClaimRewardReq, publicKey, ip string) (string, error) {
	var (
		netParams                                                                 *chaincfg.Params = GetNetParams(req.Net)
		orderId                                                                   string           = ""
		entityOrder                                                               *model.PoolRewardOrderModel
		nowTime                                                                   int64 = tool.MakeTimestamp()
		entityRewardOrderCount                                                    *model.PoolRewardOrderCount
		entityBlockReward                                                         *model.PoolRewardBlockUserCount
		totalRewardAmount                                                         uint64           = 0
		totalRewardExtraAmount                                                    uint64           = 0
		hadClaimRewardAmount                                                      uint64           = 0
		remainingRewardAmount                                                     int64            = 0
		tick                                                                      string           = "rdex"
		rewardTick                                                                string           = config.EventOneRewardTick
		rewardType                                                                model.RewardType = model.RewardTypeEventOneBid
		revealOutValue                                                            int64            = 546
		_, platformAddressRewardBrc20                                             string           = config.EventPlatformPrivateKeyRewardBrc20, config.EventPlatformAddressRewardBrc20
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string           = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(req.Net)
		feeAmountForRewardInscription                                             int64            = 4000
		feeAmountForRewardSend                                                    int64            = 4000
		entityExtraReward                                                         *model.RewardCount

		extraRewardType model.RewardType = model.RewardTypeEventOneLpUnusedV2
		eventRewardType model.RewardType = model.RewardTypeEventOneBid
	)

	if req.RewardType == model.RewardTypeExtra {
		tick = req.Tick
		rewardType = model.RewardTypeExtra
		rewardTick = config.PlatformRewardTick
		extraRewardType = model.RewardTypeExtra
		eventRewardType = model.RewardTypeExtra
		_, platformAddressRewardBrc20 = GetPlatformRewardPrivateKeyAndAddress(req.Net, req.RewardType)
	} else {
		return "", errors.New(fmt.Sprintf("rewardType wrong:%d", req.RewardType))
	}

	verified, err := CheckPublicKeyAddress(netParams, publicKey, req.Address)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
	}
	if !verified {
		return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
	}

	_ = entityBlockReward

	entityExtraReward, _ = mongo_service.CountRewardRecord(req.Net, tick, "", "", req.Address, extraRewardType)
	if entityExtraReward != nil {
		totalRewardExtraAmount = uint64(entityExtraReward.RewardAmountTotal)
	}

	entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, rewardTick, "", req.Address, eventRewardType)
	if entityRewardOrderCount != nil {
		hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)

		remainingRewardAmount = int64(totalRewardAmount+totalRewardExtraAmount) - int64(hadClaimRewardAmount)
	}

	if remainingRewardAmount < 0 {
		remainingRewardAmount = 0
	}
	if remainingRewardAmount < req.RewardAmount || req.RewardAmount <= 0 {
		return "", errors.New(fmt.Sprintf("You only have %d rdex to claim.", remainingRewardAmount))
	}

	brc20BalanceResult, err := oklink_service.GetAddressBrc20BalanceResult(platformAddressRewardBrc20, rewardTick, 1, 50)
	if err != nil {
		return "", err
	}
	availableBalance, _ := strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	fmt.Printf("availableBalance:%d, req.InscribeTransferAmount*req.Count: %d\n", availableBalance, req.RewardAmount*1)
	if availableBalance < req.RewardAmount*1 {
		return "", errors.New("AvailableBalance not enough. ")
	}

	//if CheckEventRemainingRewardTotal() {
	//	return "", errors.New("event remaining reward have been capped. ")
	//}

	orderId = fmt.Sprintf("%s_%s_%s_%d_%d_%d", req.Net, rewardTick, req.Address, req.RewardAmount, nowTime, model.RewardTypeEventOneBid)
	orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
	entityOrder, _ = mongo_service.FindPoolRewardOrderModelByOrderId(orderId)
	if entityOrder != nil {
		return "", errors.New("already exist")
	}

	entityOrder = &model.PoolRewardOrderModel{
		Net:              req.Net,
		Tick:             rewardTick,
		OrderId:          orderId,
		Pair:             fmt.Sprintf("%s-BTC", strings.ToUpper(rewardTick)),
		RewardCoinAmount: req.RewardAmount,
		Address:          req.Address,
		RewardState:      model.RewardStateCreate,
		Timestamp:        nowTime,
		RewardType:       rewardType,
		FeeRawTx:         req.FeeRawTx,
		FeeUtxoTxId:      req.FeeUtxoTxId,
		FeeInscription:   req.FeeInscription,
		FeeSend:          req.FeeSend,
		NetworkFeeRate:   req.NetworkFeeRate,
		Version:          req.Version,
	}

	txRawByte, _ := hex.DecodeString(req.FeeRawTx)
	tx := wire.NewMsgTx(2)
	err = tx.Deserialize(bytes.NewReader(txRawByte))
	if err != nil {
		fmt.Printf(fmt.Sprintf("[REWARD-INSCRIPTION]  feeRawTx Deserialize err:%s", err.Error()))
		return "", err
	}
	entityOrder.FeeUtxoTxId = tx.TxHash().String()

	_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
	if err != nil {
		return "", errors.New("create order err")
	}

	addr, err := btcutil.DecodeAddress(platformAddressRewardBrc20FeeUtxos, GetNetParams(entityOrder.Net))
	if err != nil {
		return "", err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}

	if len(tx.TxOut) < 2 {
		return "", errors.New("feeRawTx wrong")
	}
	if tx.TxOut[0].Value != entityOrder.FeeInscription || tx.TxOut[1].Value != entityOrder.FeeSend ||
		hex.EncodeToString(tx.TxOut[0].PkScript) != hex.EncodeToString(pkScriptByte) ||
		hex.EncodeToString(tx.TxOut[1].PkScript) != hex.EncodeToString(pkScriptByte) ||
		tx.TxHash().String() != entityOrder.FeeUtxoTxId {
		return "", errors.New("feeRawTx wrong")
	}

	if entityOrder.Version == 2 {
		_, feeAmountForRewardInscription, feeAmountForRewardSend = GenerateBidTakerFee(req.NetworkFeeRate)
		if entityOrder.FeeInscription != feeAmountForRewardInscription || entityOrder.FeeSend != feeAmountForRewardSend {
			return "", errors.New("feeAmount wrong")
		}

		txRaw, err := ToRaw(tx)
		if err != nil {
			return "", err
		}
		_, err = unisat_service.BroadcastTx(req.Net, txRaw)
		if err != nil {
			return "", err
		}

		utxoRewardInscriptionList := make([]*model.OrderUtxoModel, 0)
		utxoRewardSendList := make([]*model.OrderUtxoModel, 0)

		utxoRewardInscriptionList = append(utxoRewardInscriptionList, &model.OrderUtxoModel{
			UtxoId:        fmt.Sprintf("%s_%d", req.FeeUtxoTxId, 0),
			Net:           entityOrder.Net,
			UtxoType:      model.UtxoTypeRewardInscription,
			Amount:        uint64(entityOrder.FeeInscription),
			Address:       platformAddressRewardBrc20FeeUtxos,
			PrivateKeyHex: platformPrivateKeyRewardBrc20FeeUtxos,
			TxId:          entityOrder.FeeUtxoTxId,
			Index:         0,
			PkScript:      hex.EncodeToString(pkScriptByte),
		})
		utxoRewardSendList = append(utxoRewardSendList, &model.OrderUtxoModel{
			UtxoId:        fmt.Sprintf("%s_%d", req.FeeUtxoTxId, 1),
			Net:           entityOrder.Net,
			UtxoType:      model.UtxoTypeRewardSend,
			Amount:        uint64(entityOrder.FeeSend),
			Address:       platformAddressRewardBrc20FeeUtxos,
			PrivateKeyHex: platformPrivateKeyRewardBrc20FeeUtxos,
			TxId:          entityOrder.FeeUtxoTxId,
			Index:         1,
			PkScript:      hex.EncodeToString(pkScriptByte),
		})

		//inscribe
		time.Sleep(1500 * time.Millisecond)
		_, inscriptionId, err := inscriptionRewardForEvent(rewardType,
			utxoRewardInscriptionList,
			entityOrder.Net, entityOrder.Tick, entityOrder.RewardCoinAmount, revealOutValue,
			entityOrder.NetworkFeeRate, req.Address)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]err: %s", entityOrder.OrderId, err.Error()))
			return "", err
		}
		entityOrder.InscriptionId = inscriptionId
		entityOrder.InscriptionOutValue = revealOutValue
		entityOrder.RewardState = model.RewardStateInscription
		_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]SetPoolRewardOrderModel err: %s", entityOrder.OrderId, err.Error()))
			return "", err
		}

		//send
		sendId, err := SendReward(rewardType, utxoRewardSendList,
			entityOrder.Net, entityOrder.InscriptionId, entityOrder.InscriptionOutValue, entityOrder.Address,
			entityOrder.NetworkFeeRate)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]send err:%s", entityOrder.OrderId, err.Error()))
			return "", err
		}

		entityOrder.SendId = sendId
		entityOrder.RewardState = model.RewardStateSend
		entityOrder.FeeRawTx = ""
		_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]SetPoolRewardOrderModel err:%s", entityOrder.OrderId, err.Error()))
			return "", err
		}
	}
	return "success", nil
}
