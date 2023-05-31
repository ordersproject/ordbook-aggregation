package order_brc20_service

import (
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
)

func FetchTickers(req *request.TickBrc20FetchReq) (*respond.Brc20TickInfoResponse, error){
	var (
		list []*respond.Brc20TickItem = make([]*respond.Brc20TickItem, 0)
	)

	list = []*respond.Brc20TickItem{
		&respond.Brc20TickItem{
			Net:                "mainnet",
			Pair:               "ORDI-BTC",
			Tick:               "ordi",
			Buy:                "0.00005470",
			//Sell:               "",
			//Low:                "",
			//High:               "",
			//Open:               "",
			//Last:               "",
			//Volume:             "",
			//Amount:             "",
			//Vol:                "",
			//AvgPrice:           "",
			QuoteSymbol:"+",
			PriceChangePercent: "4.32",
			Ut:                 0,
		},
		&respond.Brc20TickItem{
			Net:                "mainnet",
			Pair:               "PEPE-BTC",
			Tick:               "pepe",
			Buy:                "0.00006853",
			QuoteSymbol:"+",
			PriceChangePercent: "1.43",
			Ut:                 0,
		},
		&respond.Brc20TickItem{
			Net:                "mainnet",
			Pair:               "MEME-BTC",
			Tick:               "meme",
			Buy:                "0.00325100",
			QuoteSymbol:"-",
			PriceChangePercent: "11.46",
			Ut:                 0,
		},
		&respond.Brc20TickItem{
			Net:                "mainnet",
			Pair:               "OMNI-BTC",
			Tick:               "omni",
			Buy:                "0.00004056",
			QuoteSymbol:"+",
			PriceChangePercent: "0.21",
			Ut:                 0,
		},
		&respond.Brc20TickItem{
			Net:                "mainnet",
			Pair:               "SATS-BTC",
			Tick:               "sats",
			Buy:                "0.00001300",
			QuoteSymbol:"-",
			PriceChangePercent: "0.35",
			Ut:                 0,
		},
	}
	return &respond.Brc20TickInfoResponse{
		Total:   5,
		Results: list,
		Flag:    0,
	}, nil
}