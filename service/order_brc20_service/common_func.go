package order_brc20_service

import (
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
)

func UpdateMarketPrice(net, tick, pair string) *model.Brc20TickModel{
	var(
		askList []*model.OrderBrc20Model
		bidList []*model.OrderBrc20Model
		marketPrice uint64 = 0
		totalPrice uint64 = 0
		total uint64 = 0
		tickInfo *model.Brc20TickModel
		sellPrice uint64 = 0
		sellTotal uint64 = 0
		buyPrice uint64 = 0
		buyTotal uint64 = 0
	)
	askList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "", model.OrderTypeSell, model.OrderStateCreate, 10, 0,
		"coinRatePrice", 1)
	bidList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "", model.OrderTypeBuy, model.OrderStateCreate, 10, 0,
		"coinRatePrice", -1)
	for _, v := range askList{
		if v.CoinRatePrice == 0 {
			continue
		}
		sellPrice = sellPrice + v.CoinRatePrice
		totalPrice = totalPrice + v.CoinRatePrice
		total++
		sellTotal++
	}
	sellPrice = sellPrice/sellTotal

	for _, v := range bidList{
		if v.CoinRatePrice == 0 {
			continue
		}
		buyPrice = buyPrice + v.CoinRatePrice
		totalPrice = totalPrice + v.CoinRatePrice
		total++
		buyTotal++
	}
	buyPrice = buyPrice/buyTotal
	marketPrice = totalPrice/total

	tickInfo, _ = mongo_service.FindBrc20TickModelByPair(pair)
	if tickInfo == nil {
		tickInfo = &model.Brc20TickModel{
			Net:                net,
			Tick:               tick,
			Pair:               pair,
			Buy:                buyPrice,
			Sell:               sellPrice,
			AvgPrice:           marketPrice,
		}
	}
	tickInfo.Buy = buyPrice
	tickInfo.Sell = sellPrice
	tickInfo.AvgPrice = marketPrice
	_, err := mongo_service.SetBrc20TickModel(tickInfo)
	if err != nil {
		return nil
	}
	return tickInfo
}

func GetMarketPrice(net, tick, pair string) uint64 {
	tickInfo, _ := mongo_service.FindBrc20TickModelByPair(pair)
	if tickInfo == nil {
		tickInfo = UpdateMarketPrice(net, tick, pair)
	}
	if tickInfo == nil {
		return 0
	}
	return tickInfo.AvgPrice
}