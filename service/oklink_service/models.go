package oklink_service

type OklinkResp struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type OklinkBrc20BalanceDetails struct {
	Page                string         `json:"page"`
	Limit               string         `json:"limit"`
	TotalPage           string         `json:"totalPage"`
	Token               string         `json:"token"`
	TokenType           string         `json:"tokenType"`
	Balance             string         `json:"balance"`
	AvailableBalance    string         `json:"availableBalance"`
	TransferBalance     string         `json:"transferBalance"`
	TransferBalanceList []*BalanceItem `json:"transferBalanceList"`
}

type BalanceItem struct {
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
	Amount            string `json:"amount"`
}

type OklinkBrc20BalanceList struct {
	Page        string             `json:"page"`
	Limit       string             `json:"limit"`
	TotalPage   string             `json:"totalPage"`
	BalanceList []*BalanceListItem `json:"balanceList"`
}

type BalanceListItem struct {
	Token            string `json:"token"`
	TokenType        string `json:"tokenType"`
	Balance          string `json:"balance"`
	AvailableBalance string `json:"availableBalance"`
	TransferBalance  string `json:"transferBalance"`
}

type OklinkInscriptionDetails struct {
	Page             string             `json:"page"`
	Limit            string             `json:"limit"`
	TotalPage        string             `json:"totalPage"`
	TotalInscription string             `json:"totalInscription"`
	InscriptionsList []*InscriptionItem `json:"inscriptionsList"`
}

type InscriptionItem struct {
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
	Location          string `json:"location"`
	Token             string `json:"token"`
	State             string `json:"state"`
	Msg               string `json:"msg"`
	TokenType         string `json:"tokenType"`
	ActionType        string `json:"actionType"`
	LogoUrl           string `json:"logoUrl"`
	OwnerAddress      string `json:"ownerAddress"`
	TxId              string `json:"txId"`
	BlockHeight       string `json:"blockHeight"`
	ContentSize       string `json:"contentSize"`
	Time              string `json:"time"`
}

type TxDetail struct {
	TxId          string        `json:"txid"`
	OutputDetails []*OutputItem `json:"outputDetails"`
}

type OutputItem struct {
	OutputHash string `json:"outputHash"`
	Tag        string `json:"tag"`
	Amount     string `json:"amount"`
}

type BroadcastTxResp struct {
	ChainFullName  string `json:"chainFullName"`
	ChainShortName string `json:"chainShortName"`
	TxId           string `json:"txid"`
}

type OklinkUtxoDetails struct {
	Page      string      `json:"page"`
	Limit     string      `json:"limit"`
	TotalPage string      `json:"totalPage"`
	UtxoList  []*UtxoItem `json:"utxoList"`
}

type UtxoItem struct {
	TxId          string `json:"txid"`
	Index         string `json:"index"`
	Height        string `json:"height"`
	BlockTime     string `json:"blockTime"`
	Address       string `json:"address"`
	UnspentAmount string `json:"unspentAmount"`
}
