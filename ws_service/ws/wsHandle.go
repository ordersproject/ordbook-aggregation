package ws

import (
	"fmt"
	"github.com/gobwas/ws"
	"net/http"
	"time"
)

var (
	wsItemMap       *WsItemMap
)

func init() {
	wsItemMap = NewWsItemMap()
}

func GetGlobalWsItemMap() *WsItemMap {
	if wsItemMap == nil {
		wsItemMap = NewWsItemMap()
	}
	return wsItemMap
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("uuid")
	if token != "" {

	}
	fmt.Println("Connection-token：", token)

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		//logger.Logger.Errorf("WsClient connect err:%s", err.Error())
		//fmt.Println(conn)
		//return
	}
	remoteAdd := conn.RemoteAddr()
	fmt.Printf("remoteAdd：[%s]-[%s]\n", token, remoteAdd.String())

	//Check client
	co := InitConnect(conn)
	//todo auth
	if _, ok := wsItemMap.Get(co); !ok {
		wsItemMap.Set(co, token)
	}

	defer func() {
		//close
		if _, ok := wsItemMap.Get(co); ok {
			wsItemMap.Deleted(co)
			fmt.Println("WsClient-" + remoteAdd.String() + "-disconnect")
		}
	}()
	//time.Sleep(500 * time.Millisecond)

	for {
		msg, opCode, err := co.ReadMessage()
		if err != nil {
			if _, ok := wsItemMap.Get(co); ok {
				wsItemMap.Deleted(co)
				break
			}
		}else {
			if opCode == ws.OpPing {
				SendHeartBeat(co)
				continue
			}

			wsData := WsDataFromStringMsg(string(msg))
			if wsData == nil {
				//fmt.Println(string(msg))
				fmt.Printf("[CO-%s]wsData is nil\n", remoteAdd)
				SendWSResponseForErr(co, wsItemMap, "wsData is nil or invalid")
				continue
			}

			if wsData.M == "" {
				fmt.Printf("[CO-%s]wsData.M is nil\n", remoteAdd)
				SendWSResponseForErr(co, wsItemMap, "wsData.M is nil")
				continue
			}
			if wsData.M != HEART_BEAT && wsData.D == "" {
				fmt.Printf("[CO-%s]wsData.D is nil\n", remoteAdd)
				SendWSResponseForErr(co, wsItemMap, "wsData.D is nil")
				continue
			}

			//Method
			switch wsData.M {
			case HEART_BEAT: //Heart beat back
				SendHeartBeat(co)
				break
			case WS_CONNECT: //Connect
				//todo
				SetConnection(co, wsItemMap, wsData.D.(string))
				break
			case WS_DISCONNECT:
				goto Back
				break
			default:
				fmt.Printf("[CO-%s]wsData.M can not find\n", remoteAdd)
				SendWSResponseForErr(co, wsItemMap, "WsData.M can not find")
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	Back :{
		//log.Warn("Jump For", 0, major.LogTimeStr())
	}
}

