package unisat_service

import (
	"fmt"
	"ordbook-aggregation/config"
	"testing"
)

func TestGetAddressUtxo(t *testing.T) {
	config.InitConfig()
	address := "bc1qjmw7nrfaqkxxjz79u3wqdudzkjm2drp2ncqnqp"
	res, err := GetAddressUtxo(address)
	if err != nil {
		fmt.Printf("Err:%s\n", err.Error())
		return
	}
	fmt.Printf("Res:%+v\n", res)
}
