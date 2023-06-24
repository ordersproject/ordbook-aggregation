package request

import "ordbook-aggregation/model"

type ColdDownUtxo struct {
	Net           string         `json:"net"`
	TxId          string         `json:"txId"`
	Index         int64          `json:"index"`
	Amount        uint64         `json:"amount"`
	PkScript      string         `json:"pkScript"`
	PreTxHex      string         `json:"preTxHex"`
	Address       string         `json:"address"`
	PriKeyHex     string         `json:"priKeyHex"`
	PerAmount     uint64         `json:"perAmount"`
	Count         int64          `json:"count"`
	UtxoType      model.UtxoType `json:"utxoType"`
	ChangeAddress string         `json:"changeAddress"`
	FeeRate       int64          `json:"feeRate"`
}

type CollectionUtxo struct {
	UtxoList  []*UtxoItem `json:"utxoList"`
	Address   string  `json:"address"`
	PriKeyHex string  `json:"priKeyHex"`
	Net       string  `json:"net"`
	FeeRate   int64   `json:"feeRate"`
}

type UtxoItem struct {
	TxId     string `json:"txId"`
	Index    int64  `json:"index"`
	Amount   uint64 `json:"amount"`
	PkScript string `json:"pkScript"`
}

type ColdDownBrcTransfer struct {
	Net                    string `json:"net"`
	Tick                    string `json:"tick"`
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
}

type ColdDownBrcTransferBatch struct {
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
}