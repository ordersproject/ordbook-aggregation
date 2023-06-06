package ws

import (
	"fmt"
	"sync"
	"unsafe"
)

type WsItemMap struct {
	WsItems map[*Connection]string
	Lock    *sync.RWMutex
}

func NewWsItemMap() *WsItemMap {
	wsItemMap := &WsItemMap{
		WsItems: make(map[*Connection]string),
		Lock: new(sync.RWMutex),
	}
	return wsItemMap
}

func (i WsItemMap) Get(c *Connection) (string, bool)  {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	if v, ok := i.WsItems[c]; ok {
		return v, true
	}
	return "", false
}

func (i WsItemMap) Set(c *Connection, w string)  {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	i.WsItems[c] = w
}

func (i WsItemMap) GetConnFromString(str string) ([]*Connection, bool)  {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	coList := make([]*Connection, 0)
	for conn, v := range i.WsItems {
		if str == v {
			coList = append(coList, conn)
			//return conn, true
		}
	}
	return coList, true
}


func (i WsItemMap) GetAllConn() ([]*Connection, bool)  {
	i.Lock.RLock()
	defer i.Lock.RUnlock()
	coList := make([]*Connection, 0)
	for conn, _ := range i.WsItems {
		coList = append(coList, conn)
	}
	return coList, true
}

func (i WsItemMap) Init()  {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	for key,_ := range i.WsItems{
		delete(i.WsItems, key)
	}
}

func (i WsItemMap) Deleted(c *Connection)  {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	newWsItems := make(map[*Connection]string)
	for conn, v := range i.WsItems {
		if conn == c {
			continue
		}
		newWsItems[conn] = v
	}
	if !c.isClose {
		c.close()
	}
	i.WsItems = nil
	i.WsItems = newWsItems
	fmt.Println(fmt.Sprintf("*****************WsItems-[%d] size-[%v]****************", len(i.WsItems), unsafe.Sizeof(i.WsItems)))
	//delete(i.WsItems, c)
}
