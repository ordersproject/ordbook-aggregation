package order_brc20_service

import (
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
)

func PushOrder(req *request.OrderBrc20PushReq) (string, error) {
	var (
		entity *model.OrderBrc20Model
		err error
		orderId string = ""
	)
	entity = &model.OrderBrc20Model{
		OrderId:        orderId,
		Tick:           req.Tick,
		Amount:         0,
		DecimalNum:     0,
		CoinAmount:     0,
		CoinDecimalNum: 0,
		CoinRatePrice:  0,
		OrderState:     req.OrderState,
		OrderType:      req.OrderType,
		SellerAddress:  req.Address,
		BuyerAddress:   "",
		PsbtRaw:        req.PsbtRaw,
	}
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return "", err
	}
	return "success", nil
}

func FetchOrders(req *request.OrderBrc20FetchReq) (*respond.OrderResponse, error) {
	var (
		entityList []*model.OrderBrc20Model
		list []*respond.Brc20Item
		total int64 = 0
		flag int64 = 0
	)
	total, _ = mongo_service.CountOrderBrc20ModelListByHash(req.Tick, req.SellerAddress, req.BuyerAddress, req.OrderType, req.OrderState)
	entityList, _ = mongo_service.FindOrderBrc20ModelListByHash(req.Tick, req.SellerAddress, req.BuyerAddress,
		req.OrderType, req.OrderState,
		req.Limit, req.Flag, req.SortKey, req.SortType)
	list = make([]*respond.Brc20Item, len(entityList))
	for _, v := range entityList {
		item := &respond.Brc20Item{
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
			PsbtRaw:        v.PsbtRaw,
			Timestamp:      v.Timestamp,
		}
		flag = v.Timestamp
		list = append(list, item)
	}
	return &respond.OrderResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}