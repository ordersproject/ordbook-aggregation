package order_brc20_service

import (
	"fmt"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/node"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
	"strings"
)

//func getOwnerRewardRate() {
//
//}
//
//func getOwnerRewardPoints(net, tick, address string) (int64, int64) {
//	var (
//		entityAllRewardOrderCount *model.PoolRewardOrderCount
//	)
//	entityAllRewardOrderCount, _ = mongo_service.CountPoolRewardOrder(net, tick, "", address, model.PoolStateClaim)
//	if entityAllRewardOrderCount == nil {
//		return 0, 0
//	}
//	return entityAllRewardOrderCount.CoinAmountTotal, entityAllRewardOrderCount.AmountTotal
//}
//
//func getAllRewardPoints(net, tick string) (int64, int64) {
//	var (
//		entityAllRewardOrderCount *model.PoolRewardOrderCount
//	)
//	entityAllRewardOrderCount, _ = mongo_service.CountPoolRewardOrder(net, tick, "", "", model.PoolStateClaim)
//	if entityAllRewardOrderCount == nil {
//		return 0, 0
//	}
//	return entityAllRewardOrderCount.CoinAmountTotal, entityAllRewardOrderCount.AmountTotal
//}

func getBaseReward() int64 {
	var (
		net        string = "livenet"
		date       string = tool.MakeDateV3(tool.GetYesterday24Time())
		baseReward int64  = 0
		entity     *model.OrderBrc20MarketInfoModel
		bidVolume  int64 = 0
	)
	entity, _ = mongo_service.FindOrderBrc20MarketInfoModelByPair(net, date)
	if entity != nil {
		bidVolume = entity.BidVolume
	}
	baseReward = config.PlatformRewardDayBase / (bidVolume + 1)
	return baseReward
}

func getSinglePoolReward() int64 {
	return getBaseReward() / 2
}

func getDoublePoolReward(ratio int64) int64 {
	return getBaseReward() * ratio / 10
}

func getRealNowReward(entityOrder *model.PoolBrc20Model) int64 {
	var (
		rewardAmount           int64 = 0
		rewardNowAmount        int64 = 0
		decreasingRewardAmount int64 = 0
	)
	rewardAmount = entityOrder.RewardAmount
	disTime := tool.MakeTimestamp() - entityOrder.DealTime
	days := disTime / (1000 * 60 * 60 * 24)
	rewardAmountDe := decimal.NewFromInt(rewardAmount)
	for i := int64(1); i <= days; i++ {
		if i <= config.PlatformRewardDiminishingDays {
			continue
		}
		if i > config.PlatformRewardDiminishingDays && i <= config.PlatformRewardDiminishingPeriod+config.PlatformRewardDiminishingDays {
			decreasingRewardAmount = decreasingRewardAmount + rewardAmountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing1)).Div(decimal.NewFromInt(100)).IntPart()
		} else if i > config.PlatformRewardDiminishingPeriod+config.PlatformRewardDiminishingDays && i <= config.PlatformRewardDiminishingPeriod*2+config.PlatformRewardDiminishingDays {
			decreasingRewardAmount = decreasingRewardAmount + rewardAmountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing2)).Div(decimal.NewFromInt(100)).IntPart()
		} else if i > config.PlatformRewardDiminishingPeriod*2+config.PlatformRewardDiminishingDays {
			decreasingRewardAmount = decreasingRewardAmount + rewardAmountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing3)).Div(decimal.NewFromInt(100)).IntPart()
		}
	}
	rewardNowAmount = rewardAmount - decreasingRewardAmount
	if rewardNowAmount <= 0 {
		rewardNowAmount = 0
	}
	return rewardNowAmount
}

func getRealNowRewardByDecreasing(rewardAmount, decreasing int64) int64 {
	var (
		rewardNowAmount int64 = 0
	)
	if decreasing <= 0 {
		return rewardAmount
	}
	if decreasing >= 100 {
		return 0
	}
	rewardAmountDe := decimal.NewFromInt(rewardAmount)
	decreasingDe := decimal.NewFromInt(100 - decreasing)
	rewardNowAmount = rewardAmountDe.Mul(decreasingDe).Div(decimal.NewFromInt(100)).IntPart()
	return rewardNowAmount
}

