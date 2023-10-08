package task

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/tool"
	"strconv"
	"time"
)

func Fix() {
	var (
		net       string = "livenet"
		nowTime   int64  = tool.MakeTimestamp()
		startTime int64  = 1692677705213
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
			updateTickKline(net, tick, startTime, endTime, model.TimeType15m, "", 0)
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
		updateTickKline(net, tick, startTime, nowTime, model.TimeType15m, "", 0)
	}
}

func jobTickKline15m() {
	jobTickKlineInTime(model.TimeType15m)
}
func jobTickKline1h() {
	jobTickKlineInTime(model.TimeType1h)
}
func jobTickKline4h() {
	jobTickKlineInTime(model.TimeType4h)
}
func jobTickKline1d() {
	jobTickKlineInTime(model.TimeType1d)
}
func jobTickKline1w() {
	jobTickKlineInTime(model.TimeType1w)
}

func jobTickKlineInTime(timeType model.TimeType) {
	var (
		net           string = "livenet"
		nowTime       int64  = tool.MakeTimestamp()
		dis           int64  = 1000 * 60 * 15
		agoTime       int64  = dis
		startTime     int64  = nowTime - agoTime
		entityList    []*model.Brc20TickModel
		tickList      []string
		date          string = ""
		dateTimestamp int64  = 0
		l, _                 = time.LoadLocation("UTC")
		timeFormat    string = "2006-01-02 15:04:05(UTC)"
		suffix        string = ""
	)
	switch timeType {
	case model.TimeType15m:
		agoTime = dis
		timeFormat = "2006-01-02 15:04"
		suffix = ":00(UTC)"
		break
	case model.TimeType1h:
		agoTime = dis * 4
		timeFormat = "2006-01-02 15"
		suffix = ":00:00(UTC)"
		break
	case model.TimeType4h:
		agoTime = dis * 4 * 4
		timeFormat = "2006-01-02 15"
		suffix = ":00:00(UTC)"
		break
	case model.TimeType1d:
		agoTime = dis * 4 * 24
		timeFormat = "2006-01-02"
		suffix = " 00:00:00(UTC)"
		break
	case model.TimeType1w:
		agoTime = dis * 4 * 24 * 7
		timeFormat = "2006-01-02"
		suffix = " 00:00:00(UTC)"
		break
	}
	endDate := time.Unix(nowTime/1000, 0).In(l).Format(timeFormat)
	date = fmt.Sprintf("%s%s", endDate, suffix)
	dateTimestamp = 0

	startTime = nowTime - agoTime

	entityList, _ = mongo_service.FindBrc20TickModelList(net, "", 0, 100)
	for _, v := range entityList {
		tickList = append(tickList, v.Tick)
	}

	for _, tick := range tickList {
		updateTickKline(net, tick, startTime, nowTime, timeType, date, dateTimestamp)
	}
}

func updateTickKline(net, tick string, startTime, endTime int64, timeType model.TimeType, date string, dateTimestamp int64) {
	var (
		limit                            int64 = 5000
		entityList                       []*model.OrderBrc20Model
		openPrice, closePrice, low, high uint64 = 0, 0, 0, 0
		volume                           int64  = 0
		kline                            *model.Brc20TickKlineModel
		newestKline                      *model.Brc20TickKlineModel
		err                              error
	)
	entityList, _ = mongo_service.FindOrderBrc20ModelListByDealTimestamp(net, tick, 0, model.OrderStateFinish,
		limit, startTime, endTime)
	if entityList == nil || int64(len(entityList)) <= 0 {
		newestKline, _ = mongo_service.FindNewestBrc20TickKlineModel(net, tick)
		if newestKline != nil {
			openPrice, _ = strconv.ParseUint(newestKline.Open, 10, 64)
			closePrice, _ = strconv.ParseUint(newestKline.Close, 10, 64)
			if closePrice >= openPrice {
				high = closePrice
				low = openPrice
			} else {
				high = openPrice
				low = closePrice
			}
		}

		kline = &model.Brc20TickKlineModel{
			//TickId:        fmt.Sprintf("%s_%s_%d", net, tick, endTime),
			TickId:        fmt.Sprintf("%s_%s_%d", net, tick, dateTimestamp),
			Net:           net,
			Tick:          tick,
			Open:          strconv.FormatUint(openPrice, 10),
			High:          strconv.FormatUint(high, 10),
			Low:           strconv.FormatUint(low, 10),
			Close:         strconv.FormatUint(closePrice, 10),
			Volume:        volume,
			Date:          date,
			DateTimestamp: dateTimestamp,
			Timestamp:     endTime,
			//TimeType:  model.TimeType15m,
			TimeType: timeType,
		}
		_, err = mongo_service.SetBrc20TickKlineModel(kline)
		if err != nil {
			major.Println(fmt.Sprintf("SetBrc20TickKlineModel err:%s", err.Error()))
			return
		}
		fmt.Printf("[KLINE][%s-%s] open[%d] - %s\n", net, tick, openPrice, tool.MakeDate(endTime))
		return
	}
	openPrice = entityList[0].CoinRatePrice
	closePrice = entityList[len(entityList)-1].CoinRatePrice
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
		TickId:        fmt.Sprintf("%s_%s_%d", net, tick, endTime),
		Net:           net,
		Tick:          tick,
		Open:          strconv.FormatUint(openPrice, 10),
		High:          strconv.FormatUint(high, 10),
		Low:           strconv.FormatUint(low, 10),
		Close:         strconv.FormatUint(closePrice, 10),
		Volume:        volume,
		Date:          date,
		DateTimestamp: dateTimestamp,
		Timestamp:     endTime,
		//TimeType:  model.TimeType15m,
		TimeType: timeType,
	}
	_, err = mongo_service.SetBrc20TickKlineModel(kline)
	if err != nil {
		major.Println(fmt.Sprintf("SetBrc20TickKlineModel err:%s", err.Error()))
		return
	}
	fmt.Printf("[KLINE][%s-%s] open[%d] - %s\n", net, tick, openPrice, tool.MakeDate(endTime))
}

func jobTickRecentlyInfo() {
	var (
		net        string = "livenet"
		entityList []*model.Brc20TickModel
		tickList   []string
	)
	entityList, _ = mongo_service.FindBrc20TickModelList(net, "", 0, 100)
	for _, v := range entityList {
		tickList = append(tickList, v.Tick)
	}

	for _, tick := range tickList {
		order_brc20_service.UpdateTickRecentlyInfo(net, tick)
	}
}
