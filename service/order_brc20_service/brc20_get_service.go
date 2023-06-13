package order_brc20_service

import (
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
	"strconv"
)

func FetchOrders(req *request.OrderBrc20FetchReq) (*respond.OrderResponse, error) {
	var (
		entityList []*model.OrderBrc20Model
		list []*respond.Brc20Item
		total int64 = 0
		flag int64 = 0
	)
	total, _ = mongo_service.CountOrderBrc20ModelList(req.Net, req.Tick, req.SellerAddress, req.BuyerAddress, req.OrderType, req.OrderState)
	entityList, _ = mongo_service.FindOrderBrc20ModelList(req.Net, req.Tick, req.SellerAddress, req.BuyerAddress,
		req.OrderType, req.OrderState,
		req.Limit, req.Flag, req.Page, req.SortKey, req.SortType)
	list = make([]*respond.Brc20Item, len(entityList))
	for k, v := range entityList {
		item := &respond.Brc20Item{
			Net:           v.Net,
			OrderId:           v.OrderId,
			Tick:           v.Tick,
			Amount:         v.Amount,
			DecimalNum:     v.DecimalNum,
			CoinAmount:     v.CoinAmount,
			CoinDecimalNum: v.CoinDecimalNum,
			CoinRatePrice:  v.CoinRatePrice,
			OrderState:     v.OrderState,
			OrderType:      v.OrderType,
			SellerAddress:  v.SellerAddress,
			BuyerAddress:   v.BuyerAddress,
			PsbtRaw:        v.PsbtRawPreAsk,
			Timestamp:      v.Timestamp,
		}
		if req.SortKey == "coinRatePrice" {
			flag = int64(v.CoinRatePrice)
		}else {
			flag = v.Timestamp
		}
		//list = append(list, item)
		list[k] = item
	}
	return &respond.OrderResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}

func FetchTickers(req *request.TickBrc20FetchReq) (*respond.Brc20TickInfoResponse, error){
	var (
		entityList []*model.Brc20TickModel
		list []*respond.Brc20TickItem = make([]*respond.Brc20TickItem, 0)
	)

	//list = []*respond.Brc20TickItem{
	//	&respond.Brc20TickItem{
	//		Net:                "mainnet",
	//		Pair:               "ORDI-BTC",
	//		Tick:               "ordi",
	//		Buy:                "0.00005470",
	//		//Sell:               "",
	//		//Low:                "",
	//		//High:               "",
	//		//Open:               "",
	//		//Last:               "",
	//		//Volume:             "",
	//		//Amount:             "",
	//		//Vol:                "",
	//		//AvgPrice:           "",
	//		QuoteSymbol:"+",
	//		PriceChangePercent: "4.32",
	//		Ut:                 0,
	//	},
	//	&respond.Brc20TickItem{
	//		Net:                "mainnet",
	//		Pair:               "PEPE-BTC",
	//		Tick:               "pepe",
	//		Buy:                "0.00006853",
	//		QuoteSymbol:"+",
	//		PriceChangePercent: "1.43",
	//		Ut:                 0,
	//	},
	//	&respond.Brc20TickItem{
	//		Net:                "mainnet",
	//		Pair:               "MEME-BTC",
	//		Tick:               "meme",
	//		Buy:                "0.00325100",
	//		QuoteSymbol:"-",
	//		PriceChangePercent: "11.46",
	//		Ut:                 0,
	//	},
	//	&respond.Brc20TickItem{
	//		Net:                "mainnet",
	//		Pair:               "OMNI-BTC",
	//		Tick:               "omni",
	//		Buy:                "0.00004056",
	//		QuoteSymbol:"+",
	//		PriceChangePercent: "0.21",
	//		Ut:                 0,
	//	},
	//	&respond.Brc20TickItem{
	//		Net:                "mainnet",
	//		Pair:               "SATS-BTC",
	//		Tick:               "sats",
	//		Buy:                "0.00001300",
	//		QuoteSymbol:"-",
	//		PriceChangePercent: "0.35",
	//		Ut:                 0,
	//	},
	//}

	entityList, _ = mongo_service.FindBrc20TickModelList(req.Net, 0, 100)
	for _, v := range entityList {
		list = append(list, &respond.Brc20TickItem{
			Net:                v.Net,
			Tick:               v.Tick,
			Pair:               v.Pair,
			Icon:               "empty",
			Buy:                strconv.FormatUint(v.Buy, 10),
			Sell:               strconv.FormatUint(v.Sell, 10),
			Low:                strconv.FormatUint(v.Low, 10),
			High:               strconv.FormatUint(v.High, 10),
			Open:               strconv.FormatUint(v.Open, 10),
			Last:               strconv.FormatUint(v.Last, 10),
			Volume:             strconv.FormatUint(v.Volume, 10),
			Amount:             strconv.FormatUint(v.Amount, 10),
			Vol:                strconv.FormatUint(v.Vol, 10),
			AvgPrice:           strconv.FormatUint(v.AvgPrice, 10),
			QuoteSymbol:        v.QuoteSymbol,
			PriceChangePercent: strconv.FormatFloat(v.PriceChangePercent, 'f', 2, 64),
			Ut:                 v.UpdateTime,
		})
	}


	return &respond.Brc20TickInfoResponse{
		Total:   5,
		Results: list,
		Flag:    0,
	}, nil
}