func CalAllPoolOrder(net string, startBlock, endBlock, nowTime int64) {
	var (
		allNoUsedEntityPoolOrderList []*model.PoolBrc20Model
		allEntityPoolOrderList       []*model.PoolBrc20Model
		limit                        int64 = 1000

		totalCoinAmount       int64                                    = 0
		totalAmount           int64                                    = 0
		allTotalValue         int64                                    = 0
		addressCoinAmountInfo map[string]int64                         = make(map[string]int64)
		addressAmountInfo     map[string]int64                         = make(map[string]int64)
		addressBlockInfo      map[string]*model.PoolBlockUserInfoModel = make(map[string]*model.PoolBlockUserInfoModel)

		totalCoinAmountNoUsed       int64                                    = 0
		totalAmountNoUsed           int64                                    = 0
		allTotalValueNoUsed         int64                                    = 0
		addressCoinAmountInfoNoUsed map[string]int64                         = make(map[string]int64)
		addressAmountInfoNoUsed     map[string]int64                         = make(map[string]int64)
		addressBlockInfoNoUsed      map[string]*model.PoolBlockUserInfoModel = make(map[string]*model.PoolBlockUserInfoModel)

		//coinPrice int64 = int64(GetMarketPrice(net, tick, fmt.Sprintf("%s-BTC", strings.ToUpper(tick))))
		coinPriceMap map[string]int64 = make(map[string]int64)

		endTime   int64 = nowTime - 1000*60*60*24*config.PlatformRewardExtraRewardDuration
		hasNoUsed bool  = false
	)

	allNoUsedEntityPoolOrderList, _ = mongo_service.FindPoolBrc20ModelListByEndTime(net, "", "", "",
		model.PoolTypeBoth, model.PoolStateAdd, limit, 0, endTime)
	if allNoUsedEntityPoolOrderList != nil && len(allNoUsedEntityPoolOrderList) != 0 {
		hasNoUsed = true
		for _, v := range allNoUsedEntityPoolOrderList {
			if strings.ToLower(v.Tick) == "rdex" {
				continue
			}

			coinPrice := int64(1)
			if _, ok := coinPriceMap[v.Tick]; ok {
				coinPrice = coinPriceMap[v.Tick]
			} else {
				coinPrice = int64(GetMarketPrice(net, v.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(v.Tick))))
				if coinPrice == 0 {
					coinPrice = 1
				}
				coinPriceMap[v.Tick] = coinPrice
			}

			totalCoinAmountNoUsed = totalCoinAmountNoUsed + int64(v.CoinAmount)*coinPrice
			if _, ok := addressCoinAmountInfoNoUsed[v.CoinAddress]; ok {
				addressCoinAmountInfoNoUsed[v.CoinAddress] = addressCoinAmountInfoNoUsed[v.CoinAddress] + int64(v.CoinAmount)*coinPrice
			} else {
				addressCoinAmountInfoNoUsed[v.CoinAddress] = int64(v.CoinAmount) * coinPrice
			}

			totalAmountNoUsed = totalAmountNoUsed + int64(v.Amount)
			if _, ok := addressAmountInfoNoUsed[v.CoinAddress]; ok {
				addressAmountInfoNoUsed[v.Address] = addressAmountInfoNoUsed[v.Address] + int64(v.Amount)
			} else {
				addressAmountInfoNoUsed[v.Address] = int64(v.Amount)
			}
		}

		allTotalValueNoUsed = totalCoinAmountNoUsed + totalAmountNoUsed
		allTotalValueNoUsedDe := decimal.NewFromInt(allTotalValueNoUsed)
		for address, coinAmount := range addressCoinAmountInfoNoUsed {
			userTotalValue := int64(0)
			amount := int64(0)
			percentage := int64(0)
			rewardAmount := int64(0)
			if _, ok := addressAmountInfoNoUsed[address]; ok {
				amount = addressAmountInfoNoUsed[address]
			}
			userTotalValue = coinAmount + amount
			userTotalValueDe := decimal.NewFromInt(userTotalValue)
			percentage = userTotalValueDe.Div(allTotalValueNoUsedDe).Mul(decimal.NewFromInt(10000)).IntPart()
			rewardAmount = getUserBlockRewardAmountNoUser(percentage)

			blockUserId := fmt.Sprintf("%d_%d_%d_%s", GetMiningBigBlock(startBlock), config.PlatformRewardCalCycleBlock, model.InfoTypeNoUsed, address)
			blockInfo := &model.PoolBlockUserInfoModel{
				BlockUserId:    blockUserId,
				Net:            net,
				InfoType:       model.InfoTypeBlock,
				HasNoUsed:      hasNoUsed,
				Address:        address,
				BigBlock:       GetMiningBigBlock(startBlock),
				StartBlock:     startBlock,
				CycleBlock:     config.PlatformRewardCalCycleBlock,
				CoinPrice:      0,
				CoinAmount:     coinAmount,
				Amount:         amount,
				UserTotalValue: userTotalValue,
				AllTotalValue:  allTotalValue,
				Percentage:     percentage,
				RewardAmount:   rewardAmount,
				Timestamp:      tool.MakeTimestamp(),
			}

			addressBlockInfoNoUsed[address] = blockInfo
		}

		for address, blockInfo := range addressBlockInfoNoUsed {
			_, err := mongo_service.SetPoolBlockUserInfoModel(blockInfo)
			if err != nil {
				major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][no-used]SetPoolBlockUserInfoModel err:%s", err.Error()))
				continue
			}
			major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][no-used]SetPoolBlockUserInfoModel success [%s]", address))
		}

	}

	allEntityPoolOrderList, _ = mongo_service.FindPoolBrc20ModelListByStartAndEndBlock(net, "", "", "",
		model.PoolTypeAll, model.PoolStateClaim, limit, 0, startBlock, endBlock)
	for _, v := range allEntityPoolOrderList {
		if strings.ToLower(v.Tick) == "rdex" {
			continue
		}

		coinAmount, amount, err := calculateDecrement(v)
		if err != nil {
			continue
		}

		coinPrice := int64(1)
		if _, ok := coinPriceMap[v.Tick]; ok {
			coinPrice = coinPriceMap[v.Tick]
		} else {
			coinPrice = int64(GetMarketPrice(net, v.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(v.Tick))))
			if coinPrice == 0 {
				coinPrice = 1
			}
			coinPriceMap[v.Tick] = coinPrice
		}

		totalCoinAmount = totalCoinAmount + int64(coinAmount)*coinPrice
		if _, ok := addressCoinAmountInfo[v.CoinAddress]; ok {
			addressCoinAmountInfo[v.CoinAddress] = addressCoinAmountInfo[v.CoinAddress] + int64(coinAmount)*coinPrice
		} else {
			addressCoinAmountInfo[v.CoinAddress] = int64(coinAmount) * coinPrice
		}

		if v.PoolType == model.PoolTypeBoth {
			totalAmount = totalAmount + int64(amount)
			if _, ok := addressAmountInfo[v.CoinAddress]; ok {
				addressAmountInfo[v.Address] = addressAmountInfo[v.Address] + int64(amount)
			} else {
				addressAmountInfo[v.Address] = int64(amount)
			}
		}
	}
	allTotalValue = totalCoinAmount + totalAmount
	if allTotalValue == 0 {
		major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][block]SetPoolBlockUserInfoModel success [allTotalValue is 0]"))
		return
	}

	allTotalValueDe := decimal.NewFromInt(allTotalValue)

	for address, coinAmount := range addressCoinAmountInfo {
		if _, ok := addressBlockInfo[address]; ok {
			continue
		}
		userTotalValue := int64(0)
		amount := int64(0)
		percentage := int64(0)
		rewardAmount := int64(0)
		if _, ok := addressAmountInfo[address]; ok {
			amount = addressAmountInfo[address]
		}
		userTotalValue = coinAmount + amount
		userTotalValueDe := decimal.NewFromInt(userTotalValue)
		percentage = userTotalValueDe.Div(allTotalValueDe).Mul(decimal.NewFromInt(10000)).IntPart()
		rewardAmount = getUserBlockRewardAmount(percentage, hasNoUsed)
		blockUserId := fmt.Sprintf("%d_%d_%d_%s", GetMiningBigBlock(startBlock), config.PlatformRewardCalCycleBlock, model.InfoTypeBlock, address)
		blockInfo := &model.PoolBlockUserInfoModel{
			BlockUserId:    blockUserId,
			Net:            net,
			InfoType:       model.InfoTypeBlock,
			HasNoUsed:      hasNoUsed,
			Address:        address,
			BigBlock:       GetMiningBigBlock(startBlock),
			StartBlock:     startBlock,
			CycleBlock:     config.PlatformRewardCalCycleBlock,
			CoinPrice:      0,
			CoinAmount:     coinAmount,
			Amount:         amount,
			UserTotalValue: userTotalValue,
			AllTotalValue:  allTotalValue,
			Percentage:     percentage,
			RewardAmount:   rewardAmount,
			Timestamp:      tool.MakeTimestamp(),
		}

		addressBlockInfo[address] = blockInfo
	}

	for address, blockInfo := range addressBlockInfo {
		_, err := mongo_service.SetPoolBlockUserInfoModel(blockInfo)
		if err != nil {
			major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][block]SetPoolBlockUserInfoModel err:%s", err.Error()))
			continue
		}
		major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][block]SetPoolBlockUserInfoModel success [%s]", address))
	}

}

