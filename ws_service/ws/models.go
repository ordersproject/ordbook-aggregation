package ws

import (
	"github.com/tidwall/gjson"
	"ordbook-aggregation/model"
	"ordbook-aggregation/tool"
	"strings"
)

type WsData struct {
	M string      `json:"M"`           //method
	C interface{} `json:"C"`           //code
	D interface{} `json:"D,omitempty"` //data
}

func WsDataFromStringMsg(msg string) *WsData  {
	ws := &WsData{}
	msg = strings.Trim(msg, "\"")
	if !gjson.Valid(msg) {
		msg = strings.ReplaceAll(msg, "\\", "")
		if !gjson.Valid(msg) {
			return nil
		}
	}
	ws.M = gjson.Get(msg, "M").String()
	ws.C = gjson.Get(msg, "C").Int()
	ws.D = gjson.Get(msg, "D").String()
	return ws
}

func (w *WsData) ToString() (string, error)  {
	return tool.ObjectToJson(w)
}

type WsNotifyTick struct {
	Net                string  `json:"net"`
	Tick               string  `json:"tick"`
	Pair               string  `json:"pair"`               //
	Buy                uint64  `json:"buy"`                //
	Sell               uint64  `json:"sell"`               //
	Low                uint64  `json:"low"`                //
	High               uint64  `json:"high"`               //
	Open               uint64  `json:"open"`               //
	Last               uint64  `json:"last"`               //
	Volume             uint64  `json:"volume"`             //
	Amount             uint64  `json:"amount"`             //
	Vol                uint64  `json:"vol"`                //
	AvgPrice           uint64  `json:"avgPrice"`           //
	QuoteSymbol        string  `json:"quoteSymbol"`        //
	PriceChangePercent float64 `json:"priceChangePercent"` //
	Ut                 int64   `json:"ut"`
}

func NewWsNotifyTick(data *model.Brc20TickModel) *WsNotifyTick {
	return &WsNotifyTick{
		Net:                data.Net,
		Tick:               data.Tick,
		Pair:               data.Pair,
		Buy:                data.Buy,
		Sell:               data.Sell,
		Low:                data.Low,
		High:               data.High,
		Open:               data.Open,
		Last:               data.Last,
		Volume:             data.Volume,
		Amount:             data.Amount,
		Vol:                data.Vol,
		AvgPrice:           data.AvgPrice,
		QuoteSymbol:        data.QuoteSymbol,
		PriceChangePercent: data.PriceChangePercent,
		Ut:                 data.UpdateTime,
	}
}

type WsConnect struct {
	Uuid      string `json:"uuid"`
}