package request

type OrderBrc20ClaimFetchOneReq struct {
	Net     string `json:"net"` //livenet/signet/testnet
	Tick    string `json:"tick"`
	Address string `json:"address"`
}

type OrderBrc20ClaimUpdateReq struct {
	Net     string `json:"net"` //livenet/signet/testnet
	OrderId string `json:"orderId"`
	PsbtRaw string `json:"psbtRaw"`
	Address string `json:"address"`
}