func CalAllPoolOrderV2(net string, startBlock, endBlock, nowTime int64) (map[string]string, int64, map[string]string, int64) {
	var (
		allNoUsedEntityPoolOrderList []*model.PoolBrc20Model
		allEntityPoolOrderList       []*model.PoolBrc20Model
		limit                        int64 = 1000

		totalCoinAmount     int64                                    = 0
		totalAmount         int64                                    = 0
		allTotalValue       int64                                    = 0
		orderCoinAmountInfo map[string]int64                         = make(map[string]int64)
		orderAmountInfo     map[string]int64                         = make(map[string]int64)
		orderBlockInfo      map[string]*model.PoolBlockUserInfoModel = make(map[string]*model.PoolBlockUserInfoModel)

		totalCoinAmountNoUsed     int64            = 0
		totalAmountNoUsed         int64            = 0
		allTotalValueNoUsed       int64            = 0
		orderCoinAmountInfoNoUsed map[string]int64 = make(map[string]int64)
		orderAmountInfoNoUsed     map[string]int64 = make(map[string]int64)
		//orderBlockInfoNoUsed      map[string]*model.PoolBlockUserInfoModel = make(map[string]*model.PoolBlockUserInfoModel)

		//coinPrice int64 = int64(GetMarketPrice(net, tick, fmt.Sprintf("%s-BTC", strings.ToUpper(tick))))
		coinPriceMap map[string]int64 = make(map[string]int64)

		endTime   int64 = nowTime - 1000*60*60*24*config.PlatformRewardExtraRewardDuration
		hasNoUsed bool  = false

		calPoolRewardInfo            map[string]string = make(map[string]string) //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
		calPoolRewardTotalValue      int64             = 0
		calPoolExtraRewardInfo       map[string]string = make(map[string]string) //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
		calPoolExtraRewardTotalValue int64             = 0
	)

	_ = coinPriceMap
	allNoUsedEntityPoolOrderList, _ = mongo_service.FindPoolBrc20ModelListByEndTime(net, "", "", "",
		model.PoolTypeBoth, model.PoolStateAdd, limit, 0, endTime)
	if allNoUsedEntityPoolOrderList != nil && len(allNoUsedEntityPoolOrderList) != 0 {
		hasNoUsed = true
		for _, v := range allNoUsedEntityPoolOrderList {
			if strings.ToLower(v.Tick) == "rdex" {
				continue
			}

			coinPrice := int64(1)
			coinPrice = int64(v.CoinRatePrice)
			if coinPrice == 0 {
				coinPrice = 1
			}

			totalCoinAmountNoUsed = totalCoinAmountNoUsed + int64(v.CoinAmount)*coinPrice
			if _, ok := orderCoinAmountInfoNoUsed[v.OrderId]; ok {
				orderCoinAmountInfoNoUsed[v.OrderId] = orderCoinAmountInfoNoUsed[v.OrderId] + int64(v.CoinAmount)*coinPrice
			} else {
				orderCoinAmountInfoNoUsed[v.OrderId] = int64(v.CoinAmount) * coinPrice
			}

			totalAmountNoUsed = totalAmountNoUsed + int64(v.Amount)
			if _, ok := orderAmountInfoNoUsed[v.OrderId]; ok {
				orderAmountInfoNoUsed[v.OrderId] = orderAmountInfoNoUsed[v.OrderId] + int64(v.Amount)
			} else {
				orderAmountInfoNoUsed[v.OrderId] = int64(v.Amount)
			}
		}

		allTotalValueNoUsed = totalCoinAmountNoUsed + totalAmountNoUsed
		allTotalValueNoUsedDe := decimal.NewFromInt(allTotalValueNoUsed)
		for orderId, coinAmount := range orderCoinAmountInfoNoUsed {
			orderTotalValue := int64(0)
			amount := int64(0)
			percentage := int64(0)
			rewardAmount := int64(0)
			if _, ok := orderAmountInfoNoUsed[orderId]; ok {
				amount = orderAmountInfoNoUsed[orderId]
			}
			orderTotalValue = coinAmount + amount
			orderTotalValueDe := decimal.NewFromInt(orderTotalValue)
			percentage = orderTotalValueDe.Div(allTotalValueNoUsedDe).Mul(decimal.NewFromInt(10000)).IntPart()
			rewardAmount = getUserBlockRewardAmountNoUser(percentage)

			orderEntity, _ := mongo_service.FindPoolBrc20ModelByOrderId(orderId)
			if orderEntity == nil {
				continue
			}
			orderEntity.PercentageExtra = percentage
			orderEntity.RewardExtraAmount = rewardAmount

			err := mongo_service.SetPoolBrc20ModelForCalExtraReward(orderEntity)
			if err != nil {
				major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][no-used]SetPoolBrc20ModelForCalExtraReward err:%s", err.Error()))
				continue
			}
			major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][no-used]SetPoolBrc20ModelForCalExtraReward success [%s]", orderId))

			calPoolExtraRewardInfo[orderId] = fmt.Sprintf("%d:%d:%d:%d:%d", orderTotalValue, percentage, orderEntity.Amount, orderEntity.CoinAmount, orderEntity.CoinRatePrice)
		}
		calPoolExtraRewardTotalValue = allTotalValueNoUsed
	}

	allEntityPoolOrderList, _ = mongo_service.FindUsedAndClaimedPoolBrc20ModelListByDealStartAndDealEndBlock(net, "", "", "",
		model.PoolTypeAll, limit, 0, startBlock, endBlock)
	for _, v := range allEntityPoolOrderList {
		if strings.ToLower(v.Tick) == "rdex" {
			continue
		}

		coinAmount, amount := v.CoinAmount, v.Amount

		coinPrice := int64(1)
		coinPrice = int64(v.CoinRatePrice)
		if coinPrice == 0 {
			coinPrice = 1
		}

		totalCoinAmount = totalCoinAmount + int64(coinAmount)*coinPrice
		if _, ok := orderCoinAmountInfo[v.OrderId]; ok {
			orderCoinAmountInfo[v.OrderId] = orderCoinAmountInfo[v.OrderId] + int64(coinAmount)*coinPrice
		} else {
			orderCoinAmountInfo[v.OrderId] = int64(coinAmount) * coinPrice
		}

		if v.PoolType == model.PoolTypeBoth {
			totalAmount = totalAmount + int64(amount)
			if _, ok := orderAmountInfo[v.OrderId]; ok {
				orderAmountInfo[v.OrderId] = orderAmountInfo[v.OrderId] + int64(amount)
			} else {
				orderAmountInfo[v.OrderId] = int64(amount)
			}
		}
	}
	allTotalValue = totalCoinAmount + totalAmount
	if allTotalValue == 0 {
		major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][block]SetPoolBlockUserInfoModel success [allTotalValue is 0]"))
		return calPoolRewardInfo, calPoolRewardTotalValue, calPoolExtraRewardInfo, calPoolExtraRewardTotalValue
	}

	allTotalValueDe := decimal.NewFromInt(allTotalValue)

	for orderId, coinAmount := range orderCoinAmountInfo {
		if _, ok := orderBlockInfo[orderId]; ok {
			continue
		}
		orderTotalValue := int64(0)
		amount := int64(0)
		percentage := int64(0)
		rewardAmount := int64(0)
		if _, ok := orderAmountInfo[orderId]; ok {
			amount = orderAmountInfo[orderId]
		}
		orderTotalValue = coinAmount + amount
		orderTotalValueDe := decimal.NewFromInt(orderTotalValue)
		percentage = orderTotalValueDe.Div(allTotalValueDe).Mul(decimal.NewFromInt(10000)).IntPart()
		rewardAmount = getUserBlockRewardAmount(percentage, hasNoUsed)

		orderEntity, _ := mongo_service.FindPoolBrc20ModelByOrderId(orderId)
		if orderEntity == nil {
			continue
		}
		orderEntity.Percentage = percentage
		orderEntity.RewardAmount = rewardAmount
		orderEntity.CalValue = orderTotalValue
		orderEntity.CalTotalValue = allTotalValue
		orderEntity.CalStartBlock = startBlock
		orderEntity.CalEndBlock = endBlock

		err := mongo_service.SetPoolBrc20ModelForCalReward(orderEntity)
		if err != nil {
			major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][block]SetPoolBrc20ModelForCalReward err:%s", err.Error()))
			continue
		}
		major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][block]SetPoolBrc20ModelForCalReward success [%s]", orderId))

		calPoolRewardInfo[orderId] = fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d", orderTotalValue, percentage, orderEntity.Amount, orderEntity.CoinAmount, orderEntity.CoinRatePrice, orderEntity.DealCoinTxBlock, orderEntity.PoolType)
	}
	calPoolRewardTotalValue = allTotalValue
	return calPoolRewardInfo, calPoolRewardTotalValue, calPoolExtraRewardInfo, calPoolExtraRewardTotalValue
}

