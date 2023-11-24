package common_service

import (
	"github.com/btcsuite/btcd/chaincfg"
	"ordbook-aggregation/service/unisat_service"
	"strings"
)

type PriceInfo struct {
	VisionPrice string `json:"visionPrice"`
	LastPrice   string `json:"lastPrice"`
	High24h     string `json:"high24h"`
	Low24h      string `json:"low24h"`
	UpdateTime  int64  `json:"updateTime"`
}

var (
	Brc20TickInscriptionMap map[string]string = map[string]string{
		"rdex": "79ddd895e48f5f250cfed1e9656e0eb9416d49711c765990bee20f891be1e386i0",
		"ordi": "b61b0172d95e266c18aea0c624db987e971a5d6d4ebc2aaed85da4642d635735i0",
		"meme": "307ffac5d20fc188f723706f85d75c926550d536f5fd1113839586f38542971ci0",
		"pepe": "54d5fe82f5d284363fec6ae6137d0e5263e237caf15211078252c0d95af8943ai0",
		"sats": "9b664bdd6f5ed80d8d88957b63364c41f3ad4efb8eee11366aa16435974d9333i0",
		"orxc": "2cdc14fe7c33a181df8ffbc4915cf72c1c1c886871d773aff9ada79fa09b5456i0",
		"btcs": "edc052335f914ee47a758cff988494fbb569d820e66ac8581008e44b26dcdb43i0",
		"oxbt": "c0e650f33432b627ac0346e9cbdfd30f2b8590c16236c42cd45498b0f27f5c4ei0",
		"vmpx": "beafe671f13b86300454d787d31e2918442d396225098a9c12ae4bf4d077196fi0",
		"trac": "b006d8e232bdd01e656c40bdbec83bb38413a8af3a58570551940d8f23d4b85ai0",
		"ibtc": "a56c773fcdb4098fe8f0f76edf6df64b162a93a270178eb27d762c37f348989bi0",
		"bili": "f07b84ea3cad6a580aa7f613bea38d9237d49ee5f1b10dbf83b5d8c6b27a06b3i0",
		"cats": "4923d5b5f469d63a8cdb27f95361a250f34d1540e525fbb796a836e9b3094d09i0",
		"fish": "82e39b612d0d00fe0c18ac2bbbe4fd23beb76f83f5a76950055114810099a804i0",
		"sayc": "85cb878918b5f7f52547afa5cc2cd32110565e513d423f8fce3a854b5b1ffcf4i0",
		"rats": "77df24c9f1bd1c6a606eb12eeae3e2a2db40774d54b839b5ae11f438353ddf47i0",
	}

	Brc20TickMarketDataMap map[string]*PriceInfo = map[string]*PriceInfo{
		"rdex": &PriceInfo{},
		"ordi": &PriceInfo{},
		"meme": &PriceInfo{},
		"pepe": &PriceInfo{},
		"sats": &PriceInfo{},
		"orxc": &PriceInfo{},
		"btcs": &PriceInfo{},
		"oxbt": &PriceInfo{},
		"vmpx": &PriceInfo{},
		"trac": &PriceInfo{},
		"ibtc": &PriceInfo{},
		"bili": &PriceInfo{},
		"cats": &PriceInfo{},
		"fish": &PriceInfo{},
		"sayc": &PriceInfo{},
		"rats": &PriceInfo{},
	}

	UpdateMarketTime int64 = 0
)

func ChangeRealTick(tick string) string {
	switch tick {
	case "rdex", "oxbt", "grum", "vmpx", "lger", "sayc", "orxc":
		tick = strings.ToUpper(tick)
	}
	return tick
}

func GetNetParams(net string) *chaincfg.Params {
	var (
		netParams *chaincfg.Params = &chaincfg.MainNetParams
	)
	switch strings.ToLower(net) {
	case "mainnet", "livenet":
		netParams = &chaincfg.MainNetParams
		break
	case "signet":
		netParams = &chaincfg.SigNetParams
		break
	case "testnet":
		netParams = &chaincfg.TestNet3Params
		break
	}
	return netParams
}

func GetFeeSummary() int64 {
	var (
		currentFee int64 = 0
		feeSummary *unisat_service.FeeSummary
	)
	feeSummary, _ = unisat_service.GetFeeDetail()
	if feeSummary != nil {
		for _, v := range feeSummary.List {
			if strings.ToLower(v.Title) == "avg" {
				currentFee = int64(v.FeeRate)
				break
			}
		}
	}
	return currentFee
}
