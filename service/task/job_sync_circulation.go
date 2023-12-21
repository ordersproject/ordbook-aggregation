package task

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/common_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"strconv"
)

func JobForSyncCirculation() {
	syncCirculation("livenet", "rdex")
}

func syncCirculation(net, tick string) {
	var (
		cirAddresses     []string = common_service.GetCirculationAddress(net)
		totalAmount      uint64   = 0
		totalSupply      uint64   = 100000000
		circulation      uint64   = 0
		orderCirculation *model.OrderCirculationModel
		err              error
	)

	for _, address := range cirAddresses {
		//tokenInfoResp, _ := own_service.GetBrc20Tokens(address, tick, "0", "100")
		//if tokenInfoResp != nil && tokenInfoResp.List != nil {
		//	for _, tokenInfo := range tokenInfoResp.List {
		//		if strings.ToLower(tokenInfo.Ticker) == tick {
		//			amountStr := tokenInfo.OverallBalance
		//			amount, _ := strconv.ParseUint(amountStr, 10, 64)
		//			totalAmount = totalAmount + amount
		//		}
		//	}
		//}

		brc20BalanceResult, err := oklink_service.GetAddressBrc20BalanceResult(address, tick, 1, 50)
		if err != nil {
			continue
		}
		fmt.Printf("brc20BalanceResult:%+v\n", brc20BalanceResult)
		amount, _ := strconv.ParseUint(brc20BalanceResult.Balance, 10, 64)
		totalAmount = totalAmount + amount
	}

	circulation = totalSupply - totalAmount
	orderCirculation, _ = mongo_service.FindOrderCirculationModelByTick(net, tick)
	if orderCirculation == nil {
		orderCirculation = &model.OrderCirculationModel{
			Net:         net,
			Tick:        tick,
			TotalSupply: totalSupply,
		}
	}
	if circulation < orderCirculation.CirculationSupply || orderCirculation.CirculationSupply == 0 {
		orderCirculation.CirculationSupply = circulation
	}

	_, err = mongo_service.SetOrderCirculationModel(orderCirculation)
	if err != nil {
		fmt.Printf("SetOrderCirculationModel err:%v\n", err)
		return
	}
	major.Println(fmt.Sprintf("[JOB]SyncCirculation %s %s %d", net, tick, circulation))

}
