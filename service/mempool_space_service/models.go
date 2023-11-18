package mempool_space_service

type TxDetail struct {
	TxId          string        `json:"txid"`
	Height        string        `json:"height"`
	OutputDetails []*OutputItem `json:"outputDetails"`
}

type OutputItem struct {
	OutputHash string `json:"outputHash"`
	Tag        string `json:"tag"`
	Amount     string `json:"amount"`
}
