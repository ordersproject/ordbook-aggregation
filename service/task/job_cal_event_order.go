package task

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/config"
	"ordbook-aggregation/model"
	"ordbook-aggregation/node"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/order_brc20_service"
)

func jobForCalEventOrder() {
	var (
		net                                 string = "livenet"
		nowTime                             int64  = 0
		processingBigBlock, currentBigBlock int64  = 0, 0
	)
	processingBigBlock = getCurrentProcessingEventBigBlock()
	currentBigBlock = getCurrentEventBigBlock()

	fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d\n", processingBigBlock, currentBigBlock)

	for i := processingBigBlock; i <= currentBigBlock; i++ {
		fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d\n", processingBigBlock, currentBigBlock, i)
		if i <= 0 {
			continue
		}
		if i >= config.EventOneEndBlock {
			fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d, event finish\n", processingBigBlock, currentBigBlock, i)
			continue
		}
		startBlock, endBlock := getEventStartBlockAndEndBlockByBigBlock(i)
		if startBlock == 0 || endBlock == 0 {
			continue
		}
		calPoolRewardInfo, calPoolRewardTotalValue,
			calPoolExtraRewardInfo, calPoolExtraRewardTotalValue,
			calEventRdexBidDealExtraRewardInfo, calEventBidDealExtraRewardTotalValue := order_brc20_service.CalAllEventOrder(net, startBlock, endBlock, nowTime)
		order_brc20_service.UpdatePoolBlockInfo(startBlock, endBlock, (endBlock-startBlock)+1, nowTime,
			calPoolRewardInfo, calPoolRewardTotalValue,
			calPoolExtraRewardInfo, calPoolExtraRewardTotalValue,
			calEventRdexBidDealExtraRewardInfo, calEventBidDealExtraRewardTotalValue,
			model.CalTypeEventOne)
	}
}

func getCurrentProcessingEventBigBlock() int64 {
	var (
		blockInfo  *model.PoolBlockInfoModel
		cycleBlock int64 = config.EventOneRewardCalCycleBlock
		err        error
	)

	blockInfo, err = mongo_service.FindNewestPoolBlockInfoModelByCycleBlockAndCalType(cycleBlock, model.CalTypeEventOne)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return -1
	}
	if blockInfo == nil {
		return 0
	}
	return blockInfo.BigBlock
}

func getCurrentEventBigBlock() int64 {
	var (
		net                string = "livenet"
		currentBlockHeight uint64 = 0
		err                error
	)
	currentBlockHeight, err = node.CurrentBlockHeight(net)
	if err != nil {
		return -1
	}
	var (
		bigBlock int64 = 0
	)
	if currentBlockHeight <= uint64(config.EventOneStartBlock) {
		bigBlock = 0
	} else {
		bigBlock = (int64(currentBlockHeight) - int64(config.EventOneStartBlock)) / config.EventOneRewardCalCycleBlock
	}
	return bigBlock
	//return order_brc20_service.GetCurrentBigBlock(int64(currentBlockHeight))
}

func getEventStartBlockAndEndBlockByBigBlock(bigBlock int64) (int64, int64) {
	var (
		startBlock, endBlock int64 = 0, 0
	)
	startBlock = config.EventOneStartBlock + (bigBlock-1)*config.EventOneRewardCalCycleBlock
	endBlock = config.EventOneStartBlock + bigBlock*config.EventOneRewardCalCycleBlock - 1
	return startBlock, endBlock
}
