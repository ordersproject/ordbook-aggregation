package parse

import (
	"fmt"
	"github.com/shopspring/decimal"
	"testing"
)

func Test_parseTx(t *testing.T) {
	txRaw := ""
	parseTx(txRaw)
}

func Test_parsePsbt(t *testing.T) {
	psbtRaw := ""
	parsePsbt(psbtRaw)
}

func Test_decimal(t *testing.T) {
	//fmt.Println(strconv.ParseInt("1.24", 10, 64))
	strDe, _ := decimal.NewFromString("0.2240030000")

	fmt.Println(strDe.String())
}
