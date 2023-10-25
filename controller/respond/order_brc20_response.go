package respond

import "ordbook-aggregation/model"

type OrderResponse struct {
	Total   int64        `json:"total,omitempty"`
	Results []*Brc20Item `json:"results,omitempty"`
	Flag    int64        `json:"flag,omitempty"`
}

type Brc20Item struct {
	Net            string           `json:"net,omitempty"`            //Net env
	OrderId        string           `json:"orderId,omitempty"`        //Order ID
	Tick           string           `json:"tick,omitempty"`           //Brc20 symbol
	Amount         uint64           `json:"amount,omitempty"`         //Btc: sat
	DecimalNum     int              `json:"decimalNum,omitempty"`     //Btc decimal
	CoinAmount     uint64           `json:"coinAmount,omitempty"`     //Brc20 amount
	CoinDecimalNum int              `json:"coinDecimalNum,omitempty"` //omitempty
	CoinRatePrice  uint64           `json:"coinRatePrice,omitempty"`  //Rate for brc20-btc
	OrderState     model.OrderState `json:"orderState,omitempty"`     //Order state：1-create,2-finish,3-cancel
	OrderType      model.OrderType  `json:"orderType,omitempty"`      //Order type：1-sell,2-buy
	FreeState      model.FreeState  `json:"freeState,omitempty"`      //1-for free
	SellerAddress  string           `json:"sellerAddress,omitempty"`  //Seller's address
	BuyerAddress   string           `json:"buyerAddress,omitempty"`   //Buyer's address
	InscriptionId  string           `json:"inscriptionId,omitempty"`  //InscriptionId
	PsbtRaw        string           `json:"psbtRaw,omitempty"`        //PSBT Raw
	Timestamp      int64            `json:"timestamp"`                //Create time
}

type Brc20TickInfoResponse struct {
	Total   int64            `json:"total,omitempty"`
	Results []*Brc20TickItem `json:"results,omitempty"`
	Flag    int64            `json:"flag,omitempty"`
}

type Brc20TickItem struct {
	Net                string `json:"net,omitempty"`                //Net env
	Tick               string `json:"tick,omitempty"`               //tick
	Pair               string `json:"pair,omitempty"`               //pair for trade
	Icon               string `json:"icon,omitempty"`               //icon
	Buy                string `json:"buy,omitempty"`                //
	Sell               string `json:"sell,omitempty"`               //
	Low                string `json:"low,omitempty"`                //
	High               string `json:"high,omitempty"`               //
	Open               string `json:"open,omitempty"`               //
	Last               string `json:"last,omitempty"`               //
	Volume             string `json:"volume,omitempty"`             //
	Amount             string `json:"amount,omitempty"`             //
	Vol                string `json:"vol,omitempty"`                //
	AvgPrice           string `json:"avgPrice,omitempty"`           //
	QuoteSymbol        string `json:"quoteSymbol,omitempty"`        //+/-
	PriceChangePercent string `json:"priceChangePercent,omitempty"` //0.11 mean 0.11%
	Ut                 int64  `json:"at,omitempty"`                 //updateTime
}

type Brc20KlineInfo struct {
	Net      string       `json:"net,omitempty"`
	Tick     string       `json:"tick"`
	Interval string       `json:"interval"` //1m/1s/15m/1h/4h/1d/1w/
	List     []*KlineItem `json:"list"`
	Flag     int64        `json:"flag"`
}
type KlineItem struct {
	Timestamp int64  `json:"timestamp"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Close     string `json:"close"`
	Volume    int64  `json:"volume"`
}

type BidPsbt struct {
	Net     string `json:"net,omitempty"`
	Tick    string `json:"tick,omitempty"`    //
	OrderId string `json:"orderId,omitempty"` //
	PsbtRaw string `json:"psbtRaw,omitempty"` //PSBT Raw
}

type BidPre struct {
	Net           string           `json:"net,omitempty"`
	Tick          string           `json:"tick,omitempty"` //
	AvailableList []*AvailableItem `json:"availableList,omitempty"`
	Total         int64            `json:"total,omitempty"`
}

type AvailableItem struct {
	InscriptionId     string         `json:"inscriptionId,omitempty"`
	InscriptionNumber string         `json:"inscriptionNumber,omitempty"`
	CoinAmount        string         `json:"coinAmount,omitempty"`
	PoolOrderId       string         `json:"poolOrderId,omitempty"`
	CoinRatePrice     uint64         `json:"coinRatePrice,omitempty"`
	PoolType          model.PoolType `json:"poolType,omitempty"`
	BtcPoolMode       model.PoolMode `json:"btcPoolMode,omitempty"` //PoolMode for btc
	BidCount          int64          `json:"bidCount"`
}

type WsUuidResp struct {
	Uuid string `json:"uuid"`
}

type CheckBrc20InscriptionReq struct {
	InscriptionId          string `json:"inscriptionId"` //
	InscriptionNumber      string `json:"inscriptionNumber"`
	Location               string `json:"location"`         //location - txid:vout:offset
	InscriptionState       string `json:"inscriptionState"` //inscribe state: success/fail
	Token                  string `json:"token"`            //tick name
	TokenType              string `json:"tokenType"`        //token type
	ActionType             string `json:"actionType"`       //
	OwnerAddress           string `json:"ownerAddress"`
	BlockHeight            string `json:"blockHeight"`
	TxId                   string `json:"txId"`
	AvailableTransferState string `json:"availableTransferState"` //Available Transfer state: success/fail
	Amount                 string `json:"amount"`                 //
}

type BalanceDetails struct {
	Page                string         `json:"page"`
	Limit               string         `json:"limit"`
	TotalPage           string         `json:"totalPage"`
	Token               string         `json:"token"`
	TokenType           string         `json:"tokenType"`
	Balance             string         `json:"balance"`
	AvailableBalance    string         `json:"availableBalance"`
	TransferBalance     string         `json:"transferBalance"`
	TransferBalanceList []*BalanceItem `json:"transferBalanceList"`
}

type BalanceItem struct {
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
	Amount            string `json:"amount"`
}

type Brc20BalanceList struct {
	Page        string             `json:"page"`
	Limit       string             `json:"limit"`
	TotalPage   string             `json:"totalPage"`
	BalanceList []*BalanceListItem `json:"balanceList"`
}

type BalanceListItem struct {
	Token            string `json:"token"`
	TokenType        string `json:"tokenType"`
	Balance          string `json:"balance"`
	AvailableBalance string `json:"availableBalance"`
	TransferBalance  string `json:"transferBalance"`
}

type DoBidResp struct {
	TxIdX string `json:"txIdX"`
	TxIdY string `json:"txIdY"`
}

type Brc20BidDummyResponse struct {
	Total   int64        `json:"total,omitempty"`
	Results []*DummyItem `json:"results,omitempty"`
	Flag    int64        `json:"flag,omitempty"`
}

type DummyItem struct {
	Order     string `json:"order"`
	DummyId   string `json:"dummyId"`
	Timestamp int64  `json:"timestamp"`
}