func GetCurrentBigBlock(startBlock int64) int64 {
	var (
		bigBlock int64 = 0
	)
	if startBlock <= config.PlatformRewardCalStartBlock {
		bigBlock = 0
	} else {
		bigBlock = (startBlock - config.PlatformRewardCalStartBlock) / config.PlatformRewardCalCycleBlock
	}
	return bigBlock
}

func GetMiningBigBlock(startBlock int64) int64 {
	if startBlock <= config.PlatformRewardCalStartBlock {
		return -1
	}
	return GetCurrentBigBlock(startBlock) + 1
}

func getUserBlockRewardAmount(percentage int64, hasNoUsed bool) int64 {
	var (
		rewardAmount              int64           = 0
		dayBaseRewardAmount       int64           = config.PlatformRewardDayBase
		dayBaseRewardAmountDe     decimal.Decimal = decimal.NewFromInt(dayBaseRewardAmount)
		percentageDe              decimal.Decimal = decimal.NewFromInt(percentage)
		extraRewardDurationRateDe decimal.Decimal = decimal.NewFromInt(100 - config.PlatformRewardExtraDurationRewardRate)
	)
	if hasNoUsed {
		dayBaseRewardAmountDe = dayBaseRewardAmountDe.Mul(extraRewardDurationRateDe).Div(decimal.NewFromInt(100))
	}
	rewardAmount = dayBaseRewardAmountDe.Mul(percentageDe).Div(decimal.NewFromInt(10000)).IntPart()
	return rewardAmount
}

