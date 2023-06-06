package ws

import (
	"fmt"
)

//Send heart
func SendHeartBeat(c *Connection) {
	if v, ok := wsItemMap.Get(c); ok {
		wsData := &WsData{
			M: HEART_BEAT,
			C: WS_CODE_HEART_BEAT_BACK,
			//D: "Heart beat back",
		}
		SendMsgToConn(c, wsData)
	}else {
		fmt.Println(fmt.Sprintf("[%s] - Disconnected", v))
	}
}

func SendTickInfo(data interface{})  {
	items, _ := wsItemMap.GetAllConn()
	wsData := &WsData{
		M: WS_SERVER_NOTIFY_TICK_INFO,
		C: WS_CODE_SERVER,
		D: data,
	}
	for _, c := range items {
		SendMsgToConn(c, wsData)
	}
}

//Common func for send
func SendMsgToConn(c *Connection, wsData *WsData)  {
	wsDataStr,err := wsData.ToString()
	if err != nil {
		fmt.Printf("wsData.ToString err:%s\n", err.Error())
		return
	}
	err = c.WriteMessage([]byte(wsDataStr))
	if err != nil {
		// handle error
		//log.Error("WriteServerMessage err:" + err.Error(), 0, major.LogTimeStr())
		fmt.Printf("WriteServerMessage err:%s\n", err.Error())
	}else {
		if wsData.M == HEART_BEAT {

		}
	}
	return
}

//Cache connection
func SetConnection(c *Connection, wsItemMap *WsItemMap, msg string)  {
	wsConnect := &WsConnect{}
	if err := new(WsService).MsgToWSConnect(msg, wsConnect); err != nil{
		//log.Error("MsgToWSConnect err:" + err.Error(), 0, major.LogTimeStr())
		return
	}
	if wsConnect.Uuid == "" || len(wsConnect.Uuid) == 0 {
		wsData := &WsData{
			M: WS_RESPONSE_ERROR,
			C: WS_CODE_SEND_ERROR,
			D: "Uuid is empty",
		}
		SendMsgToConn(c, wsData)
		return
	}
	wsItemMap.Set(c, wsConnect.Uuid)
	wsData := &WsData{
		M: WS_RESPONSE_SUCCESS,
		C: WS_CODE_SEND_SUCCESS,
		D: "Connected",
	}
	SendMsgToConn(c, wsData)
}

//WS reply err
func SendWSResponseForErr(c *Connection, wsItemMap *WsItemMap, msg string) {
	if _, ok := wsItemMap.Get(c); !ok {
		fmt.Println("conn is not exist")
		//log.Error("conn is not exist", 0, major.LogTimeStr())
		return
	} else {
		wsData := &WsData{
			M: WS_RESPONSE_ERROR,
			C: WS_CODE_SEND_ERROR,
			D: msg,
		}
		SendMsgToConn(c, wsData)
		return
	}
}



