package respond

type Brc20ClaimItem struct {
	Net           string `json:"net,omitempty"`           //Net env
	OrderId       string `json:"orderId,omitempty"`       //Order ID
	Tick          string `json:"tick,omitempty"`          //Brc20 symbol
	Fee           uint64 `json:"fee,omitempty"`           //claim fee
	CoinAmount    uint64 `json:"coinAmount,omitempty"`    //Brc20 amount
	InscriptionId string `json:"inscriptionId,omitempty"` //InscriptionId
	PsbtRaw       string `json:"psbtRaw,omitempty"`       //PSBT Raw
}
