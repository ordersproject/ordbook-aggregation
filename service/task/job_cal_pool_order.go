package task

import (
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/node"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/tool"
)

func jobForCalPoolOrder() {
	var (
		net                                 string = "livenet"
		chain                               string = "btc"
		nowTime                             int64  = tool.MakeTimestamp()
		processingBigBlock, currentBigBlock int64  = 0, 0
	)
	processingBigBlock = getCurrentProcessingBigBlock()
	currentBigBlock = getCurrentBigBlock()

	fmt.Printf("[JOP][CalPoolOrder] processingBigBlock:%d, currentBigBlock:%d\n", processingBigBlock, currentBigBlock)

	for i := processingBigBlock + 1; i < currentBigBlock; i++ {
		fmt.Printf("[JOP][CalPoolOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d\n", processingBigBlock, currentBigBlock, i)
		if i <= 0 {
			continue
		}
		startBlock, endBlock := getStartBlockAndEndBlockByBigBlock(i + 1)
		fmt.Printf("[JOP][CalPoolOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d, startBlock:%d, endBlock:%d\n", processingBigBlock, currentBigBlock, i, startBlock, endBlock)
		if startBlock == 0 || endBlock == 0 {
			continue
		}
		startBlockTime, endBlockTime := getBlockTime(net, chain, startBlock), getBlockTime(net, chain, endBlock)
		if startBlockTime == 0 || endBlockTime == 0 {
			major.Println(fmt.Sprintf("[JOP][CalPoolOrder] getBlockTime err, startBlockTime:%d, endBlockTime:%d", startBlockTime, endBlockTime))
			continue
		}
		fmt.Printf("[JOP][CalPoolOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d, startBlock:%d, endBlock:%d, startBlockTime:%d[%s], endBlockTime:%d[%s]\n",
			processingBigBlock, currentBigBlock, i, startBlock, endBlock, startBlockTime, tool.MakeDate(startBlockTime), endBlockTime, tool.MakeDate(endBlockTime))

		calPoolRewardInfo, calPoolRewardTotalValue, calPoolExtraRewardInfo, calPoolExtraRewardTotalValue := order_brc20_service.CalAllPoolOrderV2(net, chain, startBlock, endBlock, i, startBlockTime, endBlockTime, nowTime)
		order_brc20_service.UpdatePoolBlockInfo(order_brc20_service.GetCurrentBigBlock(startBlock), startBlock, endBlock, (endBlock-startBlock)+1, nowTime,
			calPoolRewardInfo, calPoolRewardTotalValue, calPoolExtraRewardInfo, calPoolExtraRewardTotalValue,
			nil, 0,
			model.CalTypePlatform)
	}
}

func getCurrentProcessingBigBlock() int64 {
	var (
		blockInfo  *model.PoolBlockInfoModel
		cycleBlock int64 = config.PlatformRewardCalCycleBlock
		err        error
	)

	blockInfo, err = mongo_service.FindNewestPoolBlockInfoModelByCycleBlockAndCalType(cycleBlock, model.CalTypePlatform)
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
