package hiro_service

import (
	"fmt"
	"testing"
)

func TestGetInscriptionContent(t *testing.T) {
	inscriptionId := "7bde351cf1e0792775847c9d5c52c5306d49e1bac17f856fcad7e0ae0092e84ei0"
	res, err := GetInscriptionContent(inscriptionId)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(res)
}