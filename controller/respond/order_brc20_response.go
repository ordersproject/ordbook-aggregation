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
	SellerAddress  string           `json:"sellerAddress,omitempty"`  //Seller's address
	BuyerAddress   string           `json:"buyerAddress,omitempty"`   //Buyer's address
	PsbtRaw        string           `json:"psbtRaw,omitempty"`        //PSBT Raw
	Timestamp      int64            `json:"timestamp,omitempty"`      //Create time
}

type Brc20TickInfoResponse struct {
	Total   int64             `json:"total,omitempty"`
	Results []*Brc20TickItem `json:"results,omitempty"`
	Flag    int64             `json:"flag,omitempty"`
}

type Brc20TickItem struct {
	Net                string `json:"net,omitempty"`                //Net env
	Tick               string `json:"tick,omitempty"`               //tick
	Pair               string `json:"pair,omitempty"`               //pair for trade
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
	OrderId string `json:"orderId,omitempty"` //
	PsbtRaw string `json:"psbtRaw,omitempty"` //PSBT Raw
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

type DoBidResp struct {
	TxIdX string `json:"txIdX"`
	TxIdY string `json:"txIdY"`
}


type Brc20BidDummyResponse struct {
	Total   int64             `json:"total,omitempty"`
	Results []*DummyItem `json:"results,omitempty"`
	Flag    int64             `json:"flag,omitempty"`
}

type DummyItem struct {
	Order     string `json:"order"`
	DummyId   string `json:"dummyId"`
	Timestamp int64  `json:"timestamp"`
}
