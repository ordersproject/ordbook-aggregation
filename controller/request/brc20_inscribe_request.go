package request

type Brc20PreReq struct {
	Net            string `json:"net"`            //mainnet/signet/testnet
	ReceiveAddress string `json:"receiveAddress"` //Address which user receive ordinals
	Content        string `json:"content"`        //
}

type Brc20CommitReq struct {
	Net        string `json:"net"`        //mainnet/signet/testnet
	FeeAddress string `json:"feeAddress"` //platform fee address
}