func getUserBlockRewardAmountNoUser(percentage int64) int64 {
	var (
		rewardAmount              int64           = 0
		dayBaseRewardAmount       int64           = config.PlatformRewardDayBase
		dayBaseRewardAmountDe     decimal.Decimal = decimal.NewFromInt(dayBaseRewardAmount)
		percentageDe              decimal.Decimal = decimal.NewFromInt(percentage)
		extraRewardDurationRateDe decimal.Decimal = decimal.NewFromInt(config.PlatformRewardExtraDurationRewardRate)
	)
	dayBaseRewardAmountDe = dayBaseRewardAmountDe.Mul(extraRewardDurationRateDe).Div(decimal.NewFromInt(100))
	rewardAmount = dayBaseRewardAmountDe.Mul(percentageDe).Div(decimal.NewFromInt(10000)).IntPart()
	return rewardAmount
}

func UpdatePoolBlockInfo(startBlock, endBlock, cycleBlock, nowTime int64,
	calPoolRewardInfo map[string]string, calPoolRewardTotalValue int64, calPoolExtraRewardInfo map[string]string, calPoolExtraRewardTotalValue int64) {
	var (
		entity   *model.PoolBlockInfoModel
		bigBlock int64 = GetCurrentBigBlock(startBlock)
	)
	entity = &model.PoolBlockInfoModel{
		BigBlockId:                   fmt.Sprintf("%d_%d", bigBlock, cycleBlock),
		BigBlock:                     bigBlock,
		StartBlock:                   startBlock,
		EndBlock:                     endBlock,
		CycleBlock:                   cycleBlock,
		Timestamp:                    nowTime,
		CalPoolRewardInfo:            calPoolRewardInfo,
		CalPoolRewardTotalValue:      calPoolRewardTotalValue,
		CalPoolExtraRewardInfo:       calPoolExtraRewardInfo,
		CalPoolExtraRewardTotalValue: calPoolExtraRewardTotalValue,
	}
	mongo_service.SetPoolBlockInfoModel(entity)
}

