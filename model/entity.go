package model

const (
	P_BRC20 = "brc-20"

	OP_DEPLOP   = "deploy"
	OP_MINT     = "mint"
	OP_TRANSFER = "transfer"
)

type Brc20Protocol struct {
	P    string `json:"p"`
	Op   string `json:"op"`
	Tick string `json:"tick"`
	Amt  string `json:"amt"`
}