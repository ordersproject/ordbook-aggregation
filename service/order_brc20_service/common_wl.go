package order_brc20_service

import (
	"errors"
	"fmt"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
)

func getWhitelistCount(net, tick, address, ip string, whitelistType model.WhitelistType) (int64, int, error) {
	var (
		claimCoinAmount              int64 = 2000
		entity                       *model.WhitelistModel
		todayStartTime, todayEndTime int64 = tool.GetToday0Time(), tool.GetToday24Time()
		count                        int64 = 0
	)
	if whitelistType == model.WhitelistTypeClaim1w {
		claimCoinAmount = 10000
	}

	_ = todayEndTime
	//for _, v := range inList {
	//	if v == address {
	//		fmt.Printf("[CLAIM]-check startTime[%d][%s], endTime[%d][%s]\n", todayStartTime, tool.MakeDate(todayStartTime), todayEndTime, tool.MakeDate(todayEndTime))
	//		count, _ := mongo_service.CountBuyerOrderBrc20ModelList(net, tick, address, "", model.OrderTypeSell, model.OrderStateFinishClaim, 0, 0)
	//		canCount := dayLimit - count
	//		if canCount <= 0 {
	//			canCount = 0
	//		}
	//		return claimCoinAmount, int(canCount), nil
	//	}
	//}

	entity, _ = mongo_service.FindWhitelistModelByIpAndType(ip, whitelistType)
	if entity != nil {
		if whitelistType == model.WhitelistTypeClaim1w {
			if entity.Timestamp < todayStartTime {
				return claimCoinAmount, 0, nil
			} else {
				count, _ := mongo_service.CountBuyerOrderBrc20ModelList(net, tick, "", ip, model.OrderTypeSell, model.OrderStateFinishClaim, 0, 0, claimCoinAmount)
				canCount := entity.Limit - count
				if canCount <= 0 {
					return claimCoinAmount, 0, errors.New("the address of this ip had claimed")
				}
			}
		} else {
			if entity.Limit == 0 {
				entity.Limit = 1
			}
			count, _ := mongo_service.CountBuyerOrderBrc20ModelList(net, tick, "", ip, model.OrderTypeSell, model.OrderStateFinishClaim, 0, 0, claimCoinAmount)
			canCount := entity.Limit - count
			if canCount <= 0 {
				return claimCoinAmount, 0, errors.New("the address of this ip had claimed")
			}
		}
	}

	entity, _ = mongo_service.FindWhitelistModelByAddressAndType(address, whitelistType)
	if entity == nil || entity.Id == 0 {
		return claimCoinAmount, 0, errors.New("not in whitelist")
	}
	if entity.Limit == 0 {
		entity.Limit = 1
	}

	if whitelistType == model.WhitelistTypeClaim1w {
		count, _ = mongo_service.CountBuyerOrderBrc20ModelList(net, tick, address, "", model.OrderTypeSell, model.OrderStateFinishClaim, 0, 0, claimCoinAmount)
	} else {
		count, _ = mongo_service.CountBuyerOrderBrc20ModelList(net, tick, address, "", model.OrderTypeSell, model.OrderStateFinishClaim, 0, 0, claimCoinAmount)
	}
	canCount := entity.Limit - count
	if canCount <= 0 {
		canCount = 0
	}

	//if entity.WhiteUseState == model.WhiteUseStateYes {
	//	return 0, nil
	//}

	return claimCoinAmount, int(canCount), nil
}

func updateWhiteListUsed(address, ip string, whitelistType model.WhitelistType) {
	var (
		entity *model.WhitelistModel
		err    error
	)
	entity, _ = mongo_service.FindWhitelistModelByAddressAndType(address, whitelistType)
	if entity == nil {
		return
	}
	entity.IP = ip
	entity.WhiteUseState = model.WhiteUseStateYes

	_, err = mongo_service.SetWhitelistModel(entity)
	if err != nil {
		return
	}
}

func setWhitelist(address string, whitelistType model.WhitelistType, limit int64, wlType int) {
	var (
		entity *model.WhitelistModel
		err    error
	)

	entity, _ = mongo_service.FindWhitelistModelByAddressAndType(address, whitelistType)
	if entity != nil {
		if wlType == 1 {
			entity.Limit = entity.Limit + limit
			_, err = mongo_service.SetWhitelistModel(entity)
			if err != nil {
				return
			}
		}
		return
	}
	entity = &model.WhitelistModel{
		AddressId:     fmt.Sprintf("%s_%d", address, whitelistType),
		Address:       address,
		IP:            "",
		WhitelistType: whitelistType,
		WhiteUseState: model.WhiteUseStateNo,
		Limit:         limit,
		Timestamp:     tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetWhitelistModel(entity)
	if err != nil {
		return
	}
	return
}