// Calculate decrement
func calculateDecrement(poolOrder *model.PoolBrc20Model) (int64, int64, error) {
	var (
		coinAmount, amount                     int64 = 0, 0
		decreasingCoinAmount, decreasingAmount int64 = 0, 0
	)
	if poolOrder == nil {
		return 0, 0, fmt.Errorf("poolOrder is nil")
	}
	coinAmount, amount = int64(poolOrder.CoinAmount), int64(poolOrder.Amount)
	coinAmountDe, amountDe := decimal.NewFromInt(coinAmount), decimal.NewFromInt(amount)
	disTime := poolOrder.ClaimTime - poolOrder.DealTime
	days := disTime / (config.PlatformRewardDecreasingCycleTime)
	for i := int64(1); i <= days; i++ {
		if i <= config.PlatformRewardDiminishingDays {
			continue
		}
		if i > config.PlatformRewardDiminishingDays && i <= config.PlatformRewardDiminishingPeriod+config.PlatformRewardDiminishingDays {
			decreasingCoinAmount = decreasingCoinAmount + coinAmountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing1)).Div(decimal.NewFromInt(100)).IntPart()
			decreasingAmount = decreasingAmount + amountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing1)).Div(decimal.NewFromInt(100)).IntPart()
		} else if i > config.PlatformRewardDiminishingPeriod+config.PlatformRewardDiminishingDays && i <= config.PlatformRewardDiminishingPeriod*2+config.PlatformRewardDiminishingDays {
			decreasingCoinAmount = decreasingCoinAmount + coinAmountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing2)).Div(decimal.NewFromInt(100)).IntPart()
			decreasingAmount = decreasingAmount + amountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing2)).Div(decimal.NewFromInt(100)).IntPart()
		} else if i > config.PlatformRewardDiminishingPeriod*2+config.PlatformRewardDiminishingDays {
			decreasingCoinAmount = decreasingCoinAmount + coinAmountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing3)).Div(decimal.NewFromInt(100)).IntPart()
			decreasingAmount = decreasingAmount + amountDe.Mul(decimal.NewFromInt(config.PlatformRewardDiminishing3)).Div(decimal.NewFromInt(100)).IntPart()
		}
	}
	coinAmount = coinAmount - decreasingCoinAmount
	amount = amount - decreasingAmount
	if coinAmount <= 0 {
		coinAmount = 0
	}
	if amount <= 0 {
		amount = 0
	}

	return coinAmount, amount, nil
}

