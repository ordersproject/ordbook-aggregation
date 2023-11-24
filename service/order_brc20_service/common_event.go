package order_brc20_service

import (
	"fmt"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"strings"
)

// rdex unused pool:20%
// rdex used pool:40%
// rdex bid:40%
func CalAllEventOrder(net string, startBlock, endBlock, nowTime int64) (map[string]string, int64, map[string]string, int64, map[string]string, int64) {
	var (
		tick                             string                   = "rdex"
		allNoUsedEntityRdexPoolOrderList []*model.PoolBrc20Model  //rdex unused pool:20%
		allEntityRdexPoolOrderList       []*model.PoolBrc20Model  //rdex used pool:40%
		allEntityRdexBidOrderList        []*model.OrderBrc20Model //rdex bid:40%

		rewardAmountNoUsed, rewardAmountUsed, rewardAmountBid int64 = getEventRewardDistribution()

		limit int64 = 1000

		totalBidDealAmount     int64                                    = 0
		totalCoinAmount        int64                                    = 0
		totalAmount            int64                                    = 0
		allDealTotalValue      int64                                    = 0
		allTotalValue          int64                                    = 0
		orderBidDealAmountInfo map[string]int64                         = make(map[string]int64)
		orderCoinAmountInfo    map[string]int64                         = make(map[string]int64)
		orderAmountInfo        map[string]int64                         = make(map[string]int64)
		orderBlockInfo         map[string]*model.PoolBlockUserInfoModel = make(map[string]*model.PoolBlockUserInfoModel)

		totalCoinAmountNoUsed     int64            = 0
		totalAmountNoUsed         int64            = 0
		allTotalValueNoUsed       int64            = 0
		orderCoinAmountInfoNoUsed map[string]int64 = make(map[string]int64)
		orderAmountInfoNoUsed     map[string]int64 = make(map[string]int64)

		coinPriceMap map[string]int64 = make(map[string]int64)

		endTime int64 = nowTime - 1000*60*60*24*config.EventOneExtraRewardLpUnusedDuration

		calRdexPoolRewardInfo       map[string]string = make(map[string]string) //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
		calRdexPoolRewardTotalValue int64             = 0

		calRdexPoolExtraRewardInfo       map[string]string = make(map[string]string) //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
		calRdexPoolExtraRewardTotalValue int64             = 0

		calRdexBidDealExtraRewardInfo       map[string]string = make(map[string]string) //{"brc20OrderId":"value:percentage:dealAmount"}
		calRdexBidDealExtraRewardTotalValue int64             = 0
	)

	_ = coinPriceMap
	allNoUsedEntityRdexPoolOrderList, _ = mongo_service.FindPoolBrc20ModelListByEndTime(net, tick, "", "",
		model.PoolTypeBoth, model.PoolStateAdd, limit, 0, endTime)
	if allNoUsedEntityRdexPoolOrderList != nil && len(allNoUsedEntityRdexPoolOrderList) != 0 {
		for _, v := range allNoUsedEntityRdexPoolOrderList {
			if strings.ToLower(v.Tick) != "rdex" {
				continue
			}
			if checkOfficialExcludedAddress(v.Address) {
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
			rewardAmount = getOrderRewardAmountUnused(rewardAmountNoUsed, percentage)

			orderEntity, _ := mongo_service.FindPoolBrc20ModelByOrderId(orderId)
			if orderEntity == nil {
				continue
			}
			orderEntity.PercentageExtra = percentage
			orderEntity.RewardExtraAmount = rewardAmount

			err := mongo_service.SetPoolBrc20ModelForCalExtraReward(orderEntity)
			if err != nil {
				major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][no-used]SetPoolBrc20ModelForCalExtraReward err:%s", err.Error()))
				continue
			}
			major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][no-used]SetPoolBrc20ModelForCalExtraReward success [%s]", orderId))

			calRdexPoolExtraRewardInfo[orderId] = fmt.Sprintf("%d:%d:%d:%d:%d", orderTotalValue, percentage, orderEntity.Amount, orderEntity.CoinAmount, orderEntity.CoinRatePrice)
		}
		calRdexPoolExtraRewardTotalValue = allTotalValueNoUsed
	}

	allEntityRdexPoolOrderList, _ = mongo_service.FindUsedAndClaimedPoolBrc20ModelListByDealStartAndDealEndBlock(net, tick, "", "",
		model.PoolTypeAll, limit, 0, startBlock, endBlock)
	if allEntityRdexPoolOrderList != nil && len(allEntityRdexPoolOrderList) != 0 {
		for _, v := range allEntityRdexPoolOrderList {
			if strings.ToLower(v.Tick) != "rdex" {
				continue
			}
			if checkOfficialExcludedAddress(v.Address) {
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
		if allTotalValue != 0 {
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
				rewardAmount = getOrderRewardAmountUsed(rewardAmountUsed, percentage)

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
					major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][block]SetPoolBrc20ModelForCalReward err:%s", err.Error()))
					continue
				}
				major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][block]SetPoolBrc20ModelForCalReward success [%s]", orderId))

				calRdexPoolRewardInfo[orderId] = fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d", orderTotalValue, percentage, orderEntity.Amount, orderEntity.CoinAmount, orderEntity.CoinRatePrice, orderEntity.DealCoinTxBlock, orderEntity.PoolType)
			}
			calRdexPoolRewardTotalValue = allTotalValue
		} else {
			major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][block]SetPoolBlockUserInfoModel success [allTotalValue is 0]"))
		}
	}

	allEntityRdexBidOrderList, _ = mongo_service.FindDealOrderBrc20ModelListByDealStartAndDealEndBlock(net, tick, model.OrderTypeBuy, model.OrderStateFinish,
		limit, 0, startBlock, endBlock, 2)
	if allEntityRdexBidOrderList != nil && len(allEntityRdexBidOrderList) != 0 {
		for _, v := range allEntityRdexBidOrderList {
			if strings.ToLower(v.Tick) != "rdex" {
				continue
			}
			if checkOfficialExcludedAddress(v.SellerAddress) || checkOfficialExcludedAddress(v.BuyerAddress) {
				continue
			}
			dealAmount := v.Amount
			totalBidDealAmount = totalBidDealAmount + int64(dealAmount)
			if _, ok := orderBidDealAmountInfo[v.OrderId]; ok {
				orderBidDealAmountInfo[v.OrderId] = orderBidDealAmountInfo[v.OrderId] + int64(dealAmount)
			} else {
				orderBidDealAmountInfo[v.OrderId] = int64(dealAmount)
			}
		}
		allDealTotalValue = totalBidDealAmount

		if allDealTotalValue != 0 {
			allDealTotalValueDe := decimal.NewFromInt(allDealTotalValue)
			for orderId, dealAmount := range orderBidDealAmountInfo {
				orderDealTotalValue := int64(dealAmount)
				percentage := int64(0)
				rewardAmount := int64(0)

				orderDealTotalValueDe := decimal.NewFromInt(orderDealTotalValue)
				percentage = orderDealTotalValueDe.Div(allDealTotalValueDe).Mul(decimal.NewFromInt(10000)).IntPart()
				rewardAmount = getOrderRewardAmountBidDeal(rewardAmountBid, percentage)

				orderEntity, _ := mongo_service.FindOrderBrc20ModelByOrderId(orderId)
				if orderEntity == nil {
					continue
				}
				orderEntity.Percentage = percentage
				orderEntity.RewardAmount = rewardAmount
				orderEntity.RewardRealAmount = rewardAmount
				orderEntity.CalValue = orderDealTotalValue
				orderEntity.CalTotalValue = allDealTotalValue
				orderEntity.CalStartBlock = startBlock
				orderEntity.CalEndBlock = endBlock

				err := mongo_service.SetOrderBrc20ModelForCalReward(orderEntity)
				if err != nil {
					major.Println(fmt.Sprintf("[EVENT][CAL-BID-BLOCK_USER][block]SetOrderBrc20ModelForCalReward err:%s", err.Error()))
					continue
				}
				major.Println(fmt.Sprintf("[EVENT][CAL-BID-BLOCK_USER][block]SetOrderBrc20ModelForCalReward success [%s]", orderId))

				calRdexBidDealExtraRewardInfo[orderId] = fmt.Sprintf("%d:%d:%d:%d:%d", orderDealTotalValue, percentage, orderEntity.Amount, orderEntity.DealTxBlock, orderEntity.OrderType)
			}
			calRdexBidDealExtraRewardTotalValue = allDealTotalValue
		} else {
			major.Println(fmt.Sprintf("[EVENT][CAL-BID-BLOCK_USER][block]SetOrderBlockUserInfoModel success [allDealTotalValue is 0]"))
		}
	}

	return calRdexPoolRewardInfo, calRdexPoolRewardTotalValue, calRdexPoolExtraRewardInfo, calRdexPoolExtraRewardTotalValue, calRdexBidDealExtraRewardInfo, calRdexBidDealExtraRewardTotalValue
}

