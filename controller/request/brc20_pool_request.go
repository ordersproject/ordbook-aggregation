package request

import "ordbook-aggregation/model"

type PoolBrc20FetchReq struct {
	Net       string          `json:"net"` //livenet/signet/testnet
	Tick      string          `json:"tick"`
	Pair      string          `json:"pair"`
	PoolType  model.PoolType  `json:"poolType"`  //1-tick,2-btc,3-both,100-all
	PoolState model.PoolState `json:"poolState"` //1-add,2-remove,3-used,4-claim
	Limit     int64           `json:"limit"`
	Flag      int64           `json:"flag"`
	Page      int64           `json:"page"`
	Address   string          `json:"address"`
	SortKey   string          `json:"sortKey"`  //timestamp
	SortType  int64           `json:"sortType"` //1/-1
}

type PoolPairFetchOneReq struct {
	Net     string `json:"net"` //livenet/signet/testnet
	Tick    string `json:"tick"`
	Pair    string `json:"pair"`
	Address string `json:"address"`
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
	PsbtRaw     string          `json:"psbtRaw"`     // BTC
	BtcPoolMode model.PoolMode  `json:"btcPoolMode"` //1-psbt,2-custody, default:custody
	BtcUtxoId   string          `json:"btcUtxoId"`   //txId_index
	Amount      uint64          `json:"amount"`
	Ratio       int64           `json:"ratio"` // ratio: 12/15/18/100//10000
}

type OrderPoolBrc20UpdateReq struct {
	Net       string          `json:"net"` //livenet/signet/testnet
	OrderId   string          `json:"orderId"`
	PoolState model.PoolState `json:"poolState"` //1-add,2-remove,3-used,4-claim
}

type PoolBrc20FetchInscriptionReq struct {
	Net     string `json:"net"` //livenet/signet/testnet
	Tick    string `json:"tick"`
	Address string `json:"address"`
}

type PoolBrc20ClaimReq struct {
	Net          string `json:"net"` //livenet/signet/testnet
	Tick         string `json:"tick"`
	Address      string `json:"address"`
	PreSigScript string `json:"preSigScript"`
	PoolOrderId  string `json:"poolOrderId"`
}

type PoolBrc20ClaimUpdateReq struct {
	PsbtRaw     string `json:"psbtRaw"`
	PoolOrderId string `json:"poolOrderId"`
	RewardIndex int64  `json:"rewardIndex"` //0-no, 1-yes
}

type PoolBrc20RewardReq struct {
	Net     string `json:"net"` //livenet/signet/testnet
	Tick    string `json:"tick"`
	Address string `json:"address"`
}

type PoolBrc20ClaimRewardReq struct {
	Net          string `json:"net"` //livenet/signet/testnet
	Tick         string `json:"tick"`
	Address      string `json:"address"`
	RewardAmount int64  `json:"rewardAmount"`
}

type PoolRewardOrderFetchReq struct {
	Net         string            `json:"net"` //livenet/signet/testnet
	Tick        string            `json:"tick"`
	Pair        string            `json:"pair"`
	RewardState model.RewardState `json:"rewardState"` //1-create,2-inscription,3-send,100-all
	Limit       int64             `json:"limit"`
	Flag        int64             `json:"flag"`
	Page        int64             `json:"page"`
	Address     string            `json:"address"`
	SortKey     string            `json:"sortKey"`  //timestamp
	SortType    int64             `json:"sortType"` //1/-1
}
