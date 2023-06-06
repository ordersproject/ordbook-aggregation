package ws

import (
	"encoding/json"
	"errors"
)

type WsService struct {

}

//parse
func (w *WsService) MsgToWSNotifyTick(msg string, req *WsNotifyTick) error  {
	err := json.Unmarshal([]byte(msg), req)
	if err != nil {
		return errors.New("json parse err")
	}
	return nil
}

//parse
func (w *WsService) MsgToWSConnect(msg string, req *WsConnect) error  {
	err := json.Unmarshal([]byte(msg), req)
	if err != nil {
		return errors.New("json parse err")
	}
	return nil
}
