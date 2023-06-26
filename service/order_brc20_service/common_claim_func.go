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
	maxOrderLimit int64 = 100
)

var (
	unoccupiedClaimOrderLock *sync.RWMutex = new(sync.RWMutex)
)

// GetUnoccupiedClaimBrc20PsbtList get psbt order for claim
func GetUnoccupiedClaimBrc20PsbtList(tick string, count int64) (*model.OrderBrc20Model, error) {
	var (
		net                      string = "livenet"
		claimOrder               *model.OrderBrc20Model
		claimOrderIdKeyList      []string                 = make([]string, 0)
		unoccupiedClaimOrderList []*model.OrderBrc20Model = make([]*model.OrderBrc20Model, 0)
	)
	entityList, _ := mongo_service.FindOrderBrc20ModelList(net, tick, "", "",
		model.OrderTypeSell, model.OrderStatePreClaim,
		maxOrderLimit, 0, 0, "timestamp", 1)
	if entityList == nil || len(entityList) == 0 {
		return nil, errors.New("no Claim order")
	}

	unoccupiedClaimOrderLock.RLock()
	defer unoccupiedClaimOrderLock.RUnlock()

	claimOrderIdKeyList, _ = redis.GetClaimOrderInfoKeyValueList(redis.CacheGetClaimOrder_)
	for _, v := range entityList {
		has := false
		for _, orderId := range claimOrderIdKeyList {
			if orderId == v.OrderId {
				has = true
				break
			}
		}
		if has {
			continue
		}
		unoccupiedClaimOrderList = append(unoccupiedClaimOrderList, v)
	}

	if int64(len(unoccupiedClaimOrderList)) < count {
		return nil, errors.New("Unoccupied-ClaimOrder: Not enough")
	}
	unoccupiedClaimOrderList = unoccupiedClaimOrderList[:count]
	cacheClaimOrderList(unoccupiedClaimOrderList)

	return claimOrder, nil
}

// ReleaseClaimOrderList release claim order cache
func ReleaseClaimOrderList(claimOrderList []*model.OrderBrc20Model) {
	for _, v := range claimOrderList {
		err := redis.UnSetClaimOrderInfo(v.OrderId)
		if err != nil {
			fmt.Printf("UnSetClaimOrderInfo err:%s\n", err.Error())
		}
	}
}

func cacheClaimOrderList(claimOrderList []*model.OrderBrc20Model) {
	for _, v := range claimOrderList {
		_, err := redis.SetRedisClaimOrderInfo(v.OrderId)
		if err != nil {
			fmt.Printf("SetRedisClaimOrderInfo err:%s\n", err.Error())
		}
	}
}
