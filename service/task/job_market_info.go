package task

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
)

func jobMarketInfo() {
	updateMarketInfo()
}

func updateMarketInfo() {
	var (
		net        string = "livenet"
		limit      int64  = 5000
		entityList []*model.OrderBrc20Model
		err        error

		entity               *model.OrderBrc20MarketInfoModel
		startTime, endTime   int64 = tool.GetYesterday0Time(), tool.GetYesterday24Time()
		date                       = tool.MakeDateV3(endTime)
		askVolume, bidVolume int64 = 0, 0
		askFees, bidFees     int64 = 0, 0
	)
	entityList, _ = mongo_service.FindOrderBrc20ModelListByDealTimestamp(net, "", 0, model.OrderStateFinish,
		limit, startTime, endTime)
	for _, v := range entityList {
		if v.OrderType == model.OrderTypeSell {
			askVolume++
		} else {
			bidVolume++
			bidFees = bidFees + int64(v.Fee)
		}
	}

	entity = &model.OrderBrc20MarketInfoModel{
		NetDate:   fmt.Sprintf("%s_%s", net, date),
		Net:       net,
		Date:      date,
		AskVolume: askVolume,
		BidVolume: bidVolume,
		AskFees:   askFees,
		BidFees:   bidFees,
		Between:   fmt.Sprintf("%s_%s", tool.MakeDate(startTime), tool.MakeDate(endTime)),
		Timestamp: tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetOrderBrc20MarketInfoModel(entity)
	if err != nil {
		major.Println(fmt.Sprintf("SetOrderBrc20MarketInfoModel err:%s", err.Error()))
		return
	}
	major.Println(fmt.Sprintf("[JOB]Update for Marker info success - %s", tool.MakeDate(tool.MakeTimestamp())))
}