func getEventRewardDistribution() (int64, int64, int64) {
	var (
		rewardAmountAll                                                   int64 = config.EventOneExtraRewardAmount
		rewardAmountNoUsedRate, rewardAmountUsedRate, rewardAmountBidRate int64 = config.EventOneExtraRewardLpUnusedRate, config.EventOneExtraRewardLpUsedRate, config.EventOneExtraRewardBidRate
		rewardAmountNoUsed, rewardAmountUsed, rewardAmountBid             int64 = 0, 0, 0
	)
	rewardAmountAllDe := decimal.NewFromInt(rewardAmountAll)
	rewardAmountNoUsedRateDe := decimal.NewFromInt(rewardAmountNoUsedRate)
	rewardAmountUsedRateDe := decimal.NewFromInt(rewardAmountUsedRate)
	rewardAmountBidRateDe := decimal.NewFromInt(rewardAmountBidRate)
	rewardAmountNoUsed = rewardAmountAllDe.Mul(rewardAmountNoUsedRateDe).Div(decimal.NewFromInt(100)).IntPart()
	rewardAmountUsed = rewardAmountAllDe.Mul(rewardAmountUsedRateDe).Div(decimal.NewFromInt(100)).IntPart()
	rewardAmountBid = rewardAmountAllDe.Mul(rewardAmountBidRateDe).Div(decimal.NewFromInt(100)).IntPart()
	return rewardAmountNoUsed, rewardAmountUsed, rewardAmountBid
}

