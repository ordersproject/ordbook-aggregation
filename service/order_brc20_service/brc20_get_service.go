package order_brc20_service

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/common_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
)

const (
	dayLimit int64 = 2
)

//todo whitelist

func FetchOneOrders(req *request.OrderBrc20FetchOneReq, publicKey, ip string) (*respond.Brc20Item, error) {
	var (
		entity                       *model.OrderBrc20Model
		netParams                    *chaincfg.Params = GetNetParams(req.Net)
		count                        int64            = 0
		todayStartTime, todayEndTime int64            = tool.GetToday0Time(), tool.GetToday24Time()
		takerPsbtRaw                 string           = ""
	)
	entity, _ = mongo_service.FindOrderBrc20ModelByOrderId(req.OrderId)
	if entity == nil {
		return nil, errors.New("Order is empty. ")
	}
	netParams = GetNetParams(entity.Net)

	verified, err := CheckPublicKeyAddress(netParams, publicKey, req.BuyerAddress)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
	}
	if !verified {
		return nil, errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
	}

	if entity.FreeState == model.FreeStateYes {
		count, _ = mongo_service.CountBuyerOrderBrc20ModelList(entity.Net, entity.Tick, req.BuyerAddress, "", model.OrderTypeSell, model.OrderStateFinish, todayStartTime, todayEndTime, 0)
		fmt.Printf("[LIMIT-address]-%s-%s-count[%d]\n\n", ip, req.BuyerAddress, count)
		if count >= dayLimit {
			return nil, errors.New(fmt.Sprintf("The number of purchases of the day has exceeded. "))
		}
		count, _ = mongo_service.CountBuyerOrderBrc20ModelList(entity.Net, entity.Tick, "", ip, model.OrderTypeSell, model.OrderStateFinish, todayStartTime, todayEndTime, 0)
		fmt.Printf("[LIMIT-ip]-%s-%s-count[%d]\n\n", ip, req.BuyerAddress, count)
		if count >= dayLimit {
			return nil, errors.New(fmt.Sprintf("The number of purchases of the day has exceeded. "))
		}
	}

	if entity.PlatformDummy == model.PlatformDummyYes {
		takerPsbtRaw, err = MakeAskTakerPsbtRaw(entity.Net, entity.PsbtRawPreAsk, req.BuyerAddress, req.BuyerChangeAmount)
		if err != nil {
			return nil, err
		}
	}

	item := &respond.Brc20Item{
		Net:                 entity.Net,
		OrderId:             entity.OrderId,
		Tick:                entity.Tick,
		Amount:              entity.Amount,
		DecimalNum:          entity.DecimalNum,
		CoinAmount:          entity.CoinAmount,
		CoinDecimalNum:      entity.CoinDecimalNum,
		CoinRatePrice:       entity.CoinRatePrice,
		CoinPrice:           entity.CoinPrice,
		CoinPriceDecimalNum: entity.CoinPriceDecimalNum,
		OrderState:          entity.OrderState,
		OrderType:           entity.OrderType,
		FreeState:           entity.FreeState,
		SellerAddress:       entity.SellerAddress,
		BuyerAddress:        entity.BuyerAddress,
		PsbtRaw:             entity.PsbtRawPreAsk,
		TakePsbtRaw:         takerPsbtRaw,
		Timestamp:           entity.Timestamp,
	}
	return item, nil
}

