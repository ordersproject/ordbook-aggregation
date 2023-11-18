package task

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/order_brc20_service"
)

func jobForCheckClaimBlock() {
	var (
		net           string = "livenet"
		poolOrderList []*model.PoolBrc20Model
		limit         int64 = 1000
	)

	poolOrderList, _ = mongo_service.FindPoolBrc20ModelListByClaimTime(net, "", "", "", model.PoolStateClaim,
		limit, 0, model.ClaimTxBlockStateUnconfirmed)

	if poolOrderList != nil && len(poolOrderList) != 0 {
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

			block := order_brc20_service.GetTxBlock(v.ClaimTx)
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

func jobForCheckPoolUsedDealTxBlock() {
	var (
		net           string = "livenet"
		poolOrderList []*model.PoolBrc20Model
		limit         int64 = 1000
	)

	poolOrderList, _ = mongo_service.FindPoolBrc20ModelListByDealTime(net, "", "", "", model.PoolStateUsed,
		limit, 0, model.ClaimTxBlockStateUnconfirmed)

	if poolOrderList != nil && len(poolOrderList) != 0 {
		for _, v := range poolOrderList {
			if v.DealCoinTxBlock != 0 {
				continue
			}
			if v.DealCoinTxBlockState != model.ClaimTxBlockStateUnconfirmed {
				continue
			}
			if v.DealCoinTx == "" {
				continue
			}

			block := order_brc20_service.GetTxBlock(v.DealCoinTx)
			if block == 0 {
				continue
			}
			v.DealCoinTxBlock = block
			v.DealCoinTxBlockState = model.ClaimTxBlockStateConfirmed
			err := mongo_service.SetPoolBrc20ModelForDealBlock(v)
			if err != nil {
				major.Println(fmt.Sprintf("[JOP-DEAL-BLOCK] SetPoolBrc20ModelForDealBlock err:%s", err.Error()))
				continue
			}
			major.Println(fmt.Sprintf("[JOP-DEAL-BLOCK] SetPoolBrc20ModelForDealBlock success [%s]", v.OrderId))
		}
	}
}
