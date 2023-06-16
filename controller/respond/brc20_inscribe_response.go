package respond

type Brc20PreResp struct {
	FeeAddress string `json:"feeAddress"`
	Fee        int64  `json:"fee"`
}

type Brc20CommitResp struct {
	CommitTxHash  string `json:"commitTxHash"`
	RevealTxHash  string `json:"revealTxHash"`
	InscriptionId string `json:"inscriptionId"`
}

type Brc20TransferCommitResp struct {
	CommitTxHash  string `json:"commitTxHash"`
	RevealTxHash  string `json:"revealTxHash"`
	InscriptionId string `json:"inscriptionId"`
}


type Brc20TransferCommitBatchResp struct {
	Fees  int64 `json:"fees"`
	CommitTxHash  string `json:"commitTxHash"`
	RevealTxHashList  []string `json:"revealTxHashList"`
	InscriptionIdList []string `json:"inscriptionIdList"`
}