func getOrderRewardAmountUnused(dayBaseRewardAmountUnused, percentage int64) int64 {
	var (
		rewardAmount          int64           = 0
		dayBaseRewardAmountDe decimal.Decimal = decimal.NewFromInt(dayBaseRewardAmountUnused)
		percentageDe          decimal.Decimal = decimal.NewFromInt(percentage)
	)
	rewardAmount = dayBaseRewardAmountDe.Mul(percentageDe).Div(decimal.NewFromInt(10000)).IntPart()
	return rewardAmount
}
func getOrderRewardAmountUsed(dayBaseRewardAmountUsed, percentage int64) int64 {
	var (
		rewardAmount          int64           = 0
		dayBaseRewardAmountDe decimal.Decimal = decimal.NewFromInt(dayBaseRewardAmountUsed)
		percentageDe          decimal.Decimal = decimal.NewFromInt(percentage)
	)
	rewardAmount = dayBaseRewardAmountDe.Mul(percentageDe).Div(decimal.NewFromInt(10000)).IntPart()
	return rewardAmount
}
func getOrderRewardAmountBidDeal(dayBaseRewardAmountBidDeal, percentage int64) int64 {
	var (
		rewardAmount          int64           = 0
		dayBaseRewardAmountDe decimal.Decimal = decimal.NewFromInt(dayBaseRewardAmountBidDeal)
		percentageDe          decimal.Decimal = decimal.NewFromInt(percentage)
	)
	rewardAmount = dayBaseRewardAmountDe.Mul(percentageDe).Div(decimal.NewFromInt(10000)).IntPart()
	return rewardAmount
}

func checkOfficialExcludedAddress(address string) bool {
	for _, v := range config.EventOneExtraRewardExcludedOfficialAddress {
		if v == address {
			return true
		}
	}
	return false
}
