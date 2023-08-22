package task

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
	"strconv"
)

func Fix() {
	var (
		net       string = "livenet"
		nowTime   int64  = tool.MakeTimestamp()
		startTime int64  = 1688112000000
		disTime   int64  = 1000 * 60 * 15

		entityList []*model.Brc20TickModel
		tickList   []string
	)
	entityList, _ = mongo_service.FindBrc20TickModelList(net, "", 0, 100)
	for _, v := range entityList {
		tickList = append(tickList, v.Tick)
	}

	for _, tick := range tickList {
		i := 1
		for {
			endTime := startTime + disTime*int64(i)
			if endTime >= nowTime {
				break
			}
			updateTickKline(net, tick, startTime, endTime)
			i++
		}
	}
}

func jobTickKline() {
	var (
		net        string = "livenet"
		nowTime    int64  = tool.MakeTimestamp()
		agoTime    int64  = 1000 * 60 * 15
		startTime  int64  = nowTime - agoTime
		entityList []*model.Brc20TickModel
		tickList   []string
	)
	entityList, _ = mongo_service.FindBrc20TickModelList(net, "", 0, 100)
	for _, v := range entityList {
		tickList = append(tickList, v.Tick)
	}

	for _, tick := range tickList {
		updateTickKline(net, tick, startTime, nowTime)
	}
}

func updateTickKline(net, tick string, startTime, endTime int64) {
	var (
		limit                      int64 = 5000
		entityList                 []*model.OrderBrc20Model
		open, closePice, low, high uint64 = 0, 0, 0, 0
		volume                     int64  = 0
		kline                      *model.Brc20TickKlineModel
		newestKline                *model.Brc20TickKlineModel
		err                        error
	)
	entityList, _ = mongo_service.FindOrderBrc20ModelListByTimestamp(net, tick, 0, model.OrderStateFinish,
		limit, startTime, endTime)
	if entityList == nil || int64(len(entityList)) <= 0 {
		newestKline, _ = mongo_service.FindNewestBrc20TickKlineModel(net, tick)
		if newestKline != nil {
			open, _ = strconv.ParseUint(newestKline.Open, 10, 64)
			closePice, _ = strconv.ParseUint(newestKline.Close, 10, 64)
			if closePice >= open {
				high = closePice
				low = open
			} else {
				high = open
				low = closePice
			}
		}

		kline = &model.Brc20TickKlineModel{
			TickId:    fmt.Sprintf("%s_%s_%d", net, tick, endTime),
			Net:       net,
			Tick:      tick,
			Open:      strconv.FormatUint(open, 10),
			High:      strconv.FormatUint(high, 10),
			Low:       strconv.FormatUint(low, 10),
			Close:     strconv.FormatUint(closePice, 10),
			Volume:    volume,
			Timestamp: endTime,
		}
		if err != nil {
			major.Println(fmt.Sprintf("SetBrc20TickKlineModel err:%s", err.Error()))
			return
		}
		fmt.Printf("[KLINE][%s-%s] open[%d] - %s\n", net, tick, open, tool.MakeDate(endTime))
		return
	}
	open = entityList[0].CoinRatePrice
	closePice = entityList[len(entityList)-1].CoinRatePrice
	for _, v := range entityList {
		volume++
		if low == 0 || low > v.CoinRatePrice {
			low = v.CoinRatePrice
		}
		if high == 0 || high < v.CoinRatePrice {
			high = v.CoinRatePrice
		}
	}

	kline = &model.Brc20TickKlineModel{
		TickId:    fmt.Sprintf("%s_%s_%d", net, tick, endTime),
		Net:       net,
		Tick:      tick,
		Open:      strconv.FormatUint(open, 10),
		High:      strconv.FormatUint(high, 10),
		Low:       strconv.FormatUint(low, 10),
		Close:     strconv.FormatUint(closePice, 10),
		Volume:    volume,
		Timestamp: endTime,
	}
	_, err = mongo_service.SetBrc20TickKlineModel(kline)
	if err != nil {
		major.Println(fmt.Sprintf("SetBrc20TickKlineModel err:%s", err.Error()))
		return
	}
	fmt.Printf("[KLINE][%s-%s] open[%d] - %s\n", net, tick, open, tool.MakeDate(endTime))
}
