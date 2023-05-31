package inscription_service

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"testing"
)

func TestCreateKeyAndCalculateInscribe(t *testing.T) {
	netParams := &chaincfg.MainNetParams
	toTaprootAddress := "tb1pa3ee48qwt2uysxsz6gq6qcfuhxq6wdachxmsmgpumsg2ljdk7rvqzgkta2"
	content := "test"
	fromPriHex, fromTaprootAddress, fee, err := CreateKeyAndCalculateInscribe(netParams, toTaprootAddress, content)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(fromPriHex)
	fmt.Println(fromTaprootAddress)
	fmt.Println(fee)
}