package ws

import (
	"fmt"
	"testing"
)

func TestWsData_ToString(t *testing.T) {
	w := &WsData{
		M: "HEART_BEAT",
		C: 10,
		//D: "",
	}
	res, _ := w.ToString()
	fmt.Println(res)
}

func TestWsDataFromStringMsg(t *testing.T) {
	//msg := `{\"M\":\"HEART_BEAT\",\"C\":0}`
	//msg := `{\"C\":0}`
	//msg := `"{\"M\":\"HEART_BEAT\",\"C\":0}"`
	msg := `{"M":"HEART_BEAT","C":10}`
	//msg := `{\"M\":\"HEART_BEAT\"}`
	res := WsDataFromStringMsg(msg)
	fmt.Println(res)
}