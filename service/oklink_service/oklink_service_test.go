package oklink_service

import (
	"fmt"
	"ordbook-aggregation/config"
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
	for _, v := range res.OutputDetails {
		fmt.Println(*v)
	}

}

func TestGetAddressSummary(t *testing.T) {
	config.InitConfig()
	address := "bc1q98hfp00j259u93szt7cnfgfy38wy8xve3lh3qr"
	res, err := GetAddressSummary(address)
	if err != nil {
		fmt.Printf("Err:%s\n", err.Error())
		return
	}
	fmt.Printf("Res: %+v\n", res)
}

func TestGetInscriptions(t *testing.T) {
	config.InitConfig()
	var (
		token             = ""
		inscriptionId     = "3b6197149e850c118a4f9121ffd29738d3b0259e334a1cb1ad6adddf7cc9527ei0"
		inscriptionNumber = ""
	)
	res, err := GetInscriptions(token, inscriptionId, inscriptionNumber, 1, 100)
	if err != nil {
		fmt.Printf("Err:%s\n", err.Error())
		return
	}
	fmt.Printf("Res: %+v\n", res)
	for _, v := range res.InscriptionsList {
		fmt.Printf("Item: %+v\n", *v)
	}
}
