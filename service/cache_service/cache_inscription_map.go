package cache_service

import (
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
	ToAddress  string
	Fee int64
	FeeRate int64
}

func NewInscribeItemMap() *InscribeItemMap {
	userInfoItemMap := &InscribeItemMap{
		InscribeInfoExist: make(map[string]*InscribeInfo),
		Lock: new(sync.RWMutex),
		//Lock: new(sync.Mutex),
	}
	return userInfoItemMap
}

func GetInscribeItemMap() *InscribeItemMap {
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
	//defer u.Lock.Unlock()
	if v, ok := u.InscribeInfoExist[fromAddress]; ok {
		return v, true
	}
	u.InscribeInfoExist[fromAddress] = newInfo
	return newInfo, false
}

func (u InscribeItemMap) Set(fromAddress string, newInfo *InscribeInfo)  {
	u.Lock.Lock()
	defer u.Lock.Unlock()
	u.InscribeInfoExist[fromAddress] = newInfo
}


func (u InscribeItemMap) Deleted(orderInfo string)  {
	u.Lock.Lock()
	defer u.Lock.Unlock()
	newInscribeInfoItems := make(map[string]*InscribeInfo)
	for m, v := range u.InscribeInfoExist {
		if m == orderInfo {
			continue
		}
		newInscribeInfoItems[m] = v
	}
	u.InscribeInfoExist = nil
	u.InscribeInfoExist = newInscribeInfoItems
}
