package hiro_service

type HiroResp struct {
	Limit   int64       `json:"limit"`
	Offset  int64       `json:"offset"`
	Total   int64       `json:"total"`
	Results interface{} `json:"results"`
	Error   string      `json:"error"`
}

type HiroInscription struct {
	Id                   string `json:"id"`
	Number               int64  `json:"number"`
	Address              string `json:"address"`
	GenesisAddress       string `json:"genesis_address"`
	GenesisBlockHeight   int64  `json:"genesis_block_height"`
	GenesisBlockHash     string `json:"genesis_block_hash"`
	GenesisTxId          string `json:"genesis_tx_id"`
	GenesisFee           string `json:"genesis_fee"`
	GenesisTimestamp     int64  `json:"genesis_timestamp"`
	TxId                 string `json:"tx_id"`
	Location             string `json:"location"`
	Output               string `json:"output"`
	Value                string `json:"value"`
	Offset               string `json:"offset"`
	SatOrdinal           string `json:"sat_ordinal"`
	SatRarity            string `json:"sat_rarity"`
	SatCoinbaseHeight    int64  `json:"sat_coinbase_height"`
	MimeType             string `json:"mime_type"`
	ContentType          string `json:"content_type"`
	ContentLength        int64  `json:"content_length"`
	Timestamp            int64  `json:"timestamp"`
}