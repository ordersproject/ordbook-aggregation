package order_brc20_service

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"testing"
)

func Test_createMultiSigAddress(t *testing.T) {
	net := &chaincfg.MainNetParams
	pubKeys := []string{
		"037651f0d9d5f5fd74aa04890168888ce01f26702faba2a5fbd820cbc1c638e7a8",
		"037355ad3caeacd0b8e69fd519bf7aac71c3c0227ae446f0c737e4616d7c1ac4f9",
	}
	res, err := createMultiSigAddress(net, pubKeys...)
	if err != nil {
		fmt.Printf("Err:%s\n", err.Error())
		return
	}
	fmt.Printf("Res:%s\n", res)

}
