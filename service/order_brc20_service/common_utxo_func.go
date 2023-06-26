package order_brc20_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/model"
	"ordbook-aggregation/redis"
	"ordbook-aggregation/service/cache_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
	"sync"
)

const (
	maxLimit int64 = 500
)

var (
	unoccupiedUtxoLock *sync.RWMutex = new(sync.RWMutex)
	saveUtxoLock       *sync.RWMutex = new(sync.RWMutex)
)

func GetUnoccupiedUtxoList(net string, limit int64, utxoType model.UtxoType) ([]*model.OrderUtxoModel, error) {
	var (
		cacheType          string                  = cache_service.CacheLockUtxoTypeDummy
		redisKeyPrefix     string                  = ""
		sortIndexList      []int                   = make([]int, 0)
		utxoIdKeyList      []string                = make([]string, 0)
		startIndex         int64                   = -1
		utxoList           []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		unoccupiedUtxoList []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
	)
	switch utxoType {
	case model.UtxoTypeDummy:
		cacheType = cache_service.CacheLockUtxoTypeDummy
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeDummy_)
		break
	case model.UtxoTypeBidY:
		cacheType = cache_service.CacheLockUtxoTypeBidpay
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeBidY_)
		break
	default:
		return nil, errors.New("Unoccupied-Utxo: wrong type")
	}
	//cache_service.GetLockUtxoItemMap().GetAndSet(cacheType, 1)
	_ = cacheType
	unoccupiedUtxoLock.RLock()
	defer unoccupiedUtxoLock.RUnlock()

	utxoIdKeyList, sortIndexList, _ = redis.GetUtxoInfoKeyValueList(redisKeyPrefix)
	for _, v := range sortIndexList {
		if startIndex == -1 {
			startIndex = int64(v)
		} else if startIndex > int64(v) {
			startIndex = int64(v)
		}
	}
	fmt.Printf("Get utxoIdKeyList: %+v\n", utxoIdKeyList)
	fmt.Printf("Get sortIndexList: %+v\n", sortIndexList)

	utxoList, _ = mongo_service.FindUtxoList(net, startIndex, maxLimit, utxoType)
	if len(utxoList) == 0 {
		return nil, errors.New("Unoccupied-Utxo: Empty utxo list")
	}
	for _, v := range utxoList {
		has := false
		for _, utxoId := range utxoIdKeyList {
			if utxoId == v.UtxoId {
				has = true
				break
			}
		}
		if has {
			continue
		}
		unoccupiedUtxoList = append(unoccupiedUtxoList, v)
	}
	if int64(len(unoccupiedUtxoList)) < limit {
		return nil, errors.New("Unoccupied-Utxo: Not enough")
	}
	unoccupiedUtxoList = unoccupiedUtxoList[:limit]
	cacheUtxoList(unoccupiedUtxoList)
	return unoccupiedUtxoList, nil
}

func ReleaseUtxoList(utxoList []*model.OrderUtxoModel) {
	for _, v := range utxoList {
		cacheUtxoType := redis.UtxoTypeDummy_
		switch v.UtxoType {
		case model.UtxoTypeDummy:
			cacheUtxoType = redis.UtxoTypeDummy_
			break
		case model.UtxoTypeBidY:
			cacheUtxoType = redis.UtxoTypeBidY_
			break
		default:
			continue
		}
		err := redis.UnSetUtxoInfo(cacheUtxoType, v.UtxoId)
		if err != nil {
			fmt.Printf("UnSetUtxoInfo err:%s\n", err.Error())
		}
	}
}

func cacheUtxoList(utxoList []*model.OrderUtxoModel) {
	for _, v := range utxoList {
		cacheUtxoType := redis.UtxoTypeDummy_
		switch v.UtxoType {
		case model.UtxoTypeDummy:
			cacheUtxoType = redis.UtxoTypeDummy_
			break
		case model.UtxoTypeBidY:
			cacheUtxoType = redis.UtxoTypeBidY_
			break
		default:
			continue
		}
		_, err := redis.SetRedisUtxoInfo(cacheUtxoType, v.UtxoId, int(v.SortIndex))
		if err != nil {
			fmt.Printf("SetRedisUtxoInfo err:%s\n", err.Error())
		}
	}
}

func GetSaveStartIndex(net string, utxoType model.UtxoType) int64 {
	saveUtxoLock.RLock()
	t1 := tool.MakeTimestamp()
	fmt.Println("[LOCK]-Save-utxo")
	defer func() {
		saveUtxoLock.RUnlock()
		fmt.Printf("[UNLOCK]-Save-utxo-timeConsuming:%d\n", tool.MakeTimestamp()-t1)
	}()
	startIndex := int64(0)
	latestUtxo, _ := mongo_service.GetLatestStartIndexUtxo(net, utxoType)
	if latestUtxo != nil {
		startIndex = latestUtxo.SortIndex
	}
	return startIndex
}
