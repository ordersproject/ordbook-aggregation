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
	Height        string        `json:"height"`
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

type OklinkBrc20HolderAddressList struct {
	Page         string               `json:"page"`
	Limit        string               `json:"limit"`
	TotalPage    string               `json:"totalPage"`
	PositionList []*HolderAddressItem `json:"positionList"`
}

type HolderAddressItem struct {
	HolderAddress string `json:"holderAddress"`
	Amount        string `json:"amount"`
	Rank          string `json:"rank"`
}

type OklinkBrc20transactionList struct {
	Page             string                  `json:"page"`
	Limit            string                  `json:"limit"`
	TotalPage        string                  `json:"totalPage"`
	InscriptionsList []*Brc20transactionItem `json:"inscriptionsList"`
}

type Brc20transactionItem struct {
	TxId              string `json:"txId"`
	BlockHeight       string `json:"blockHeight"`
	State             string `json:"state"`
	TokenType         string `json:"tokenType"`
	ActionType        string `json:"actionType"`
	FromAddress       string `json:"fromAddress"`
	ToAddress         string `json:"toAddress"`
	Amount            string `json:"amount"`
	Token             string `json:"token"`
	InscriptionId     string `json:"inscriptionId"`
	InscriptionNumber string `json:"inscriptionNumber"`
	Index             string `json:"index"`
	Location          string `json:"location"`
	Msg               string `json:"msg"`
	Time              string `json:"time"`
}

type AddressSummary struct {
	ChainFullName                 string `json:"chainFullName"`
	ChainShortName                string `json:"chainShortName"`
	Address                       string `json:"address"`
	ContractAddress               string `json:"contractAddress"`
	IsProducerAddress             bool   `json:"isProducerAddress"`
	Balance                       string `json:"balance"`
	BalanceSymbol                 string `json:"balanceSymbol"`
	TransactionCount              string `json:"transactionCount"`
	Verifying                     string `json:"verifying"`
	SendAmount                    string `json:"sendAmount"`
	ReceiveAmount                 string `json:"receiveAmount"`
	TokenAmount                   string `json:"tokenAmount"`
	TotalTokenValue               string `json:"totalTokenValue"`
	CreateContractAddress         string `json:"createContractAddress"`
	CreateContractTransactionHash string `json:"createContractTransactionHash"`
	FirstTransactionTime          string `json:"firstTransactionTime"`
	LastTransactionTime           string `json:"lastTransactionTime"`
	Token                         string `json:"token"`
	Bandwidth                     string `json:"bandwidth"`
	Energy                        string `json:"energy"`
	VotingRights                  string `json:"votingRights"`
	UnclaimedVotingRewards        string `json:"unclaimedVotingRewards"`
	IsAaAddress                   bool   `json:"isAaAddress"`
}

type TickMarketInfo struct {
	LastPrice string `json:"lastPrice"`
	High24h   string `json:"high24h"`
	Low24h    string `json:"low24h"`
}
