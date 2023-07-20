package request

import "ordbook-aggregation/model"

type PoolBrc20FetchReq struct {
	Net       string          `json:"net"` //livenet/signet/testnet
	Tick      string          `json:"tick"`
	Pair      string          `json:"pair"`
	PoolType  model.PoolType  `json:"poolType"`  //1-tick,2-btc
	PoolState model.PoolState `json:"poolState"` //1-add,2-remove,3-used,4-claim
	Limit     int64           `json:"limit"`
	Flag      int64           `json:"flag"`
	Page      int64           `json:"page"`
	Address   string          `json:"address"`
	SortKey   string          `json:"sortKey"`  //timestamp
	SortType  int64           `json:"sortType"` //1/-1
}

type PoolPairFetchOneReq struct {
	Net  string `json:"net"` //livenet/signet/testnet
	Tick string `json:"tick"`
	Pair string `json:"pair"`
}

type PoolBrc20FetchOneReq struct {
	Net     string `json:"net"` //livenet/signet/testnet
	Tick    string `json:"tick"`
	OrderId string `json:"orderId"`
	Address string `json:"address"`
}

type PoolBrc20PushReq struct {
	Net         string          `json:"net"` //livenet/signet/testnet
	Tick        string          `json:"tick"`
	Pair        string          `json:"pair"`
	PoolType    model.PoolType  `json:"poolType"`  //1-tick,2-btc,3-both
	PoolState   model.PoolState `json:"poolState"` //1-add,2-remove,3-used,4-claim
	Address     string          `json:"address"`
	CoinPsbtRaw string          `json:"coinPsbtRaw"`
	CoinAmount  uint64          `json:"coinAmount"`
	PsbtRaw     string          `json:"psbtRaw"`
	Amount      uint64          `json:"amount"`
}
