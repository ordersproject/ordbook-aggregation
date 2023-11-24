package unisat_service

type BroadcastTxResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}
type FeeSummary struct {
	List []struct {
		Title   string `json:"title"`
		Desc    string `json:"desc"`
		FeeRate int    `json:"feeRate"`
	} `json:"list"`
}
