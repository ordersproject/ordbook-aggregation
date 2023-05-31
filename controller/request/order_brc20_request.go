package request

import "ordbook-aggregation/model"

type OrderBrc20PushReq struct {
	Net        string           `json:"net"` //mainnet/signet/testnet
	Tick       string           `json:"tick"`
	OrderState model.OrderState `json:"orderState"` //1-create
	OrderType  model.OrderType  `json:"orderType"`  //1-sell,2-buy
	Address    string           `json:"address"`
	PsbtRaw    string           `json:"psbtRaw"`
	CoinAmount uint64           `json:"coinAmount"`
}

type OrderBrc20FetchReq struct {
	Net        string           `json:"net"` //mainnet/signet/testnet
	Tick          string           `json:"tick"`
	OrderState    model.OrderState `json:"orderState"` //1-create,2-finish,3-cancel
	OrderType     model.OrderType  `json:"orderType"`  //1-sell,2-buy
	Limit         int64            `json:"limit"`
	Flag          int64            `json:"flag"`
	SellerAddress string           `json:"sellerAddress"`
	BuyerAddress  string           `json:"buyerAddress"`
	SortKey       string           `json:"sortKey"`//coinRatePrice/timestamp
	SortType      int64              `json:"sortType"`//1/-1
}

type TickBrc20FetchReq struct {
	Net        string           `json:"net"` //mainnet/signet/testnet
	Tick     string `json:"tick"`
	Limit    int64  `json:"limit"`
	Flag     int64  `json:"flag"`
	SortKey  string `json:"sortKey"`
	SortType int64  `json:"sortType"`
}

type TickKlineFetchReq struct {
	Net        string           `json:"net"` //mainnet/signet/testnet
	Tick     string `json:"tick"`
	Limit    int64  `json:"limit"`//默认1000
	Interval string `json:"interval"` //1m/1s/15m/1h/4h/1d/1w/
}

type OrderBrc20GetBidReq struct {
	Net               string `json:"net"` //mainnet/signet/testnet
	Pair              string `json:"pair"`
	Tick              string `json:"tick"`
	Amount            uint64 `json:"amount"`
	Address           string `json:"address"`
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
	CoinAmount        string `json:"amount"`
}

type OrderBrc20UpdateBidReq struct {
	Net     string `json:"net"` //mainnet/signet/testnet
	Tick    string `json:"tick"`
	Amount  uint64 `json:"amount"`
	Address string `json:"address"`
	OrderId string `json:"orderId"`
	PsbtRaw string `json:"psbtRaw"`
}

type OrderBrc20UpdateReq struct {
	Net        string           `json:"net"` //mainnet/signet/testnet
	OrderId    string           `json:"orderId"`
	OrderState model.OrderState `json:"orderState"` //2-finish/3-cancel
	PsbtRaw    string           `json:"psbtRaw"`
}


type OrderBrc20DoBidReq struct {
	Net               string `json:"net"` //mainnet/signet/testnet
	Tick              string `json:"tick"`
	OrderId           string `json:"orderId"`
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
	CoinAmount        string `json:"amount"`
	PsbtRaw           string `json:"psbtRaw"`
}