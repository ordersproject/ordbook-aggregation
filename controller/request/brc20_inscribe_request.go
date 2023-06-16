package request

type Brc20PreReq struct {
	Net            string `json:"net"`            //livenet/signet/testnet
	ReceiveAddress string `json:"receiveAddress"` //Address which user receive ordinals
	Content        string `json:"content"`        //
	FeeRate        int64 `json:"feeRate"`        //
}

type Brc20CommitReq struct {
	Net        string         `json:"net"`        //livenet/signet/testnet
	FeeAddress string         `json:"feeAddress"` //platform fee address
	Utxos      []*ReceiveUtxo `json:"utxos"`//utxo list
}

type ReceiveUtxo struct {
	OutTx     string `json:"outTx"`
	OutIndex  int64  `json:"outIndex"`
	OutAmount int64  `json:"outAmount"`
}