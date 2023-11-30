package order_brc20_service

import (
	"encoding/hex"
	"fmt"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
	"strings"
)

// rdex unused pool:20%
// rdex used pool:40%
// rdex bid:40%
func CalAllEventOrder(net, chain string, startBlock, endBlock, bigBlock, startBlockTime, endBlockTime, nowTime int64) (map[string]string, int64, map[string]string, int64, map[string]string, int64) {
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
		orderCalBlockIndex        map[string]int64 = make(map[string]int64)

		coinPriceMap map[string]int64 = make(map[string]int64)

		//endTime int64 = nowTime - 1000*60*60*24*config.EventOneExtraRewardLpUnusedDuration
		timeDis int64 = 1000 * 60 * 60 * 24

		calRdexPoolRewardInfo       map[string]string = make(map[string]string) //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
		calRdexPoolRewardTotalValue int64             = 0

		calRdexPoolExtraRewardInfo       map[string]string = make(map[string]string) //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
		calRdexPoolExtraRewardTotalValue int64             = 0

		calRdexBidDealExtraRewardInfo       map[string]string = make(map[string]string) //{"brc20OrderId":"value:percentage:dealAmount"}
		calRdexBidDealExtraRewardTotalValue int64             = 0
	)

	_ = coinPriceMap
	_ = rewardAmountNoUsed
	_ = timeDis
	_ = totalCoinAmountNoUsed
	_ = totalAmountNoUsed
	_ = orderCoinAmountInfoNoUsed
	_ = orderAmountInfoNoUsed

	allNoUsedEntityRdexPoolOrderList, _ = mongo_service.FindPoolBrc20ModelListByStartTimeAndEndTimeAndNoRemove(net, tick, "", "",
		model.PoolTypeBoth, limit, 0, config.EventOneStartTime, endBlockTime)
	if allNoUsedEntityRdexPoolOrderList != nil && len(allNoUsedEntityRdexPoolOrderList) != 0 {
		major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][no-used]allNoUsedEntityRdexPoolOrderList len:%d", len(allNoUsedEntityRdexPoolOrderList)))
		for _, v := range allNoUsedEntityRdexPoolOrderList {
			if strings.ToLower(v.Tick) != "rdex" {
				continue
			}
			if checkOfficialExcludedAddress(v.Address) {
				continue
			}
			if v.Timestamp < config.EventOneStartTime || v.Timestamp > config.EventOneEndTime {
				continue
			}
			lpTimestamp := v.Timestamp
			lpRemoveTime := v.UpdateTime
			lpStartBlock, _ := getEventBlockStartTimeByTimestamp(net, chain, lpTimestamp)

			if lpStartBlock < config.EventOneStartBlock {
				//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] lpStartBlock[%d] not in event block[%d]\n", v.OrderId, lpStartBlock, config.EventOneStartBlock)
				continue
			}
			if lpStartBlock > endBlock {
				//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] lpStartBlock[%d] not in currrent block[%d]\n", v.OrderId, lpStartBlock, endBlock)
				continue
			}
			calBlockIndex := int64(0)
			if v.DealTime != 0 {
				lpDealStartBlock, _ := getEventBlockStartTimeByTimestamp(net, chain, v.DealTime)
				if lpDealStartBlock < endBlock {
					//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] lpDealStartBlock[%d] not in currrent block[%d]\n", v.OrderId, lpDealStartBlock, endBlock)
					continue
				}
			} else if v.PoolState == model.PoolStateRemove {
				lpRemoveStartBlock, _ := getEventBlockStartTimeByTimestamp(net, chain, lpRemoveTime)
				if lpRemoveStartBlock < endBlock {
					//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] lpRemoveStartBlock[%d] not in currrent block[%d]\n", v.OrderId, lpRemoveStartBlock, endBlock)
					continue
				}
			}
			calBlockIndex = calBlockByStartBlock(lpStartBlock, startBlock)
			if calBlockIndex <= 0 {
				//fmt.Printf("[EVENT][CAL-POOL-BLOCK_USER][no-used] order[%s] calBlockIndex[%d] <= 0 startBlock[%d], lpStartBlock[%s]\n", v.OrderId, calBlockIndex, startBlock, lpStartBlock)
				continue
			}
			if _, ok := orderCalBlockIndex[v.OrderId]; !ok {
				orderCalBlockIndex[v.OrderId] = calBlockIndex
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

			calBlockIndex := int64(0)
			if _, ok := orderCalBlockIndex[orderId]; ok {
				calBlockIndex = orderCalBlockIndex[orderId]
			}
			if calBlockIndex <= 0 {
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
				FromOrderId:         orderEntity.OrderId,
				FromOrderRole:       "",
				FromOrderTotalValue: 0,
				FromOrderOwnValue:   0,
				Address:             orderEntity.CoinAddress,
				TotalValue:          allTotalValueNoUsed,
				OwnValue:            orderTotalValue,
				Percentage:          orderEntity.PercentageExtra,
				RewardAmount:        orderEntity.RewardExtraAmount,
				RewardType:          model.RewardTypeEventOneLpUnusedV2,
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
				major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][no-used]SetRewardRecordModel err:%s", err.Error()))
			}
			major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][no-used]SetRewardRecordModel success [%s]", orderId))

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

				//calStartTime, calEndTime, calDay := GetEventNowCalStartTimeAndEndTime()
				percentageDe := decimal.NewFromInt(percentage)
				fromOrderTotalValue := orderEntity.SellerTotalFee + orderEntity.BuyerTotalFee
				fromOrderTotalValueDe := decimal.NewFromInt(fromOrderTotalValue)

				//seller
				sellerTotalFee := orderEntity.SellerTotalFee
				sellerTotalFeeDe := decimal.NewFromInt(sellerTotalFee)
				sellerFromPercentage := sellerTotalFeeDe.Div(fromOrderTotalValueDe).Mul(decimal.NewFromInt(10000)).IntPart()
				finalSellerPercentage := percentageDe.Mul(decimal.NewFromInt(sellerFromPercentage)).Div(decimal.NewFromInt(10000)).IntPart()
				sellerRewardAmount := getOrderRewardAmountBidDeal(rewardAmount, sellerFromPercentage)
				sellerAddress := orderEntity.SellerAddress
				sellerRecordOrderId := fmt.Sprintf("%s_%s_seller_%s_%d_%s_%d_%d", orderEntity.Net, orderEntity.Tick, sellerAddress, bigBlock, orderEntity.OrderId, orderEntity.CalStartBlock, orderEntity.CalEndBlock)
				sellerRecordOrderId = hex.EncodeToString(tool.SHA256([]byte(sellerRecordOrderId)))

				sellerRewardRecord, _ := mongo_service.FindRewardRecordModelByOrderId(sellerRecordOrderId)
				if sellerRewardRecord == nil {
					sellerRewardRecord = &model.RewardRecordModel{
						Net:                 orderEntity.Net,
						Tick:                orderEntity.Tick,
						OrderId:             sellerRecordOrderId,
						Pair:                fmt.Sprintf("%s-BTC", strings.ToUpper(orderEntity.Tick)),
						FromOrderId:         orderEntity.OrderId,
						FromOrderRole:       "seller",
						FromOrderTotalValue: fromOrderTotalValue,
						FromOrderOwnValue:   sellerTotalFee,
						Address:             sellerAddress,
						TotalValue:          allDealTotalValue,
						OwnValue:            orderDealTotalValue,
						Percentage:          finalSellerPercentage,
						RewardAmount:        sellerRewardAmount,
						RewardType:          model.RewardTypeEventOneBid,
						CalBigBlock:         bigBlock,
						CalDay:              0,
						CalStartTime:        startBlockTime,
						CalEndTime:          endBlockTime,
						CalStartBlock:       startBlock,
						CalEndBlock:         endBlock,
						Version:             1,
						Timestamp:           nowTime,
					}
					_, err = mongo_service.SetRewardRecordModel(sellerRewardRecord)
					if err != nil {
						major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][block-bid]SetRewardRecordModel err:%s", err.Error()))
					}
				}

				//buyer
				buyerTotalFee := orderEntity.BuyerTotalFee
				buyerTotalFeeDe := decimal.NewFromInt(buyerTotalFee)
				buyerFromPercentage := buyerTotalFeeDe.Div(fromOrderTotalValueDe).Mul(decimal.NewFromInt(10000)).IntPart()
				finalBuyerPercentage := percentageDe.Mul(decimal.NewFromInt(buyerFromPercentage)).Div(decimal.NewFromInt(10000)).IntPart()
				buyerRewardAmount := getOrderRewardAmountBidDeal(rewardAmount, buyerFromPercentage)
				buyerAddress := orderEntity.BuyerAddress
				buyerRecordOrderId := fmt.Sprintf("%s_%s_buyer_%s_%d_%s_%d_%d", orderEntity.Net, orderEntity.Tick, buyerAddress, bigBlock, orderEntity.OrderId, orderEntity.CalStartBlock, orderEntity.CalEndBlock)
				buyerRecordOrderId = hex.EncodeToString(tool.SHA256([]byte(buyerRecordOrderId)))

				buyerRewardRecord, _ := mongo_service.FindRewardRecordModelByOrderId(buyerRecordOrderId)
				if buyerRewardRecord == nil {
					buyerRewardRecord = &model.RewardRecordModel{
						Net:                 orderEntity.Net,
						Tick:                orderEntity.Tick,
						OrderId:             buyerRecordOrderId,
						Pair:                fmt.Sprintf("%s-BTC", strings.ToUpper(orderEntity.Tick)),
						FromOrderId:         orderEntity.OrderId,
						FromOrderRole:       "buyer",
						FromOrderTotalValue: fromOrderTotalValue,
						FromOrderOwnValue:   buyerTotalFee,
						Address:             buyerAddress,
						TotalValue:          allDealTotalValue,
						OwnValue:            orderDealTotalValue,
						Percentage:          finalBuyerPercentage,
						RewardAmount:        buyerRewardAmount,
						RewardType:          model.RewardTypeEventOneBid,
						CalBigBlock:         bigBlock,
						CalDay:              0,
						CalStartTime:        startBlockTime,
						CalEndTime:          endBlockTime,
						CalStartBlock:       startBlock,
						CalEndBlock:         endBlock,
						Version:             1,
						Timestamp:           nowTime,
					}
					_, err = mongo_service.SetRewardRecordModel(buyerRewardRecord)
					if err != nil {
						major.Println(fmt.Sprintf("[EVENT][CAL-POOL-BLOCK_USER][block-bid]SetRewardRecordModel err:%s", err.Error()))
					}
				}

				calRdexBidDealExtraRewardInfo[orderId] = fmt.Sprintf("%d:%d:%d:%d:%d", orderDealTotalValue, percentage, orderEntity.Amount, orderEntity.DealTxBlock, orderEntity.OrderType)
			}
			calRdexBidDealExtraRewardTotalValue = allDealTotalValue
		} else {
			major.Println(fmt.Sprintf("[EVENT][CAL-BID-BLOCK_USER][block-bid]SetOrderBlockUserInfoModel success [allDealTotalValue is 0]"))
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

func CheckEventRemainingRewardTotal() bool {
	var (
		hadClaimRewardAmount uint64 = GetEventHadClaimTotal()
	)
	if hadClaimRewardAmount >= uint64(config.EventOneExtraRewardAmount*((config.EventOneEndBlock-config.EventOneStartBlock)/config.EventOneRewardCalCycleBlock)) {
		return false
	}
	return true
}

func GetEventHadClaimTotal() uint64 {
	var (
		net                       string = "livenet"
		rewardTick                string = config.EventOneRewardTick
		entityRewardOrderCountBid *model.PoolRewardOrderCount
		entityRewardOrderCountLp  *model.PoolRewardOrderCount
		hadClaimRewardAmount      uint64 = 0
	)
	entityRewardOrderCountBid, _ = mongo_service.CountOwnPoolRewardOrder(net, rewardTick, "", "", model.RewardTypeEventOneBid)
	if entityRewardOrderCountBid != nil {
		hadClaimRewardAmount = hadClaimRewardAmount + uint64(entityRewardOrderCountBid.RewardCoinAmountTotal)
	}
	entityRewardOrderCountLp, _ = mongo_service.CountOwnPoolRewardOrder(net, rewardTick, "", "", model.RewardTypeEventOneLp)
	if entityRewardOrderCountLp != nil {
		hadClaimRewardAmount = hadClaimRewardAmount + uint64(entityRewardOrderCountLp.RewardCoinAmountTotal)
	}
	return hadClaimRewardAmount
}

func GetEventNowCalStartTimeAndEndTime() (int64, int64, int64) {
	var (
		nowTime                  int64 = tool.MakeTimestamp()
		calDay                   int64 = 0
		calStartTime, calEndTime int64 = 0, 0
		dayDistance              int64 = 1000 * 60 * 60 * 24
		dayCount                 int64 = (config.EventOneEndTime - config.EventOneStartTime) / dayDistance
	)
	for i := int64(0); i <= dayCount; i++ {
		if nowTime >= config.EventOneStartTime+i*dayDistance && nowTime < config.EventOneStartTime+(i+1)*dayDistance {
			calDay = i + 1
			calStartTime = config.EventOneStartTime + i*dayDistance
			calEndTime = config.EventOneStartTime + (i+1)*dayDistance - 1
			break
		}
	}
	return calStartTime, calEndTime, calDay
}

func GetEventNowCalStartTimeAndEndTimeByLpTimestamp(lpTime int64) (int64, int64, int64) {
	var (
		nowTime                  int64 = tool.MakeTimestamp()
		calDay                   int64 = 0
		calStartTime, calEndTime int64 = 0, 0
		dayDistance              int64 = 1000 * 60 * 60 * 24
		dayCount                 int64 = (config.EventOneEndTime - config.EventOneStartTime) / dayDistance
	)
	calDay = (nowTime - lpTime) / dayDistance
	for i := int64(0); i <= dayCount; i++ {
		if nowTime >= config.EventOneStartTime+i*dayDistance && nowTime < config.EventOneStartTime+(i+1)*dayDistance {
			calStartTime = config.EventOneStartTime + i*dayDistance
			calEndTime = config.EventOneStartTime + (i+1)*dayDistance - 1
			break
		}
	}

	return calStartTime, calEndTime, calDay
}

func getEventDayStartTimeByTimestamp(lpTime int64) int64 {
	var (
		lpStartTime int64 = 0
		dayDistance int64 = 1000 * 60 * 60 * 24
		dayCount    int64 = (config.EventOneEndTime - config.EventOneStartTime) / dayDistance
	)
	for i := int64(0); i <= dayCount; i++ {
		if lpTime >= config.EventOneStartTime+i*dayDistance && lpTime < config.EventOneStartTime+(i+1)*dayDistance {
			lpStartTime = config.EventOneStartTime + i*dayDistance
			break
		}
	}
	return lpStartTime
}

func getEventBlockStartTimeByTimestamp(net, chain string, lpTime int64) (int64, int64) {
	var (
		lpStartBlockTime int64 = 0
		lpStartBlock     int64 = 0
		lpBlock          int64 = 0
		lpBlockInfo      *model.BlockInfoModel
		calStartBlock    int64 = config.EventOneStartBlock
		calCycleBlock    int64 = config.EventOneRewardCalCycleBlock
		dayCount         int64 = (config.EventOneEndBlock - config.EventOneStartBlock) / calCycleBlock
	)

	lpBlockInfo, _ = mongo_service.FindNewestHeightBlockInfoModelByBlockTime(net, chain, lpTime/1000)
	if lpBlockInfo == nil {
		return 0, 0
	}
	lpBlock = lpBlockInfo.Height
	for i := int64(0); i <= dayCount; i++ {
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

func calDayByStartTime(startTime1, startTime2 int64) int64 {
	var (
		dayDistance int64 = 1000 * 60 * 60 * 24
		dayCount    int64 = (startTime2 - startTime1) / dayDistance
	)
	return dayCount
}

func calBlockByStartBlock(startBlock1, startBlock2 int64) int64 {
	var (
		cycleBlock int64 = config.EventOneRewardCalCycleBlock
		dayCount   int64 = (startBlock2 - startBlock1) / cycleBlock
	)
	return dayCount
}
