package order_brc20_service

import (
	"fmt"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
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

		if _, ok := addressCoinAmountInfo[v.CoinAddress]; ok {
			addressCoinAmountInfo[v.CoinAddress] = addressCoinAmountInfo[v.CoinAddress] + int64(v.CoinAmount)*coinPrice
		} else {
			addressCoinAmountInfo[v.CoinAddress] = int64(v.CoinAmount) * coinPrice
		}

		totalCoinAmount = totalCoinAmount + int64(v.CoinAmount)*coinPrice
		if v.PoolType == model.PoolTypeBoth {
			totalAmount = totalAmount + int64(v.Amount)
			if _, ok := addressAmountInfo[v.CoinAddress]; ok {
				addressAmountInfo[v.Address] = addressAmountInfo[v.Address] + int64(v.Amount)
			} else {
				addressAmountInfo[v.Address] = int64(v.Amount)
			}
		}
	}
	allTotalValue = totalCoinAmount + totalAmount
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

func UpdatePoolBlockInfo(startBlock, cycleBlock, nowTime int64) {
	var (
		entity   *model.PoolBlockInfoModel
		bigBlock int64 = GetCurrentBigBlock(startBlock)
	)
	entity = &model.PoolBlockInfoModel{
		BigBlockId: fmt.Sprintf("%d_%d", bigBlock, cycleBlock),
		BigBlock:   bigBlock,
		StartBlock: startBlock,
		CycleBlock: cycleBlock,
		Timestamp:  nowTime,
	}
	mongo_service.SetPoolBlockInfoModel(entity)
}

// calculate the proportion of the total amount of the pool in release order
// pool order in state-used
func calEstimatedProportion() {

}