func FetchOrders(req *request.OrderBrc20FetchReq) (*respond.OrderResponse, error) {
	var (
		entityList []*model.OrderBrc20Model
		list       []*respond.Brc20Item
		total      int64 = 0
		flag       int64 = 0
	)
	if req.Limit < 0 || req.Limit >= 1000 {
		req.Limit = 1000
	}
	total, _ = mongo_service.CountOrderBrc20ModelList(req.Net, req.Tick, req.SellerAddress, req.BuyerAddress, req.OrderType, req.OrderState)
	entityList, _ = mongo_service.FindOrderBrc20ModelList(req.Net, req.Tick, req.SellerAddress, req.BuyerAddress,
		req.OrderType, req.OrderState,
		req.Limit, req.Flag, req.Page, req.SortKey, req.SortType, 0, 0)
	list = make([]*respond.Brc20Item, 0)
	for _, v := range entityList {
		if req.Address != "" && v.PoolOrderId != "" {
			poolOwner := checkPoolAddress(v.PoolOrderId, req.Address)
			if poolOwner > 0 {
				continue
			}

			if v.OrderType == model.OrderTypeBuy && req.Address != v.BuyerAddress {
				if checkPoolType(v.PoolOrderId) != model.PoolTypeBoth {
					continue
				}
			}
		}

		item := &respond.Brc20Item{
			Net:                 v.Net,
			OrderId:             v.OrderId,
			Tick:                v.Tick,
			Amount:              v.Amount,
			DecimalNum:          v.DecimalNum,
			CoinAmount:          v.CoinAmount,
			CoinDecimalNum:      v.CoinDecimalNum,
			CoinRatePrice:       v.CoinRatePrice,
			CoinPrice:           v.CoinPrice,
			CoinPriceDecimalNum: v.CoinPriceDecimalNum,
			OrderState:          v.OrderState,
			OrderType:           v.OrderType,
			FreeState:           v.FreeState,
			SellerAddress:       v.SellerAddress,
			BuyerAddress:        v.BuyerAddress,
			//PsbtRaw:        v.PsbtRawPreAsk,
			Timestamp: v.Timestamp,
		}
		if req.SortKey == "coinRatePrice" {
			flag = int64(v.CoinRatePrice)
		} else if req.SortKey == "coinPrice" {
			flag = int64(v.CoinPrice)
		} else {
			flag = v.Timestamp
		}
		list = append(list, item)
		//list[k] = item
	}
	return &respond.OrderResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}

func FetchTickers(req *request.TickBrc20FetchReq) (*respond.Brc20TickInfoResponse, error) {
	var (
		entityList []*model.Brc20TickModel
		list       []*respond.Brc20TickItem = make([]*respond.Brc20TickItem, 0)
	)

	_ = entityList
	entityList, _ = mongo_service.FindBrc20TickModelVersionList(req.Net, req.Tick, 0, 100, 2)
	for _, v := range entityList {

		//coinPriceDe := decimal.NewFromInt(int64(v.CoinPrice))
		coinPriceDe := decimal.NewFromInt(int64(v.LastTop))
		coinPriceDe = coinPriceDe.Div(decimal.New(1, 8))
		UpdateMarketPriceV2(v.Net, v.Tick, v.Pair)

		avgPrice := coinPriceDe.String()
		if coinPriceDe.Cmp(decimal.NewFromInt(1)) > 0 {
			avgPrice = coinPriceDe.StringFixed(0)
		}

		item := &respond.Brc20TickItem{
			Net:    v.Net,
			Tick:   v.Tick,
			Pair:   v.Pair,
			Icon:   "empty",
			Buy:    strconv.FormatUint(v.Buy, 10),
			Sell:   strconv.FormatUint(v.Sell, 10),
			Low:    strconv.FormatUint(v.Low, 10),
			High:   strconv.FormatUint(v.High, 10),
			Open:   strconv.FormatUint(v.Open, 10),
			Last:   strconv.FormatUint(v.Last, 10),
			Volume: strconv.FormatUint(v.Volume, 10),
			Amount: strconv.FormatUint(v.Amount, 10),
			Vol:    strconv.FormatUint(v.Vol, 10),
			//AvgPrice:           strconv.FormatUint(v.AvgPrice, 10),
			//AvgPrice: strconv.FormatUint(v.Last, 10),
			//AvgPrice:            coinPriceDe.String(),
			AvgPrice:            avgPrice,
			CoinPrice:           v.CoinPrice,
			CoinPriceDecimalNum: v.CoinPriceDecimalNum,
			QuoteSymbol:         v.QuoteSymbol,
			PriceChangePercent:  strconv.FormatFloat(v.PriceChangePercent, 'f', 2, 64),
			Ut:                  v.UpdateTime,
		}
		if v.AvgPrice == 0 || v.LastTotal < 5 {
			priceInfo := getOtherMarketPrice(v.Tick)
			if priceInfo != nil {
				item.AvgPrice = priceInfo.VisionPrice
			}
		}

		list = append(list, item)
	}

	for tick, priceInfo := range common_service.Brc20TickMarketDataMap {
		if req.Tick != "" {
			if tick != req.Tick {
				continue
			}
		}

		has := false
		for _, v := range list {
			if v.Tick == tick {
				has = true
				break
			}
		}
		if has {
			continue
		}
		if priceInfo == nil || priceInfo.UpdateTime == 0 {
			priceInfo = getOtherMarketPrice(tick)
		}
		list = append(list, &respond.Brc20TickItem{
			Net:  req.Net,
			Tick: tick,
			Pair: fmt.Sprintf("%s-BTC", strings.ToUpper(tick)),
			Icon: "empty",
			//Buy:                strconv.FormatUint(v.Buy, 10),
			//Sell:               strconv.FormatUint(v.Sell, 10),
			//Low:                strconv.FormatUint(v.Low, 10),
			//High:               strconv.FormatUint(v.High, 10),
			//Open:               strconv.FormatUint(v.Open, 10),
			//Last:               strconv.FormatUint(v.Last, 10),
			//Volume:             strconv.FormatUint(v.Volume, 10),
			//Amount:             strconv.FormatUint(v.Amount, 10),
			//Vol:                strconv.FormatUint(v.Vol, 10),
			AvgPrice: priceInfo.VisionPrice,
			//QuoteSymbol:        v.QuoteSymbol,
			//PriceChangePercent: strconv.FormatFloat(v.PriceChangePercent, 'f', 2, 64),
			Ut: priceInfo.UpdateTime,
		})
	}

	return &respond.Brc20TickInfoResponse{
		//Total:   5,
		Total:   int64(len(common_service.Brc20TickMarketDataMap)),
		Results: list,
		Flag:    0,
	}, nil
}

