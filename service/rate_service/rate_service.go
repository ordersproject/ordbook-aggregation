package rate_service

import (
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/service/oklink_service"
)

func FetchRate() (*respond.RateResp, error) {
	var (
		usd map[string]string = make(map[string]string)
	)
	btcPriceInfo, _ := oklink_service.GetBrc20TickMarketData("")
	if btcPriceInfo != nil && len(btcPriceInfo) != 0 {
		usd["btc"] = btcPriceInfo[0].LastPrice
	}

	return &respond.RateResp{
		USD: usd,
	}, nil
}
