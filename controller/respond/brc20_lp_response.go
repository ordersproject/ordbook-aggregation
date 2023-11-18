package respond

type Brc20LpAddBatchStep1Resp struct {
	Fees              int64    `json:"fees"`
	CommitTxHash      string   `json:"commitTxHash"`
	RevealTxHashList  []string `json:"revealTxHashList"`
	InscriptionIdList []string `json:"inscriptionIdList"`
	LpOrderIdList     []string `json:"lpOrderIdList"`
}

type Brc20LpAddStep2Resp struct {
	Fees          int64  `json:"fees"`
	TxId          string `json:"txId"`
	CoinPrice     int64  `json:"coinPrice"`
	LpOrderId     string `json:"lpOrderId"`
	BtcUtxoId     string `json:"btcUtxoId"`
	BtcAmount     int64  `json:"btcAmount"`
	CoinRatePrice uint64 `json:"coinRatePrice"`
	Ratio         int64  `json:"ratio"`
}

type Brc20LpAddStep2BatchResp struct {
	Fees int64                           `json:"fees"`
	TxId string                          `json:"txId"`
	List []*Brc20LpAddStep2BatchItemResp `json:"list"`
}

type Brc20LpAddStep2BatchItemResp struct {
	CoinPrice     uint64 `json:"coinPrice"`
	LpOrderId     string `json:"lpOrderId"`
	BtcUtxoId     string `json:"btcUtxoId"`
	BtcAmount     int64  `json:"btcAmount"`
	CoinRatePrice uint64 `json:"coinRatePrice"`
	Ratio         int64  `json:"ratio"`
}
