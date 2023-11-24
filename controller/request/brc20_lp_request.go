package request

type LpAddOneStep1Request struct {
	Net                    string `json:"net"`
	Tick                   string `json:"tick"`
	TxId                   string `json:"txId"`
	Index                  int64  `json:"index"`
	Amount                 uint64 `json:"amount"`
	PkScript               string `json:"pkScript"`
	PreTxHex               string `json:"preTxHex"`
	Address                string `json:"address"`
	PriKeyHex              string `json:"priKeyHex"`
	InscribeTransferAmount int64  `json:"inscribeTransferAmount"`
	ChangeAddress          string `json:"changeAddress"`
	FeeRate                int64  `json:"feeRate"`
	Count                  int64  `json:"count"`
	IsOnlyCal              bool   `json:"isOnlyCal"`
	OutAddressType         string `json:"outAddressType"`
	Brc20InValue           int64  `json:"brc20InValue"`
}

type LpAddOneStep2Request struct {
	LpOrderId     string `json:"lpOrderId"`
	TxId          string `json:"txId"`
	Index         int64  `json:"index"`
	Amount        uint64 `json:"amount"`
	PkScript      string `json:"pkScript"`
	Address       string `json:"address"`
	PriKeyHex     string `json:"priKeyHex"`
	ChangeAddress string `json:"changeAddress"`
	FeeRate       int64  `json:"feeRate"`
	Ratio         int64  `json:"ratio"`
	BtcOutValue   int64  `json:"btcOutValue"`
}

type LpAddOneStep2BatchRequest struct {
	LpOrderIdList []string `json:"lpOrderIdList"`
	Net           string   `json:"net"`
	TxId          string   `json:"txId"`
	Index         int64    `json:"index"`
	Amount        uint64   `json:"amount"`
	PkScript      string   `json:"pkScript"`
	Address       string   `json:"address"`
	PriKeyHex     string   `json:"priKeyHex"`
	ChangeAddress string   `json:"changeAddress"`
	FeeRate       int64    `json:"feeRate"`
	Ratio         int64    `json:"ratio"`
	BtcOutValue   int64    `json:"btcOutValue"`
}

type LpCancelOneBatchRequest struct {
	LpOrderIdList []string `json:"lpOrderIdList"`
	Net           string   `json:"net"`
	Address       string   `json:"address"`
	FeeRate       int64    `json:"feeRate"`
	IsCalOnly     bool     `json:"isCalOnly"`
}
