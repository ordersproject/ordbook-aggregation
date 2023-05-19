package respond

import "ordbook-aggregation/model"

type OrderResponse struct {
	Total   int64             `json:"total,omitempty"`
	Results []*Brc20Item `json:"results,omitempty"`
	Flag    int64               `json:"flag,omitempty"`
}

type Brc20Item struct {
	Tick           string     `json:"tick,omitempty"`
	Amount         uint64     `json:"amount,omitempty"`
	DecimalNum     int        `json:"decimalNum,omitempty"`
	CoinAmount     uint64     `json:"coinAmount,omitempty"`
	CoinDecimalNum int        `json:"coinDecimalNum,omitempty"`
	CoinRatePrice  uint64     `json:"coinRatePrice,omitempty"`
	OrderState     model.OrderState `json:"orderState,omitempty"` //1-create,2-finish,3-cancel
	OrderType      model.OrderType  `json:"orderType,omitempty"`   //1-sell,2-buy
	SellerAddress  string     `json:"sellerAddress,omitempty"`
	BuyerAddress   string     `json:"buyerAddress,omitempty"`
	PsbtRaw        string     `json:"psbtRaw,omitempty"`
	Timestamp      int64      `json:"timestamp,omitempty"`
}