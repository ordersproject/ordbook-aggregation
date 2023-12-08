package order_brc20_service

import (
	"encoding/hex"
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

func CalAllPoolOrderV2(net, chain string, startBlock, endBlock, bigBlock, startBlockTime, endBlockTime, nowTime int64) (map[string]string, int64, map[string]string, int64) {
	var (
		platformStartTime int64  = config.PlatformRewardCalStartTime
		rewardTick        string = config.PlatformRewardTick

		allNoUsedEntityPoolOrderList []*model.PoolBrc20Model
		allEntityPoolOrderList       []*model.PoolBrc20Model
		limit                        int64 = 1000
		unusedLimit                  int64 = 5000
		calBlockLimit                int64 = config.PlatformRewardExtraRewardDuration

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
		orderCalBlockIndex        map[string]int64 = make(map[string]int64)

		//coinPrice int64 = int64(GetMarketPrice(net, tick, fmt.Sprintf("%s-BTC", strings.ToUpper(tick))))
		coinPriceMap map[string]int64 = make(map[string]int64)

		hasNoUsed bool = false

		calPoolRewardInfo            map[string]string = make(map[string]string) //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
		calPoolRewardTotalValue      int64             = 0
		calPoolExtraRewardInfo       map[string]string = make(map[string]string) //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
		calPoolExtraRewardTotalValue int64             = 0
	)

	_ = coinPriceMap
	allNoUsedEntityPoolOrderList, _ = mongo_service.FindPoolBrc20ModelListByStartTimeAndEndTimeAndNoRemove(net, "", "", "",
		model.PoolTypeBoth, unusedLimit, 0, platformStartTime, endBlockTime)
	if allNoUsedEntityPoolOrderList != nil && len(allNoUsedEntityPoolOrderList) != 0 {
		hasNoUsed = true
		for _, v := range allNoUsedEntityPoolOrderList {
			if strings.ToLower(v.Tick) == "rdex" {
				continue
			}

			if v.Timestamp < platformStartTime {
				continue
			}

			lpTimestamp := v.Timestamp
			lpRemoveTime := v.UpdateTime
			lpStartBlock, _ := getPlatformBlockStartTimeByTimestamp(net, chain, lpTimestamp)

			if lpStartBlock < config.PlatformRewardCalStartBlock {
				//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] lpStartBlock[%d] not in event block[%d]\n", v.OrderId, lpStartBlock, config.EventOneStartBlock)
				continue
			}
			if lpStartBlock > endBlock {
				//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] lpStartBlock[%d] not in currrent block[%d]\n", v.OrderId, lpStartBlock, endBlock)
				continue
			}
			calBlockIndex := int64(0)
			if v.DealTime != 0 {
				lpDealStartBlock, _ := getPlatformBlockStartTimeByTimestamp(net, chain, v.DealTime)
				if lpDealStartBlock < endBlock {
					//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] lpDealStartBlock[%d] not in currrent block[%d]\n", v.OrderId, lpDealStartBlock, endBlock)
					continue
				}
			} else if v.PoolState == model.PoolStateRemove {
				lpRemoveStartBlock, _ := getPlatformBlockStartTimeByTimestamp(net, chain, lpRemoveTime)
				if lpRemoveStartBlock < endBlock {
					//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] lpRemoveStartBlock[%d] not in currrent block[%d]\n", v.OrderId, lpRemoveStartBlock, endBlock)
					continue
				}
			}
			calBlockIndex = calBlockPlatformByStartBlock(lpStartBlock, startBlock)
			if calBlockIndex < calBlockLimit {
				//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] calBlockIndex[%d] <= 0 startBlock[%d], lpStartBlock[%s]\n", v.OrderId, calBlockIndex, startBlock, lpStartBlock)
				continue
			}
			if _, ok := orderCalBlockIndex[v.OrderId]; !ok {
				orderCalBlockIndex[v.OrderId] = calBlockIndex
			}

			coinPrice := int64(1)
			coinPrice = int64(v.CoinPrice)
			coinPriceDecimalNum := v.CoinPriceDecimalNum
			if coinPrice == 0 {
				coinPrice = 1
			}
			coinAmountDe := decimal.NewFromInt(int64(v.CoinAmount))
			coinPriceDe := decimal.NewFromInt(coinPrice)
			coinAmountToAmount := coinAmountDe.Mul(coinPriceDe).Div(decimal.New(1, coinPriceDecimalNum)).IntPart()

			totalCoinAmountNoUsed = totalCoinAmountNoUsed + coinAmountToAmount
			if _, ok := orderCoinAmountInfoNoUsed[v.OrderId]; ok {
				orderCoinAmountInfoNoUsed[v.OrderId] = orderCoinAmountInfoNoUsed[v.OrderId] + coinAmountToAmount
			} else {
				orderCoinAmountInfoNoUsed[v.OrderId] = coinAmountToAmount
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

			calBlockIndex := int64(0)
			if _, ok := orderCalBlockIndex[orderId]; ok {
				calBlockIndex = orderCalBlockIndex[orderId]
			}
			if calBlockIndex < calBlockLimit {
				continue
			}

			recordOrderId := fmt.Sprintf("%s_%s_%d_%d_%s_%d_%d_%d", orderEntity.Net, orderEntity.Tick, calBlockIndex, calBlockIndex, orderEntity.OrderId, startBlock, endBlock, orderEntity.Timestamp)
			recordOrderId = hex.EncodeToString(tool.SHA256([]byte(recordOrderId)))

			poolExtraRewardRecord, _ := mongo_service.FindRewardRecordModelByOrderId(recordOrderId)
			if poolExtraRewardRecord != nil {
				continue
			}
			poolExtraRewardRecord = &model.RewardRecordModel{
				Net:                 orderEntity.Net,
				Tick:                orderEntity.Tick,
				OrderId:             recordOrderId,
				Pair:                orderEntity.Pair,
				RewardTick:          rewardTick,
				FromOrderId:         orderEntity.OrderId,
				FromOrderRole:       "",
				FromOrderTotalValue: 0,
				FromOrderOwnValue:   0,
				Address:             orderEntity.CoinAddress,
				TotalValue:          allTotalValueNoUsed,
				OwnValue:            orderTotalValue,
				Percentage:          percentage,
				RewardAmount:        rewardAmount,
				RewardType:          model.RewardTypeExtra,
				CalBigBlock:         bigBlock,
				CalDayIndex:         calBlockIndex,
				CalDay:              calBlockIndex,
				CalStartTime:        startBlockTime,
				CalEndTime:          endBlockTime,
				CalStartBlock:       startBlock,
				CalEndBlock:         endBlock,
				Version:             1,
				Timestamp:           nowTime,
			}
			_, err = mongo_service.SetRewardRecordModel(poolExtraRewardRecord)
			if err != nil {
				major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][no-used]SetRewardRecordModel err:%s", err.Error()))
			}
			major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][no-used]SetPoolBrc20ModelForCalExtraReward success [%s]", orderId))

			coinPrice := int64(orderEntity.CoinPrice)
			coinPriceDecimalNum := orderEntity.CoinPriceDecimalNum
			coinPriceDe := decimal.NewFromInt(coinPrice)
			coinPriceRateStr := "1"
			coinPriceRateStr = coinPriceDe.Div(decimal.New(1, coinPriceDecimalNum)).String()

			calPoolExtraRewardInfo[orderId] = fmt.Sprintf("%d:%d:%d:%d:%s", orderTotalValue, percentage, orderEntity.Amount, orderEntity.CoinAmount, coinPriceRateStr)
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
		coinPrice = int64(v.CoinPrice)
		coinPriceDecimalNum := v.CoinPriceDecimalNum
		if coinPrice == 0 {
			coinPrice = 1
		}
		coinAmountDe := decimal.NewFromInt(int64(coinAmount))
		coinPriceDe := decimal.NewFromInt(coinPrice)
		coinAmountToAmount := coinAmountDe.Mul(coinPriceDe).Div(decimal.New(1, coinPriceDecimalNum)).IntPart()

		totalCoinAmount = totalCoinAmount + coinAmountToAmount
		if _, ok := orderCoinAmountInfo[v.OrderId]; ok {
			orderCoinAmountInfo[v.OrderId] = orderCoinAmountInfo[v.OrderId] + coinAmountToAmount
		} else {
			orderCoinAmountInfo[v.OrderId] = coinAmountToAmount
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

		if orderEntity.ClaimTxBlockState == model.ClaimTxBlockStateConfirmed {
			rewardNowAmount := getRealNowRewardByDecreasing(orderEntity.RewardAmount, orderEntity.Decreasing)
			orderEntity.RewardRealAmount = rewardNowAmount
			err := mongo_service.SetPoolBrc20ModelForClaim(orderEntity)
			if err != nil {
				major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][block]SetPoolBrc20ModelForClaim err:%s", err.Error()))
			}
		}

		coinPrice := int64(orderEntity.CoinPrice)
		coinPriceDecimalNum := orderEntity.CoinPriceDecimalNum
		coinPriceDe := decimal.NewFromInt(coinPrice)
		coinPriceRateStr := "1"
		coinPriceRateStr = coinPriceDe.Div(decimal.New(1, coinPriceDecimalNum)).String()

		major.Println(fmt.Sprintf("[CAL-POOL-BLOCK_USER][block]SetPoolBrc20ModelForCalReward success [%s]", orderId))

		calPoolRewardInfo[orderId] = fmt.Sprintf("%d:%d:%d:%d:%s:%d:%d", orderTotalValue, percentage, orderEntity.Amount, orderEntity.CoinAmount, coinPriceRateStr, orderEntity.DealCoinTxBlock, orderEntity.PoolType)
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

func UpdatePoolBlockInfo(bigBlock, startBlock, endBlock, cycleBlock, nowTime int64,
	calPoolRewardInfo map[string]string, calPoolRewardTotalValue int64,
	calPoolExtraRewardInfo map[string]string, calPoolExtraRewardTotalValue int64,
	calEventBidDealExtraRewardInfo map[string]string, calEventBidDealExtraRewardTotalValue int64,
	calType model.CalType) {
	var (
		entity *model.PoolBlockInfoModel
	)

	entity = &model.PoolBlockInfoModel{
		BigBlockId:                           fmt.Sprintf("%d_%d_%d", bigBlock, cycleBlock, calType),
		BigBlock:                             bigBlock,
		StartBlock:                           startBlock,
		EndBlock:                             endBlock,
		CycleBlock:                           cycleBlock,
		Timestamp:                            nowTime,
		CalPoolRewardInfo:                    calPoolRewardInfo,
		CalPoolRewardTotalValue:              calPoolRewardTotalValue,
		CalPoolExtraRewardInfo:               calPoolExtraRewardInfo,
		CalPoolExtraRewardTotalValue:         calPoolExtraRewardTotalValue,
		CalEventBidDealExtraRewardInfo:       calEventBidDealExtraRewardInfo,
		CalEventBidDealExtraRewardTotalValue: calEventBidDealExtraRewardTotalValue,
		CalType:                              calType,
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

func GetNowCalStartTimeAndEndTime() (int64, int64, int64) {
	var (
		nowTime                  int64 = tool.MakeTimestamp()
		calDay                   int64 = 0
		calStartTime, calEndTime int64 = 0, 0
		dayDistance              int64 = 1000 * 60 * 60 * 24
	)
	for {
		if nowTime >= config.PlatformRewardCalStartTime+calDay*dayDistance && nowTime < config.EventOneStartTime+(calDay+1)*dayDistance {
			calDay = calDay + 1
			calStartTime = config.PlatformRewardCalStartTime + calDay*dayDistance
			calEndTime = config.PlatformRewardCalStartTime + (calDay+1)*dayDistance - 1
			break
		}
	}
	return calStartTime, calEndTime, calDay
}

func getPlatformBlockStartTimeByTimestamp(net, chain string, lpTime int64) (int64, int64) {
	var (
		lpStartBlockTime int64 = 0
		lpStartBlock     int64 = 0
		lpBlock          int64 = 0
		lpBlockInfo      *model.BlockInfoModel
		calStartBlock    int64 = config.PlatformRewardCalStartBlock
		calEndBlock      int64 = config.PlatformRewardCalEndBlock
		calCycleBlock    int64 = config.PlatformRewardCalCycleBlock
		blockCount       int64 = (calEndBlock - calStartBlock) / calCycleBlock
	)

	lpBlockInfo, _ = mongo_service.FindNewestHeightBlockInfoModelByBlockTime(net, chain, lpTime/1000)
	if lpBlockInfo == nil {
		return 0, 0
	}
	lpBlock = lpBlockInfo.Height
	for i := int64(0); i <= blockCount; i++ {
		if lpBlock >= calStartBlock+i*calCycleBlock && lpBlock < calStartBlock+(i+1)*calCycleBlock {
			lpStartBlock = calStartBlock + i*calCycleBlock
			break
		}
	}
	lpStartBlockId := fmt.Sprintf("%s_%s_%d", net, chain, lpStartBlock)
	lpStartBlockInfo, _ := mongo_service.FindBlockInfoModelByBlockId(lpStartBlockId)
	if lpStartBlockInfo != nil {
		lpStartBlockTime = lpStartBlockInfo.BlockTime
	}
	return lpStartBlock, lpStartBlockTime
}

func calBlockPlatformByStartBlock(startBlock1, startBlock2 int64) int64 {
	var (
		cycleBlock int64 = config.PlatformRewardCalCycleBlock
		dayCount   int64 = (startBlock2 - startBlock1) / cycleBlock
	)
	return dayCount
}
