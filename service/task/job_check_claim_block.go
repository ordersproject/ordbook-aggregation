package task

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"strconv"
)

func jobForCheckClaimBlock() {
	var (
		net           string = "livenet"
		poolOrderList []*model.PoolBrc20Model
		limit         int64 = 1000
	)

	poolOrderList, _ = mongo_service.FindPoolBrc20ModelListByClaimTime(net, "", "", "", model.PoolStateClaim,
		limit, 1, model.ClaimTxBlockStateUnconfirmed)

	if poolOrderList == nil || len(poolOrderList) == 0 {
		for _, v := range poolOrderList {
			if v.ClaimTxBlock != 0 {
				continue
			}
			if v.ClaimTxBlockState != model.ClaimTxBlockStateUnconfirmed {
				continue
			}
			if v.ClaimTx == "" {
				continue
			}

			block := getClaimTxBlock(v.ClaimTx)
			if block == 0 {
				continue
			}
			v.ClaimTxBlock = block
			v.ClaimTxBlockState = model.ClaimTxBlockStateConfirmed
			err := mongo_service.SetPoolBrc20ModelForBlock(v)
			if err != nil {
				major.Println(fmt.Sprintf("[JOP-CLAIM-BLOCK] SetPoolBrc20ModelForBlock err:%s", err.Error()))
				continue
			}
			major.Println(fmt.Sprintf("[JOP-CLAIM-BLOCK] SetPoolBrc20ModelForBlock success [%s]", v.OrderId))
		}
	}
}

func getClaimTxBlock(claimTx string) int64 {
	var (
		blockHeight int64 = 0
	)

	tx, err := oklink_service.GetTxDetail(claimTx)
	if err != nil {
		return 0
	}
	blockHeight, _ = strconv.ParseInt(tx.Height, 10, 64)
	return blockHeight
}
