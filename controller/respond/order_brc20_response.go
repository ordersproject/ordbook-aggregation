package respond

import "ordbook-aggregation/model"

type OrderResponse struct {
	Total   int64        `json:"total,omitempty"`
	Results []*Brc20Item `json:"results,omitempty"`
	Flag    int64        `json:"flag,omitempty"`
}

type Brc20Item struct {
	Net            string           `json:"net,omitempty"`            //网络环境
	OrderId           string           `json:"orderId,omitempty"`           //订单ID
	Tick           string           `json:"tick,omitempty"`           //brc20代币symbol
	Amount         uint64           `json:"amount,omitempty"`         //btc买卖值，按最小单位，sat
	DecimalNum     int              `json:"decimalNum,omitempty"`     //btc小数位数
	CoinAmount     uint64           `json:"coinAmount,omitempty"`     //brc20代币买卖值，没有最小单位，没有小数
	CoinDecimalNum int              `json:"coinDecimalNum,omitempty"` //brc20代币暂时忽略
	CoinRatePrice  uint64           `json:"coinRatePrice,omitempty"`  //brc20代币对应btc的汇率
	OrderState     model.OrderState `json:"orderState,omitempty"`     //订单状态：1-create,2-finish,3-cancel
	OrderType      model.OrderType  `json:"orderType,omitempty"`      //订单类型：1-sell,2-buy
	SellerAddress  string           `json:"sellerAddress,omitempty"`  //出售地址
	BuyerAddress   string           `json:"buyerAddress,omitempty"`   //购买地址
	PsbtRaw        string           `json:"psbtRaw,omitempty"`        //PSBT生交易
	Timestamp      int64            `json:"timestamp,omitempty"`      //创建时间
}

type Brc20TickInfoResponse struct {
	Total   int64             `json:"total,omitempty"`
	Results []*Brc20TickItem `json:"results,omitempty"`
	Flag    int64             `json:"flag,omitempty"`
}

type Brc20TickItem struct {
	Net                string `json:"net,omitempty"`                //网络环境
	Tick               string `json:"tick,omitempty"`               //tick
	Pair               string `json:"pair,omitempty"`               //交易对
	Buy                string `json:"buy,omitempty"`                //最新成交价
	Sell               string `json:"sell,omitempty"`               //
	Low                string `json:"low,omitempty"`                //
	High               string `json:"high,omitempty"`               //
	Open               string `json:"open,omitempty"`               //
	Last               string `json:"last,omitempty"`               //
	Volume             string `json:"volume,omitempty"`             //
	Amount             string `json:"amount,omitempty"`             //
	Vol                string `json:"vol,omitempty"`                //
	AvgPrice           string `json:"avgPrice,omitempty"`           //
	QuoteSymbol        string `json:"quoteSymbol,omitempty"`        //涨跌符号：+/-
	PriceChangePercent string `json:"priceChangePercent,omitempty"` //变动百分比：0.11表示0.11%
	Ut                 int64  `json:"at,omitempty"`                 //updateTime
}

type KlineItem struct {
	Net           string           `json:"net,omitempty"`
	Data0 string `json:"0"`
	Data1 string `json:"1"`
	Data2 string `json:"2"`
	Data3 string `json:"3"`
	Data4 string `json:"4"`
	Data5 string `json:"5"`
	Data6 string `json:"6"`
	Data7 string `json:"7"`
	Data8 string `json:"8"`
	Data9 string `json:"9"`
}

type BidPsbt struct {
	Net     string `json:"net,omitempty"`
	Tick    string `json:"tick,omitempty"`    //
	PsbtRaw string `json:"psbtRaw,omitempty"` //PSBT生交易
}

type BidPre struct {
	Net           string           `json:"net,omitempty"`
	Tick          string           `json:"tick,omitempty"` //
	AvailableList []*AvailableItem `json:"availableList,omitempty"`
}

type AvailableItem struct {
	InscriptionId     string `json:"inscriptionId,omitempty"`
	InscriptionNumber string `json:"inscriptionNumber,omitempty"`
	CoinAmount        string `json:"coinAmount,omitempty"`
}