func GetWsUuid(ip string) (*respond.WsUuidResp, error) {
	uuid, err := tool.GetUUID()
	if err != nil {
		return nil, err
	}
	return &respond.WsUuidResp{Uuid: uuid}, nil
}

// GetBrc20BalanceDetail get brc20 token detail
func GetBrc20BalanceDetail(req *request.Brc20AddressReq) (*respond.BalanceDetails, error) {
	var (
		balanceDetail *oklink_service.OklinkBrc20BalanceDetails
		err           error
		list          []*respond.BalanceItem = make([]*respond.BalanceItem, 0)
		//utxoList      []*unisat_service.UtxoDetailItem
	)
	balanceDetail, err = oklink_service.GetAddressBrc20BalanceResult(req.Address, req.Tick, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	//utxoList, _ = unisat_service.GetAddressUtxo(req.Address)

	for _, v := range balanceDetail.TransferBalanceList {
		//fmt.Printf("Transfer:[%s]\n", v.InscriptionId)
		//has := false
		//if utxoList != nil && len(utxoList) != 0 {
		//	for _, u := range utxoList {
		//		inscriptionId := fmt.Sprintf("%si%d", u.TxId, u.OutputIndex)
		//		fmt.Printf("Live inscriptionId:[%s]\n", inscriptionId)
		//		if inscriptionId == v.InscriptionId {
		//			has = true
		//			break
		//		}
		//	}
		//}
		//if has {
		//	list = append(list, &respond.BalanceItem{
		//		InscriptionId:     v.InscriptionId,
		//		InscriptionNumber: v.InscriptionNumber,
		//		Amount:            v.Amount,
		//	})
		//}

		//check order which is sold
		soldOrderCount, _ := mongo_service.FindSoldInscriptionOrder(v.InscriptionId)
		if soldOrderCount != 0 {
			fmt.Printf("Used Inscription soldOrderCount: [%s]\n", v.InscriptionId)
			continue
		}

		//check order which is used
		usedOrderCount, _ := mongo_service.FindUsedInscriptionOrder(v.InscriptionId)
		if usedOrderCount != 0 {
			fmt.Printf("Used Inscription usedOrderCount: [%s]\n", v.InscriptionId)
			continue
		}
		usedOrderCount, _ = mongo_service.FindUsedInscriptionOrderV2(v.InscriptionId)
		if usedOrderCount != 0 {
			fmt.Printf("Used Inscription usedOrderCount2: [%s]\n", v.InscriptionId)
			continue
		}

		//check pool which is used
		usedCount, _ := mongo_service.FindUsedInscriptionPool(v.InscriptionId)
		if usedCount != 0 {
			fmt.Printf("Used InscriptionPool: [%s]\n", v.InscriptionId)
			continue
		}

		list = append(list, &respond.BalanceItem{
			InscriptionId:     v.InscriptionId,
			InscriptionNumber: v.InscriptionNumber,
			Amount:            v.Amount,
		})
	}

	return &respond.BalanceDetails{
		Page:                balanceDetail.Page,
		Limit:               balanceDetail.Limit,
		TotalPage:           balanceDetail.TotalPage,
		Token:               balanceDetail.Token,
		TokenType:           balanceDetail.TokenType,
		Balance:             balanceDetail.Balance,
		AvailableBalance:    balanceDetail.AvailableBalance,
		TransferBalance:     balanceDetail.TransferBalance,
		TransferBalanceList: list,
	}, nil
}

// GetBrc20BalanceList get brc20 token list
func GetBrc20BalanceList(req *request.Brc20AddressReq) (*respond.Brc20BalanceList, error) {
	var (
		balanceListResp *oklink_service.OklinkBrc20BalanceList
		err             error
		list            []*respond.BalanceListItem = make([]*respond.BalanceListItem, 0)
	)
	balanceListResp, err = oklink_service.GetAddressBrc20BalanceListResult(req.Address, req.Tick, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	for _, v := range balanceListResp.BalanceList {
		list = append(list, &respond.BalanceListItem{
			Token:            v.Token,
			TokenType:        v.TokenType,
			Balance:          v.Balance,
			AvailableBalance: v.AvailableBalance,
			TransferBalance:  v.TransferBalance,
		})
	}

	return &respond.Brc20BalanceList{
		Page:        balanceListResp.Page,
		Limit:       balanceListResp.Limit,
		TotalPage:   balanceListResp.TotalPage,
		BalanceList: list,
	}, nil
}

func GetBidDummyList(req *request.Brc20BidAddressDummyReq) (*respond.Brc20BidDummyResponse, error) {
	var (
		entityList []*model.OrderBrc20BidDummyModel
		list       []*respond.DummyItem = make([]*respond.DummyItem, 0)
		total      int64                = 0
	)
	total, _ = mongo_service.CountOrderBrc20BidDummyModelList("", req.Address, model.DummyStateLive)
	entityList, _ = mongo_service.FindOrderBrc20BidDummyModelList("", req.Address, model.DummyStateLive, req.Skip, req.Limit)
	for _, v := range entityList {
		list = append(list, &respond.DummyItem{
			Order:     v.OrderId,
			DummyId:   v.DummyId,
			Timestamp: v.Timestamp,
		})
	}
	return &respond.Brc20BidDummyResponse{
		Total:   total,
		Results: list,
		Flag:    0,
	}, nil
}

func FetchUserOrders(req *request.Brc20OrderAddressReq) (*respond.OrderResponse, error) {
	var (
		entityList []*model.OrderBrc20Model
		list       []*respond.Brc20Item
		total      int64 = 0
		flag       int64 = 0
	)
	total, _ = mongo_service.CountAddressOrderBrc20ModelList(req.Net, req.Tick, req.Address, req.OrderType, req.OrderState)
	entityList, _ = mongo_service.FindAddressOrderBrc20ModelList(req.Net, req.Tick, req.Address,
		req.OrderType, req.OrderState,
		req.Limit, req.Flag, req.Page, req.SortKey, req.SortType)
	list = make([]*respond.Brc20Item, len(entityList))
	for k, v := range entityList {
		item := &respond.Brc20Item{
			Net:                 v.Net,
			OrderId:             v.OrderId,
			Tick:                v.Tick,
			Amount:              v.Amount,
			DecimalNum:          v.DecimalNum,
			CoinAmount:          v.CoinAmount,
			CoinDecimalNum:      v.CoinDecimalNum,
			CoinRatePrice:       v.CoinRatePrice,
			CoinPrice:           v.CoinPrice,
			CoinPriceDecimalNum: v.CoinPriceDecimalNum,
			OrderState:          v.OrderState,
			OrderType:           v.OrderType,
			FreeState:           v.FreeState,
			SellerAddress:       v.SellerAddress,
			BuyerAddress:        v.BuyerAddress,
			InscriptionId:       v.InscriptionId,
			//PsbtRaw:        v.PsbtRawPreAsk,
			Timestamp: v.Timestamp,
		}
		if req.SortKey == "coinRatePrice" {
			flag = int64(v.CoinRatePrice)
		} else if req.SortKey == "coinPrice" {
			flag = int64(v.CoinPrice)
		} else {
			flag = v.Timestamp
		}

		list[k] = item
	}
	return &respond.OrderResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}

func FetchTickerInfo(req *request.TickBrc20FetchReq) {

}

func FetchTickKline(req *request.TickKlineFetchReq) (*respond.Brc20KlineInfo, error) {
	var (
		entityList         []*model.Brc20TickKlineModel
		list               []*respond.KlineItem = make([]*respond.KlineItem, 0)
		startTime, endTime int64                = 0, tool.MakeTimestamp() //1m/1s/15m/1h/4h/1d/1w/
		limit              int64                = req.Limit
		dis                int64                = 1000 * 60 * 15
	)
	if req.Flag != 0 {
		endTime = req.Flag
	}
	if req.Limit == 0 {
		limit = 100
	}
	switch req.Interval {
	case "15m":
		startTime = endTime - limit*dis
		break
	case "1h":
		startTime = endTime - limit*dis*4
		break
	case "4h":
		startTime = endTime - limit*dis*4*4
		break
	case "1d":
		startTime = endTime - limit*dis*4*24
		break
	case "1w":
		startTime = endTime - limit*dis*4*24*7
		break
	default:
		startTime = endTime - limit*dis
	}
	//fmt.Printf("%s-%s, %s-%s\n", req.Net, req.Tick, tool.MakeDate(startTime), tool.MakeDate(endTime))
	entityList, _ = mongo_service.FindBrc20TickKlineModelList(req.Net, req.Tick, startTime, endTime)
	for _, v := range entityList {
		list = append(list, &respond.KlineItem{
			Timestamp: v.Timestamp,
			Open:      v.Open,
			High:      v.High,
			Low:       v.Low,
			Close:     v.Close,
			Volume:    v.Volume,
		})
	}
	return &respond.Brc20KlineInfo{
		Net:      req.Net,
		Tick:     req.Tick,
		Interval: req.Interval,
		List:     list,
	}, nil
}

func FetchTickRecentlyInfo(req *request.TickRecentlyInfoFetchReq) {
	//var (
	//	entityList         []*model.OrderBrc20Model
	//	list               []*respond.KlineItem = make([]*respond.KlineItem, 0)
	//	startTime, endTime int64                = 0, tool.MakeTimestamp() //1m/1s/15m/1h/4h/1d/1w/
	//	limit              int64                = req.Limit
	//	dis                int64                = 1000 * 60 * 15
	//)
}

func FetchEventOrders(req *request.Brc20EventOrderReq) (*respond.OrderEventResponse, error) {
	var (
		entityList []*model.OrderBrc20Model
		list       []*respond.Brc20EventItem
		total      int64 = 0
		flag       int64 = 0
	)
	if req.Limit < 0 || req.Limit >= 1000 {
		req.Limit = 1000
	}
	total, _ = mongo_service.CountEventOrderBrc20ModelList(req.Net, req.Tick, req.Address,
		model.OrderTypeBuy, model.OrderStateFinish, 2,
		config.EventOneStartTime)
	entityList, _ = mongo_service.FindEventOrderBrc20ModelList(req.Net, req.Tick, req.Address,
		model.OrderTypeBuy, model.OrderStateFinish,
		req.Limit, req.Page, 2,
		config.EventOneStartTime)
	list = make([]*respond.Brc20EventItem, len(entityList))
	for k, v := range entityList {
		item := &respond.Brc20EventItem{
			Net:                 v.Net,
			OrderId:             v.OrderId,
			Tick:                v.Tick,
			Amount:              v.Amount,
			DecimalNum:          v.DecimalNum,
			CoinAmount:          v.CoinAmount,
			CoinDecimalNum:      v.CoinDecimalNum,
			CoinRatePrice:       v.CoinRatePrice,
			CoinPrice:           v.CoinPrice,
			CoinPriceDecimalNum: v.CoinPriceDecimalNum,
			OrderState:          v.OrderState,
			OrderType:           v.OrderType,
			FreeState:           v.FreeState,
			SellerAddress:       v.SellerAddress,
			BuyerAddress:        v.BuyerAddress,
			//InscriptionId:       v.InscriptionId,
			//PsbtRaw:        v.PsbtRawPreAsk,
			Timestamp:        v.Timestamp,
			DealTxBlockState: v.DealTxBlockState,
			DealTxBlock:      v.DealTxBlock,
			Percentage:       v.Percentage,
			CalValue:         v.CalValue,
			CalTotalValue:    v.CalTotalValue,
			CalStartBlock:    v.CalStartBlock,
			CalEndBlock:      v.CalEndBlock,
			RewardAmount:     v.RewardAmount / 2,
			RewardRealAmount: v.RewardRealAmount / 2,
			Version:          v.Version,
		}
		//list = append(list, item)
		list[k] = item
	}
	return &respond.OrderEventResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}