func GetWsUuid(ip string) (*respond.WsUuidResp, error) {
	uuid, err := tool.GetUUID()
	if err != nil {
		return nil, err
	}
	return &respond.WsUuidResp{Uuid:uuid}, nil
}

func GetBrc20BalanceDetail(req *request.Brc20AddressReq) (*respond.BalanceDetails, error)  {
	var (
		balanceDetail *oklink_service.OklinkBrc20BalanceDetails
		err error
		list []*respond.BalanceItem = make([]*respond.BalanceItem, 0)
	)
	balanceDetail, err = oklink_service.GetAddressBrc20BalanceResult(req.Address, req.Tick, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	for _, v := range balanceDetail.TransferBalanceList {
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


func GetInscriptionOut() {

}

func GetBidDummyList(req *request.Brc20BidAddressDummyReq) (*respond.Brc20BidDummyResponse, error) {
	var(
		entityList []*model.OrderBrc20BidDummyModel
		list []*respond.DummyItem = make([]*respond.DummyItem, 0)
		total int64 = 0
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

func FetchUserOrders(req *request.Brc20OrderAddressReq) (*respond.OrderResponse, error){
	var (
		entityList []*model.OrderBrc20Model
		list []*respond.Brc20Item
		total int64 = 0
		flag int64 = 0
	)
	total, _ = mongo_service.CountAddressOrderBrc20ModelList(req.Net, req.Tick, req.Address, req.OrderType, req.OrderState)
	entityList, _ = mongo_service.FindAddressOrderBrc20ModelList(req.Net, req.Tick, req.Address,
		req.OrderType, req.OrderState,
		req.Limit, req.Flag, req.Page, req.SortKey, req.SortType)
	list = make([]*respond.Brc20Item, len(entityList))
	for k, v := range entityList {
		item := &respond.Brc20Item{
			Net:           v.Net,
			OrderId:           v.OrderId,
			Tick:           v.Tick,
			Amount:         v.Amount,
			DecimalNum:     v.DecimalNum,
			CoinAmount:     v.CoinAmount,
			CoinDecimalNum: v.CoinDecimalNum,
			CoinRatePrice:  v.CoinRatePrice,
			OrderState:     v.OrderState,
			OrderType:      v.OrderType,
			SellerAddress:  v.SellerAddress,
			BuyerAddress:   v.BuyerAddress,
			PsbtRaw:        v.PsbtRawPreAsk,
			Timestamp:      v.Timestamp,
		}
		if req.SortKey == "coinRatePrice" {
			flag = int64(v.CoinRatePrice)
		}else {
			flag = v.Timestamp
		}

		//list = append(list, item)
		list[k] = item
	}
	return &respond.OrderResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}