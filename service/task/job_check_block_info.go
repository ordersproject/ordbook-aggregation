package task

import (
	"fmt"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/node"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
)

func jobForCheckBlockInfo() {
	var (
		net, chain         string = "livenet", "btc"
		currentBlockHeight uint64 = 0
		processBlockHeight int64  = config.PlatformRewardCalStartBlock
		err                error
		newestBlockInfo    *model.BlockInfoModel
	)

	currentBlockHeight, err = node.CurrentBlockHeight(net)
	if err != nil {
		return
	}
	newestBlockInfo, _ = mongo_service.FindNewestHeightBlockInfoModel(net, chain)
	if newestBlockInfo != nil {
		processBlockHeight = newestBlockInfo.Height
	}
	if processBlockHeight < config.PlatformRewardCalStartBlock {
		processBlockHeight = config.PlatformRewardCalStartBlock
	}
	syncBlockInfo(net, chain, config.PlatformRewardCalStartBlock)
	for i := processBlockHeight + 1; i <= int64(currentBlockHeight); i++ {
		syncBlockInfo(net, chain, i)
	}
}

func syncBlockInfo(net, chain string, blockHeight int64) {
	var (
		blockInfo *model.BlockInfoModel
		block     *node.Block
		err       error
		blockId   string = "" //net_chain_height
	)
	block, err = node.GetBlockInfo(net, blockHeight)
	if err != nil {
		major.Println(fmt.Sprintf("[JOP][syncBlockInfo]err:%s\n", err.Error()))
		return
	}
	blockId = fmt.Sprintf("%s_%s_%d", net, chain, block.Height)
	blockInfo, _ = mongo_service.FindBlockInfoModelByBlockId(blockId)
	if blockInfo != nil {
		return
	}
	blockInfo = &model.BlockInfoModel{
		Id:           0,
		BlockId:      blockId,
		Net:          net,
		Chain:        chain,
		Height:       int64(block.Height),
		Hash:         block.Hash,
		BlockTime:    int64(block.Time),
		BlockTimeStr: tool.MakeDate(int64(block.Time) * 1000),
		Timestamp:    tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetBlockInfoModel(blockInfo)
	if err != nil {
		major.Println(fmt.Sprintf("[JOP][syncBlockInfo]err:%s\n", err.Error()))
		return
	}
	fmt.Printf("[JOP][syncBlockInfo]blockHeight-[%d][%s] \n", blockInfo.Height, blockInfo.BlockTimeStr)
}
