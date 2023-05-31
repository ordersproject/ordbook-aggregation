package cache_service

import (
	"fmt"
	"sync"
)

var _inscribeItemMap *InscribeItemMap

func init() {
	_inscribeItemMap = NewInscribeItemMap()
}

type InscribeItemMap struct {
	InscribeInfoExist map[string]*InscribeInfo
	Lock          *sync.RWMutex
	//Lock          *sync.Mutex
}

type InscribeInfo struct {
	FromPrivateKeyHex string
	Content           string
	ToTaprootAddress  string
	Fee int64
}

func NewInscribeItemMap() *InscribeItemMap {
	userInfoItemMap := &InscribeItemMap{
		InscribeInfoExist: make(map[string]*InscribeInfo),
		Lock: new(sync.RWMutex),
		//Lock: new(sync.Mutex),
	}
	return userInfoItemMap
}

func GetMetaNameOrderItemMap() *InscribeItemMap {
	if _inscribeItemMap == nil {
		_inscribeItemMap = NewInscribeItemMap()
	}
	return _inscribeItemMap
}

func (u InscribeItemMap) Get(fromAddress string) (*InscribeInfo, bool)  {
	u.Lock.RLock()
	defer u.Lock.RUnlock()
	//u.Lock.Lock()
	//defer u.Lock.Unlock()
	if v, ok := u.InscribeInfoExist[fromAddress]; ok {
		return v, true
	}
	return nil, false
}

func (u InscribeItemMap) GetAndSet(fromAddress string, newInfo *InscribeInfo) (*InscribeInfo, bool)  {
	u.Lock.RLock()
	defer u.Lock.RUnlock()
	//u.Lock.Lock()
	//fmt.Printf("锁住[%s]Get\n", orderInfo)
	//defer u.Lock.Unlock()
	if v, ok := u.InscribeInfoExist[fromAddress]; ok {
		return v, true
	}
	u.InscribeInfoExist[fromAddress] = newInfo
	return newInfo, false
}

func (u InscribeItemMap) Set(fromAddress string, newInfo *InscribeInfo)  {
	u.Lock.Lock()
	fmt.Printf("锁住[%s]Set\n", fromAddress)
	defer u.Lock.Unlock()
	u.InscribeInfoExist[fromAddress] = newInfo
}


func (u InscribeItemMap) Deleted(orderInfo string)  {
	u.Lock.Lock()
	defer u.Lock.Unlock()
	newInscribeInfoItems := make(map[string]*InscribeInfo)
	fmt.Println("*********清除前：", len(u.InscribeInfoExist))
	for m, v := range u.InscribeInfoExist {
		if m == orderInfo {
			continue
		}
		newInscribeInfoItems[m] = v
	}
	fmt.Println("*********清除后：", len(newInscribeInfoItems))
	u.InscribeInfoExist = nil
	u.InscribeInfoExist = newInscribeInfoItems
}
