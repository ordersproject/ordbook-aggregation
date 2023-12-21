package own_service

type OwnServiceResp struct {
	Code int64       `json:"code"`
	Data interface{} `json:"data"`
}

type UtxoInfo struct {
	IsExist     bool   `json:"isExist"`
	TxConfirm   bool   `json:"txConfirm"`
	SpendStatus string `json:"spendStatus"`
	Height      int64  `json:"height"`
	Date        int64  `json:"date"`
	Value       int64  `json:"value"`
	Where       string `json:"where"`
	SpendInfo   struct {
		SpendTx string `json:"spendTx"`
		Height  int64  `json:"height"`
		Date    int64  `json:"date"`
		Where   string `json:"where"`
	} `json:"spendInfo"`
}

type TokenInfoResp struct {
	List  []*TokenInfo `json:"list"`
	Total int64        `json:"total"`
}

type TokenInfo struct {
	AvailableBalance    string `json:"availableBalance"`
	Decimal             string `json:"decimal"`
	OverallBalance      string `json:"overallBalance"`
	Ticker              string `json:"ticker"`
	TransferableBalance string `json:"transferableBalance"`
}