// Calculate decrement for no-release pool order
func calculateDecrementFoNoReleasePool(poolOrder *model.PoolBrc20Model) int64 {
	var (
		proportion int64 = 0
	)
	if poolOrder == nil || poolOrder.DealCoinTxBlock == 0 {
		return 0
	}

	//Use time to calculate the decreasing proportion
	endTime := poolOrder.ClaimTime
	if endTime == 0 {
		endTime = tool.MakeTimestamp()
	}
	disTime := endTime - poolOrder.DealTime
	days := disTime / (config.PlatformRewardDecreasingCycleTime)

	//Use block to calculate the decreasing proportion
	startCalBlock := poolOrder.DealCoinTxBlock
	currentBlockHeight := poolOrder.ClaimTxBlock
	if currentBlockHeight == 0 {
		blockHeight, _ := node.CurrentBlockHeight(poolOrder.Net)
		currentBlockHeight = int64(blockHeight)
		if currentBlockHeight >= startCalBlock {
			disTime = currentBlockHeight - startCalBlock
			days = disTime / (config.PlatformRewardDecreasingCycleBlock)
		}
	}

	fmt.Printf("Days[%d]\n", days)
	for i := int64(1); i <= days; i++ {
		if i <= config.PlatformRewardDiminishingDays {
			continue
		}
		if i > config.PlatformRewardDiminishingDays && i <= config.PlatformRewardDiminishingPeriod+config.PlatformRewardDiminishingDays {
			proportion = proportion + config.PlatformRewardDiminishing1
		} else if i > config.PlatformRewardDiminishingPeriod+config.PlatformRewardDiminishingDays && i <= config.PlatformRewardDiminishingPeriod*2+config.PlatformRewardDiminishingDays {
			proportion = proportion + config.PlatformRewardDiminishing2
		} else if i > config.PlatformRewardDiminishingPeriod*2+config.PlatformRewardDiminishingDays {
			proportion = proportion + config.PlatformRewardDiminishing3
		}
	}

	if proportion >= 100 {
		proportion = 100
	}

	return proportion
}
