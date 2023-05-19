package request

import "ordbook-aggregation/model"

type OrderBrc20PushReq struct {
	Tick       string           `json:"tick"`
	OrderState model.OrderState `json:"orderState"` //1-create
	OrderType  model.OrderType  `json:"orderType"`  //1-sell,2-buy
	Address    string           `json:"address"`
	PsbtRaw    string           `json:"psbtRaw"`
}

type OrderBrc20FetchReq struct {
	Tick          string           `json:"tick"`
	OrderState    model.OrderState `json:"orderState"` //1-create,2-finish,3-cancel
	OrderType     model.OrderType  `json:"orderType"`  //1-sell,2-buy
	Limit         int64            `json:"limit"`
	Flag          int64            `json:"flag"`
	SellerAddress string           `json:"sellerAddress"`
	BuyerAddress  string           `json:"buyerAddress"`
	SortKey       string           `json:"sortKey"`
	SortType      int64              `json:"sortType"`
}