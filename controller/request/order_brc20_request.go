package request

import "ordbook-aggregation/model"

type OrderBrc20PushReq struct {
	Net        string           `json:"net"` //livenet/signet/testnet
	Tick       string           `json:"tick"`
	OrderState model.OrderState `json:"orderState"` //1-create
	OrderType  model.OrderType  `json:"orderType"`  //1-sell,2-buy
	Address    string           `json:"address"`
	PsbtRaw    string           `json:"psbtRaw"`
	CoinAmount uint64           `json:"coinAmount"`
}

type OrderBrc20FetchReq struct {
	Net           string           `json:"net"` //livenet/signet/testnet
	Tick          string           `json:"tick"`
	OrderState    model.OrderState `json:"orderState"` //1-create,2-finish,3-cancel
	OrderType     model.OrderType  `json:"orderType"`  //1-sell,2-buy
	Limit         int64            `json:"limit"`
	Flag          int64            `json:"flag"`
	Page          int64            `json:"page"`
	SellerAddress string           `json:"sellerAddress"`
	BuyerAddress  string           `json:"buyerAddress"`
	SortKey       string           `json:"sortKey"`  //coinRatePrice/timestamp
	SortType      int64            `json:"sortType"` //1/-1
}

type OrderBrc20FetchOneReq struct {
	Net          string `json:"net"` //livenet/signet/testnet
	Tick         string `json:"tick"`
	OrderId      string `json:"orderId"`
	BuyerAddress string `json:"buyerAddress"`
}

type TickBrc20FetchReq struct {
	Net      string `json:"net"` //livenet/signet/testnet
	Tick     string `json:"tick"`
	Limit    int64  `json:"limit"`
	Flag     int64  `json:"flag"`
	SortKey  string `json:"sortKey"`
	SortType int64  `json:"sortType"`
}

type TickKlineFetchReq struct {
	Net      string `json:"net"` //livenet/signet/testnet
	Tick     string `json:"tick"`
	Limit    int64  `json:"limit"`    //默认1000
	Flag     int64  `json:"flag"`     //
	Interval string `json:"interval"` //1m/1s/15m/1h/4h/1d/1w/
}

type OrderBrc20GetBidReq struct {
	Net               string `json:"net"` //livenet/signet/testnet
	Pair              string `json:"pair"`
	Tick              string `json:"tick"`
	Amount            uint64 `json:"amount"`
	Address           string `json:"address"`
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
	CoinAmount        string `json:"coinAmount"`
	IsPool            bool   `json:"isPool"`
	PoolOrderId       string `json:"poolOrderId"`
	Limit             int64  `json:"limit"`
	Page              int64  `json:"page"`
}

type OrderBrc20UpdateBidReq struct {
	Net          string `json:"net"` //livenet/signet/testnet
	Tick         string `json:"tick"`
	Amount       uint64 `json:"amount"`       //the purchase value of input
	BuyerInValue uint64 `json:"buyerInValue"` //the real value of input
	Address      string `json:"address"`
	OrderId      string `json:"orderId"`
	PsbtRaw      string `json:"psbtRaw"`
	Rate         int    `json:"rate"` //sats/B
	Fee          uint64 `json:"fee"`  //fee
}

type OrderBrc20UpdateReq struct {
	Net            string           `json:"net"` //livenet/signet/testnet
	OrderId        string           `json:"orderId"`
	OrderState     model.OrderState `json:"orderState"` //2-finish/3-cancel
	PsbtRaw        string           `json:"psbtRaw"`
	BroadcastIndex int              `json:"broadcastIndex"` //1

	Address string `json:"address"`
}

type OrderBrc20DoBidReq struct {
	Net        string `json:"net"` //livenet/signet/testnet
	OrderId    string `json:"orderId"`
	PsbtRaw    string `json:"psbtRaw"`
	Value      uint64 `json:"value"`
	Address    string `json:"address"`
	CoinAmount string `json:"amount"`

	Tick              string `json:"tick"`
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
}

type CheckBrc20InscriptionReq struct {
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
	//PreTxId           string `json:"preTxId"`
	//PreIndex          int64  `json:"preIndex"`
}

type Brc20AddressReq struct {
	Net     string `json:"net"` //livenet/signet/testnet
	Tick    string `json:"tick"`
	Address string `json:"address"`
	Page    int64  `json:"page"`
	Limit   int64  `json:"limit"`
}

type Brc20BidAddressDummyReq struct {
	Net     string `json:"net"` //livenet/signet/testnet
	Tick    string `json:"tick"`
	Address string `json:"address"`
	Skip    int64  `json:"skip"`
	Limit   int64  `json:"limit"`
}

type Brc20MarketPriceSetReq struct {
	Net        string `json:"net"` //livenet/signet/testnet
	Tick       string `json:"tick"`
	Pair       string `json:"pair"`
	GuidePrice int64  `json:"guidePrice"`
}

type Brc20OrderAddressReq struct {
	Net        string           `json:"net"` //livenet/signet/testnet
	Tick       string           `json:"tick"`
	Address    string           `json:"address"`
	OrderState model.OrderState `json:"orderState"` //1-create,2-finish,3-cancel,5-timeout,6-err,100-all
	OrderType  model.OrderType  `json:"orderType"`  //1-sell,2-buy
	Flag       int64            `json:"flag"`
	Page       int64            `json:"page"`
	Limit      int64            `json:"limit"`
	SortKey    string           `json:"sortKey"`  //coinRatePrice/timestamp
	SortType   int64            `json:"sortType"` //1/-1
}
