package ws

import (
	"fmt"
	"net/http"
	"ordbook-aggregation/config"
)

func StartWS() {
	http.HandleFunc("/ws", WsHandler)
	fmt.Printf("Start WS base service - WsPort[%s]\n", config.WsPort)
	err := http.ListenAndServe(fmt.Sprintf(":%s", config.WsPort), nil)
	if err != nil {
		//log.Fatal(util.AddStr("ListenAndServe:", err), 0, major.LogTimeStr())
		panic(err)
	}
}