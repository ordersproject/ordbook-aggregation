package cache_service

import (
	"sync"
)

const (
	CacheLockUtxoTypeDummy                          = "dummy"
	CacheLockUtxoTypeDummy1200                      = "dummy1200"
	CacheLockUtxoTypeDummyAsk                       = "dummyask"
	CacheLockUtxoTypeDummy1200Ask                   = "dummy1200ask"
	CacheLockUtxoTypeDummyBidX                      = "dummybidx"
	CacheLockUtxoTypeDummy1200BidX                  = "dummy1200bidx"
	CacheLockUtxoTypeBidpay                         = "bidpay"
	CacheLockUtxoTypeMultiSigInscription            = "multisiginscriptionpay"
	CacheLockUtxoTypeMultiSigInscriptionFromRelease = "multisiginscriptionpayfromrelease"
	CacheLockUtxoTypeRewardInscription              = "rewardinscriptionpay"
	CacheLockUtxoTypeRewardSend                     = "rewardsendpay"
	CacheLockUtxoTypeLoop                           = "looppay"
)

var _lockUtxoItemMap *LockUtxoItemMap

func init() {
	_lockUtxoItemMap = NewLockUtxoItemMap()
}

type LockUtxoItemMap struct {
	LockUtxoInfoExist map[string]int // 0-unlock, 1-lock
	Lock              *sync.RWMutex
	//Lock          *sync.Mutex
}

func NewLockUtxoItemMap() *LockUtxoItemMap {
	userInfoItemMap := &LockUtxoItemMap{
		LockUtxoInfoExist: make(map[string]int),
		Lock:              new(sync.RWMutex),
	}
	return userInfoItemMap
}

func GetLockUtxoItemMap() *LockUtxoItemMap {
	if _lockUtxoItemMap == nil {
		_lockUtxoItemMap = NewLockUtxoItemMap()
	}
	return _lockUtxoItemMap
}

func (u LockUtxoItemMap) Get(cacheLockUtxoType string) (int, bool) {
	u.Lock.RLock()
	defer u.Lock.RUnlock()
	if v, ok := u.LockUtxoInfoExist[cacheLockUtxoType]; ok {
		return v, true
	}
	return 0, false
}

func (u LockUtxoItemMap) GetAndSet(cacheLockUtxoType string, lockState int) (int, bool) {
	u.Lock.RLock()
	defer u.Lock.RUnlock()
	if v, ok := u.LockUtxoInfoExist[cacheLockUtxoType]; ok {
		return v, true
	}
	u.LockUtxoInfoExist[cacheLockUtxoType] = lockState
	return lockState, false
}

func (u LockUtxoItemMap) Set(cacheLockUtxoType string, lockState int) {
	u.Lock.Lock()
	defer u.Lock.Unlock()
	u.LockUtxoInfoExist[cacheLockUtxoType] = lockState
}

func (u LockUtxoItemMap) Deleted(cacheLockUtxoType string) {
	u.Lock.Lock()
	defer u.Lock.Unlock()
	//delete() not usable
	newLockUtxoInfoItems := make(map[string]int)
	for m, v := range u.LockUtxoInfoExist {
		if m == cacheLockUtxoType {
			continue
		}
		newLockUtxoInfoItems[m] = v
	}
	u.LockUtxoInfoExist = nil
	u.LockUtxoInfoExist = newLockUtxoInfoItems
}
