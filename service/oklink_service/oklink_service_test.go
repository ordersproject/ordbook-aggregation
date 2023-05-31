package oklink_service

import (
	"fmt"
	"testing"
)

func TestGetTxDetail(t *testing.T) {
	txId := "a9ebb4a92acb44b7bac6ab5d7f07482da66aa48f5496107f5592b23c221d0e8f"
	res, err := GetTxDetail(txId)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)
	fmt.Println(res.TxId)
	for _,v := range res.OutputDetails{
		fmt.Println(*v)
	}

}