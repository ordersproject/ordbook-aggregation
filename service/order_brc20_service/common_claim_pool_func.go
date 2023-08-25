package order_brc20_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/model"
	"ordbook-aggregation/redis"
	"ordbook-aggregation/service/mongo_service"
	"sync"
)

const (
	maxPoolClaimOrderLimit int64 = 100
)

var (
	unoccupiedPoolClaimOrderLock *sync.RWMutex = new(sync.RWMutex)
)

// GetUnoccupiedPoolClaimBrc20PsbtList get psbt order for pool claim
func GetUnoccupiedPoolClaimBrc20PsbtList(net, tick string, count int64, coinAmount int64) (*model.OrderBrc20Model, error) {
	var (
		//net                      string = "livenet"
		poolClaimOrder               *model.OrderBrc20Model
		poolClaimOrderIdKeyList      []string                 = make([]string, 0)
		unoccupiedPoolClaimOrderList []*model.OrderBrc20Model = make([]*model.OrderBrc20Model, 0)
	)
	entityList, _ := mongo_service.FindOrderBrc20ModelList(net, tick, "", "",
		model.OrderTypeSell, model.OrderStatePoolPreClaim,
		maxPoolClaimOrderLimit, 0, 0, "timestamp", 1, model.FreeStatePoolClaim, coinAmount)
	if entityList == nil || len(entityList) == 0 {
		return nil, errors.New("no pool claim order")
	}

	unoccupiedPoolClaimOrderLock.RLock()
	defer unoccupiedPoolClaimOrderLock.RUnlock()

	poolClaimOrderIdKeyList, _ = redis.GetPoolClaimOrderInfoKeyValueList(redis.CacheGetPoolClaimOrder_)
	for _, v := range entityList {
		has := false
		for _, orderId := range poolClaimOrderIdKeyList {
			if orderId == v.OrderId {
				has = true
				break
			}
		}
		if has {
			continue
		}
		unoccupiedPoolClaimOrderList = append(unoccupiedPoolClaimOrderList, v)
	}

	if int64(len(unoccupiedPoolClaimOrderList)) < count {
		return nil, errors.New("Unoccupied-PoolClaimOrder: Not enough")
	}

	unoccupiedPoolClaimOrderList = unoccupiedPoolClaimOrderList[:count]
	fmt.Printf("[Cache][PoolClaimOrder]-count:%d\n", len(unoccupiedPoolClaimOrderList))
	cachePoolClaimOrderList(unoccupiedPoolClaimOrderList)
	poolClaimOrder = unoccupiedPoolClaimOrderList[0]

	return poolClaimOrder, nil
}

// ReleasePoolClaimOrderList release claim order cache
func ReleasePoolClaimOrderList(poolClaimOrderList []*model.OrderBrc20Model) {
	for _, v := range poolClaimOrderList {
		err := redis.UnSetPoolClaimOrderInfo(v.OrderId)
		if err != nil {
			fmt.Printf("UnSetPoolClaimOrderInfo err:%s\n", err.Error())
		}
	}
}

func cachePoolClaimOrderList(poolClaimOrderList []*model.OrderBrc20Model) {
	for _, v := range poolClaimOrderList {
		_, err := redis.SetRedisPoolClaimOrderInfo(v.OrderId)
		if err != nil {
			fmt.Printf("SetRedisPoolClaimOrderInfo err:%s\n", err.Error())
		}
	}
}
