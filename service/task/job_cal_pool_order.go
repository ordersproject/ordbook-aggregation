package task

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/config"
	"ordbook-aggregation/model"
	"ordbook-aggregation/node"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/order_brc20_service"
)

func jobForCalPoolOrder() {
	var (
		net                                 string = "livenet"
		nowTime                             int64  = 0
		processingBigBlock, currentBigBlock int64  = 0, 0
	)
	processingBigBlock = getCurrentProcessingBigBlock()
	currentBigBlock = getCurrentBigBlock()

	for i := processingBigBlock; i <= currentBigBlock; i++ {
		if i <= 0 {
			continue
		}
		startBlock, endBlock := getStartBlockAndEndBlockByBigBlock(i)
		order_brc20_service.CalAllPoolOrder(net, startBlock, endBlock, nowTime)
	}

}

func getCurrentProcessingBigBlock() int64 {
	var (
		blockInfo  *model.PoolBlockInfoModel
		cycleBlock int64 = config.PlatformRewardCalCycleBlock
		err        error
	)

	blockInfo, err = mongo_service.FindNewestPoolBlockInfoModelByCycleBlock(cycleBlock)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return -1
	}
	if blockInfo == nil {
		return 0
	}
	return blockInfo.BigBlock
}

func getCurrentBigBlock() int64 {
	var (
		net                string = "livenet"
		currentBlockHeight uint64 = 0
		err                error
	)
	currentBlockHeight, err = node.CurrentBlockHeight(net)
	if err != nil {
		return -1
	}
	return order_brc20_service.GetCurrentBigBlock(int64(currentBlockHeight))
}

func getStartBlockAndEndBlockByBigBlock(bigBlock int64) (int64, int64) {
	var (
		startBlock, endBlock int64 = 0, 0
	)
	startBlock = config.PlatformRewardCalStartBlock + (bigBlock-1)*config.PlatformRewardCalCycleBlock
	endBlock = config.PlatformRewardCalStartBlock + bigBlock*config.PlatformRewardCalCycleBlock - 1
	return startBlock, endBlock
}
