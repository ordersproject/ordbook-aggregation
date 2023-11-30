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

func JobForCalEventOrder() {
	var (
		net                                 string = "livenet"
		chain                               string = "btc"
		nowTime                             int64  = tool.MakeTimestamp()
		processingBigBlock, currentBigBlock int64  = 0, 0
	)
	processingBigBlock = getCurrentProcessingEventBigBlock()
	currentBigBlock = getCurrentEventBigBlock()

	fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d\n", processingBigBlock, currentBigBlock)

	for i := processingBigBlock + 1; i < currentBigBlock; i++ {
		fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d\n", processingBigBlock, currentBigBlock, i)
		if i <= 0 {
			continue
		}

		startBlock, endBlock := getEventStartBlockAndEndBlockByBigBlock(i + 1)
		fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d, startBlock:%d, endBlock:%d\n", processingBigBlock, currentBigBlock, i, startBlock, endBlock)
		if startBlock == 0 || endBlock == 0 {
			continue
		}
		if endBlock >= config.EventOneEndBlock {
			fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d, startBlock:%d, endBlock:%d, event finish\n", processingBigBlock, currentBigBlock, startBlock, endBlock, i)
			continue
		}

		startBlockTime, endBlockTime := getBlockTime(net, chain, startBlock), getBlockTime(net, chain, endBlock)
		if startBlockTime == 0 || endBlockTime == 0 {
			major.Println(fmt.Sprintf("[JOP][CalEvenOrder] getBlockTime err, startBlockTime:%d, endBlockTime:%d", startBlockTime, endBlockTime))
			continue
		}
		fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d, startBlock:%d, endBlock:%d, startBlockTime:%d[%s], endBlockTime:%d[%s]\n",
			processingBigBlock, currentBigBlock, i, startBlock, endBlock, startBlockTime, tool.MakeDate(startBlockTime), endBlockTime, tool.MakeDate(endBlockTime))

		calPoolRewardInfo, calPoolRewardTotalValue,
			calPoolExtraRewardInfo, calPoolExtraRewardTotalValue,
			calEventRdexBidDealExtraRewardInfo, calEventBidDealExtraRewardTotalValue := order_brc20_service.CalAllEventOrder(net, chain, startBlock, endBlock, i, startBlockTime, endBlockTime, nowTime)
		order_brc20_service.UpdatePoolBlockInfo(GetCurrentEventBigBlock(startBlock), startBlock, endBlock, (endBlock-startBlock)+1, nowTime,
			calPoolRewardInfo, calPoolRewardTotalValue,
			calPoolExtraRewardInfo, calPoolExtraRewardTotalValue,
			calEventRdexBidDealExtraRewardInfo, calEventBidDealExtraRewardTotalValue,
			model.CalTypeEventOne)
	}
}

//func JobForCalEventOrderByTime() {
//	var (
//		net                                 string = "livenet"
//		nowTime                             int64  = tool.MakeTimestamp()
//		processingBigBlock, currentBigBlock int64  = 0, 0
//	)
//	processingBigBlock = getCurrentProcessingEventBigBlock()
//	currentBigBlock = getCurrentEventBigBlock()
//
//	fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d\n", processingBigBlock, currentBigBlock)
//
//	for i := processingBigBlock + 1; i < currentBigBlock; i++ {
//		fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d\n", processingBigBlock, currentBigBlock, i)
//		if i <= 0 {
//			continue
//		}
//
//		startBlock, endBlock := getEventStartBlockAndEndBlockByBigBlock(i + 1)
//		fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d, startBlock:%d, endBlock:%d\n", processingBigBlock, currentBigBlock, i, startBlock, endBlock)
//		if startBlock == 0 || endBlock == 0 {
//			continue
//		}
//		if endBlock >= config.EventOneEndBlock {
//			fmt.Printf("[JOP][CalEvenOrder] processingBigBlock:%d, currentBigBlock:%d, bigBlock:%d, startBlock:%d, endBlock:%d, event finish\n", processingBigBlock, currentBigBlock, startBlock, endBlock, i)
//			continue
//		}
//		calPoolExtraRewardInfo, calPoolExtraRewardTotalValue := order_brc20_service.CalAllEventOrderForUnusedPool(net, startBlock, endBlock, i, nowTime)
//		order_brc20_service.UpdatePoolBlockInfo(GetCurrentEventBigBlock(startBlock), startBlock, endBlock, (endBlock-startBlock)+1, nowTime,
//			calPoolRewardInfo, calPoolRewardTotalValue,
//			calPoolExtraRewardInfo, calPoolExtraRewardTotalValue,
//			calEventRdexBidDealExtraRewardInfo, calEventBidDealExtraRewardTotalValue,
//			model.CalTypeEventOne)
//	}
//}

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

func GetCurrentEventBigBlock(startBlock int64) int64 {
	var (
		bigBlock int64 = 0
	)
	if startBlock <= config.EventOneStartBlock {
		bigBlock = 0
	} else {
		bigBlock = (startBlock - config.EventOneStartBlock) / config.EventOneRewardCalCycleBlock
	}
	return bigBlock
}

func getEventStartBlockAndEndBlockByBigBlock(bigBlock int64) (int64, int64) {
	var (
		startBlock, endBlock int64 = 0, 0
	)
	startBlock = config.EventOneStartBlock + (bigBlock-1)*config.EventOneRewardCalCycleBlock
	endBlock = config.EventOneStartBlock + bigBlock*config.EventOneRewardCalCycleBlock - 1
	return startBlock, endBlock
}

func getBlockTime(net, chain string, blockHeight int64) int64 {
	var (
		blockInfo *model.BlockInfoModel
		blockId   string = fmt.Sprintf("%s_%s_%d", net, chain, blockHeight)
	)
	blockInfo, _ = mongo_service.FindBlockInfoModelByBlockId(blockId)
	if blockInfo == nil {
		return 0
	}
	return blockInfo.BlockTime * 1000
}
