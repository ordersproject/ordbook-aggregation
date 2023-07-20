package order_brc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
	"strconv"
)

func FetchPoolPairInfo(req *request.PoolPairFetchOneReq) (*respond.PoolInfoResponse, error) {
	var (
		total      int64 = 0
		entityList []*model.PoolInfoModel
		//entity *model.PoolInfoModel
		list []*respond.PoolInfoItem = make([]*respond.PoolInfoItem, 0)
	)
	//entity, _ = mongo_service.FindPoolInfoModelByPair(req.Net, strings.ToUpper(req.Pair))
	//if entity == nil || entity.Id == 0 {
	//	return nil, errors.New("pool info ie empty")
	//}
	//return &respond.PoolInfoResponse{
	//	Net:            entity.Net,
	//	Tick:           entity.Tick,
	//	Pair:           entity.Pair,
	//	CoinAmount:     entity.CoinAmount,
	//	CoinDecimalNum: entity.CoinDecimalNum,
	//	Amount:         entity.Amount,
	//	DecimalNum:     entity.DecimalNum,
	//}, nil

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
			Net:           v.Net,
			OrderId:       v.OrderId,
			Tick:          v.Tick,
			Pair:          v.Pair,
			Amount:        v.Amount,
			DecimalNum:    v.DecimalNum,
			PoolType:      v.PoolType,
			PoolState:     v.PoolState,
			Address:       v.Address,
			InscriptionId: v.InscriptionId,
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
		Net:           entity.Net,
		OrderId:       entity.OrderId,
		Tick:          entity.Tick,
		Amount:        entity.Amount,
		DecimalNum:    entity.DecimalNum,
		PoolType:      entity.PoolType,
		PoolState:     entity.PoolState,
		Address:       entity.Address,
		InscriptionId: entity.InscriptionId,
		PsbtRaw:       entity.PsbtRaw,
		Timestamp:     entity.Timestamp,
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
		netParams     *chaincfg.Params = GetNetParams(req.Net)
		entity        *model.PoolBrc20Model
		err           error
		orderId       string = ""
		psbtBuilder   *PsbtBuilder
		address       string = ""
		coinAmount    uint64 = 0
		coinDec       int    = 18
		outAmount     uint64 = 0
		amountDec     int    = 8
		coinRatePrice uint64 = 0
		inscriptionId string = ""

		_, platformPublicKeyMultiSig = GetPlatformKeyMultiSig(req.Net)

		multiSigAddress       string = ""
		multiSigSegWitAddress string = ""
	)

	multiSigAddress, multiSigSegWitAddress, err = createMultiSigAddress(netParams, []string{publicKey, platformPublicKeyMultiSig}...)
	if err != nil {
		return "", err
	}
	_ = multiSigSegWitAddress

	psbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
	if err != nil {
		return "", err
	}
	switch req.PoolType {
	case model.PoolTypeTick:
		var (
			inscriptionBrc20BalanceItem *oklink_service.BalanceItem
			has                         = false
		)
		address = req.Address
		coinAmount = req.CoinAmount

		preOutList := psbtBuilder.GetInputs()
		if preOutList == nil || len(preOutList) == 0 {
			return "", errors.New("Wrong Psbt: empty inputs. ")
		}
		for _, v := range preOutList {
			inscriptionId = fmt.Sprintf("%s:%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
			inscriptionBrc20BalanceItem, err = CheckBrc20Ordinals(v, req.Tick, address)
			if err != nil {
				continue
			}
			has = true
		}

		if req.Net == "mainnet" || req.Net == "livenet" {
			if !has || inscriptionBrc20BalanceItem == nil {
				return "", errors.New("Wrong Psbt: Empty inscription. ")
			}
			coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
		}
		verified, err := CheckPublicKeyAddress(netParams, publicKey, address)
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
			outAmount = uint64(v.Value)
			_, addrs, _, err := txscript.ExtractPkScriptAddrs(v.PkScript, netParams)
			if err != nil {
				return "", err
			}
			if addrs[0].EncodeAddress() != multiSigSegWitAddress {
				return "", errors.New("Wrong Psbt: wrong multiAddress of out for pool psbt. ")
			}
		}

		outAmountDe := decimal.NewFromInt(int64(outAmount))
		coinAmountDe := decimal.NewFromInt(int64(coinAmount))
		coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
		coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)

		orderId = fmt.Sprintf("%s_%s_%s_%s_%d_%d", req.Net, req.Tick, inscriptionId, address, outAmount, coinAmount)
		orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
		break
	case model.PoolTypeBtc:
		return "", errors.New("Not yet. ")
		break
	case model.PoolTypeBoth:
		return "", errors.New("Not yet. ")
		break
	default:
		return "", errors.New("Wrong OrderState. ")
	}

	//todo Fix pool
	entity = &model.PoolBrc20Model{
		Net:            req.Net,
		OrderId:        orderId,
		Tick:           req.Tick,
		CoinAmount:     coinAmount,
		CoinDecimalNum: coinDec,
		Amount:         outAmount,
		DecimalNum:     amountDec,
		CoinRatePrice:  coinRatePrice,
		CoinAddress:    address,
		Address:        multiSigAddress,
		CoinPsbtRaw:    req.PsbtRaw,
		PsbtRaw:        "",
		InscriptionId:  inscriptionId,
		UtxoId:         "",
		PoolType:       req.PoolType,
		PoolState:      req.PoolState,
		Timestamp:      tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetPoolBrc20Model(entity)
	if err != nil {
		return "", err
	}

	return "success", nil
}
