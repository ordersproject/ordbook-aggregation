package mongo_service

import (
	"context"
	"errors"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/model"
)

func FindBrc20TickModelByPair(net, pair string, version int) (*model.Brc20TickModel, error) {
	collection, err := model.Brc20TickModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"pair", pair},
		{"version", version},
		//{"state", model.STATE_EXIST},
	}
	//if version != 0 {
	//	queryBson = append(queryBson, bson.E{"version", version})
	//}
	entity := &model.Brc20TickModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createBrc20TickModel(brc20Tick *model.Brc20TickModel) (*model.Brc20TickModel, error) {
	collection, err := model.Brc20TickModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	//CreateUniqueIndex(collection, "pair")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "volume")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "version")

	entity := &model.Brc20TickModel{
		Id:                  util.GetUUIDInt64(),
		Net:                 brc20Tick.Net,
		Tick:                brc20Tick.Tick,
		Pair:                brc20Tick.Pair,
		Buy:                 brc20Tick.Buy,
		Sell:                brc20Tick.Sell,
		Low:                 brc20Tick.Low,
		High:                brc20Tick.High,
		Open:                brc20Tick.Open,
		Last:                brc20Tick.Last,
		LastTop:             brc20Tick.LastTop,
		LastTotal:           brc20Tick.LastTotal,
		Volume:              brc20Tick.Volume,
		Amount:              brc20Tick.Amount,
		Vol:                 brc20Tick.Vol,
		AvgPrice:            brc20Tick.AvgPrice,
		QuoteSymbol:         brc20Tick.QuoteSymbol,
		PriceChangePercent:  brc20Tick.PriceChangePercent,
		CoinPrice:           brc20Tick.CoinPrice,
		CoinPriceDecimalNum: brc20Tick.CoinPriceDecimalNum,
		Version:             brc20Tick.Version,
		CreateTime:          util.Time(),
		State:               model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetBrc20TickModel(brc20Tick *model.Brc20TickModel) (*model.Brc20TickModel, error) {
	entity, err := FindBrc20TickModelByPair(brc20Tick.Net, brc20Tick.Pair, brc20Tick.Version)
	if err == nil && entity != nil {
		collection, err := model.Brc20TickModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"net", brc20Tick.Net},
			{"pair", brc20Tick.Pair},
			{"version", brc20Tick.Version},
			//{"state", model.STATE_EXIST},
		}
		//if brc20Tick.Version != 0 {
		//	filter = append(filter, bson.E{"version", brc20Tick.Version})
		//}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: brc20Tick.Net})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: brc20Tick.Tick})
		bsonData = append(bsonData, bson.E{Key: "pair", Value: brc20Tick.Pair})
		bsonData = append(bsonData, bson.E{Key: "buy", Value: brc20Tick.Buy})
		bsonData = append(bsonData, bson.E{Key: "sell", Value: brc20Tick.Sell})
		bsonData = append(bsonData, bson.E{Key: "low", Value: brc20Tick.Low})
		bsonData = append(bsonData, bson.E{Key: "high", Value: brc20Tick.High})
		bsonData = append(bsonData, bson.E{Key: "open", Value: brc20Tick.Open})
		bsonData = append(bsonData, bson.E{Key: "last", Value: brc20Tick.Last})
		bsonData = append(bsonData, bson.E{Key: "lastTop", Value: brc20Tick.LastTop})
		bsonData = append(bsonData, bson.E{Key: "lastTotal", Value: brc20Tick.LastTotal})
		bsonData = append(bsonData, bson.E{Key: "volume", Value: brc20Tick.Volume})
		bsonData = append(bsonData, bson.E{Key: "amount", Value: brc20Tick.Amount})
		bsonData = append(bsonData, bson.E{Key: "vol", Value: brc20Tick.Vol})
		bsonData = append(bsonData, bson.E{Key: "avgPrice", Value: brc20Tick.AvgPrice})
		bsonData = append(bsonData, bson.E{Key: "quoteSymbol", Value: brc20Tick.QuoteSymbol})
		bsonData = append(bsonData, bson.E{Key: "coinPrice", Value: brc20Tick.CoinPrice})
		bsonData = append(bsonData, bson.E{Key: "coinPriceDecimalNum", Value: brc20Tick.CoinPriceDecimalNum})
		bsonData = append(bsonData, bson.E{Key: "version", Value: brc20Tick.Version})
		bsonData = append(bsonData, bson.E{Key: "priceChangePercent", Value: brc20Tick.PriceChangePercent})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return brc20Tick, nil
	} else {
		return createBrc20TickModel(brc20Tick)
	}
}

func CountBrc20TickModelList(net string) (int64, error) {
	collection, err := model.Brc20TickModel{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state": model.STATE_EXIST,
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindBrc20TickModelList(net, tick string, skip, limit int64) ([]*model.Brc20TickModel, error) {
	collection, err := model.Brc20TickModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":   net,
		"state": model.STATE_EXIST,
	}

	if tick != "" {
		find["tick"] = tick
	}

	models := make([]*model.Brc20TickModel, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(0)
	sort := options.Find().SetSort(bson.M{"updateTime": -1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.Brc20TickModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get Brc20TickModel Error")
	}
	return models, nil
}

func FindBrc20TickModelVersionList(net, tick string, skip, limit, version int64) ([]*model.Brc20TickModel, error) {
	collection, err := model.Brc20TickModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":     net,
		"version": version,
		"state":   model.STATE_EXIST,
	}

	if tick != "" {
		find["tick"] = tick
	}

	models := make([]*model.Brc20TickModel, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(0)
	sort := options.Find().SetSort(bson.M{"updateTime": -1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.Brc20TickModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get Brc20TickModel Error")
	}
	return models, nil
}

func FindBrc20TickInfoModelByTick(net, tick string) (*model.Brc20TickInfoModel, error) {
	collection, err := model.Brc20TickInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"tick", tick},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.Brc20TickInfoModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createBrc20TickInfoModel(brc20Tick *model.Brc20TickInfoModel) (*model.Brc20TickInfoModel, error) {
	collection, err := model.Brc20TickInfoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "tick")
	CreateIndex(collection, "net")
	CreateIndex(collection, "timestamp")

	entity := &model.Brc20TickInfoModel{
		Id:             util.GetUUIDInt64(),
		Net:            brc20Tick.Net,
		Tick:           brc20Tick.Tick,
		Name:           brc20Tick.Name,
		Decimal:        brc20Tick.Decimal,
		Supply:         brc20Tick.Supply,
		Icon:           brc20Tick.Icon,
		DefaultLimit:   brc20Tick.DefaultLimit,
		Deployer:       brc20Tick.Deployer,
		DeployTime:     brc20Tick.DeployTime,
		DeployContract: brc20Tick.DeployContract,
		Description:    brc20Tick.Description,
		CreateTime:     util.Time(),
		State:          model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetBrc20TickInfoModel(brc20Tick *model.Brc20TickInfoModel) (*model.Brc20TickInfoModel, error) {
	entity, err := FindBrc20TickInfoModelByTick(brc20Tick.Net, brc20Tick.Tick)
	if err == nil && entity != nil {
		collection, err := model.Brc20TickInfoModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"net", brc20Tick.Net},
			{"tick", brc20Tick.Tick},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: brc20Tick.Net})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: brc20Tick.Tick})
		bsonData = append(bsonData, bson.E{Key: "name", Value: brc20Tick.Name})
		bsonData = append(bsonData, bson.E{Key: "decimal", Value: brc20Tick.Decimal})
		bsonData = append(bsonData, bson.E{Key: "supply", Value: brc20Tick.Supply})
		bsonData = append(bsonData, bson.E{Key: "icon", Value: brc20Tick.Icon})
		bsonData = append(bsonData, bson.E{Key: "defaultLimit", Value: brc20Tick.DefaultLimit})
		bsonData = append(bsonData, bson.E{Key: "deployer", Value: brc20Tick.Deployer})
		bsonData = append(bsonData, bson.E{Key: "deployTime", Value: brc20Tick.DeployTime})
		bsonData = append(bsonData, bson.E{Key: "deployContract", Value: brc20Tick.DeployContract})
		bsonData = append(bsonData, bson.E{Key: "description", Value: brc20Tick.Description})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return brc20Tick, nil
	} else {
		return createBrc20TickInfoModel(brc20Tick)
	}
}

func FindBrc20TickKlineModelByTickId(tickId string) (*model.Brc20TickKlineModel, error) {
	collection, err := model.Brc20TickKlineModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"tickId", tickId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.Brc20TickKlineModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createBrc20TickKlineModel(brc20TickKline *model.Brc20TickKlineModel) (*model.Brc20TickKlineModel, error) {
	collection, err := model.Brc20TickKlineModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "tickId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "volume")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "timeType")
	CreateIndex(collection, "date")

	entity := &model.Brc20TickKlineModel{
		Id:            util.GetUUIDInt64(),
		TickId:        brc20TickKline.TickId,
		Net:           brc20TickKline.Net,
		Tick:          brc20TickKline.Tick,
		Open:          brc20TickKline.Open,
		High:          brc20TickKline.High,
		Low:           brc20TickKline.Low,
		Close:         brc20TickKline.Close,
		Volume:        brc20TickKline.Volume,
		Date:          brc20TickKline.Date,
		DateTimestamp: brc20TickKline.DateTimestamp,
		Timestamp:     brc20TickKline.Timestamp,
		TimeType:      brc20TickKline.TimeType,
		CreateTime:    util.Time(),
		State:         model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetBrc20TickKlineModel(brc20TickKline *model.Brc20TickKlineModel) (*model.Brc20TickKlineModel, error) {
	entity, err := FindBrc20TickKlineModelByTickId(brc20TickKline.TickId)
	if err == nil && entity != nil {
		collection, err := model.Brc20TickKlineModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"tickId", brc20TickKline.TickId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "tickId", Value: brc20TickKline.TickId})
		bsonData = append(bsonData, bson.E{Key: "net", Value: brc20TickKline.Net})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: brc20TickKline.Tick})
		bsonData = append(bsonData, bson.E{Key: "open", Value: brc20TickKline.Open})
		bsonData = append(bsonData, bson.E{Key: "high", Value: brc20TickKline.High})
		bsonData = append(bsonData, bson.E{Key: "low", Value: brc20TickKline.Low})
		bsonData = append(bsonData, bson.E{Key: "close", Value: brc20TickKline.Close})
		bsonData = append(bsonData, bson.E{Key: "volume", Value: brc20TickKline.Volume})
		bsonData = append(bsonData, bson.E{Key: "date", Value: brc20TickKline.Date})
		bsonData = append(bsonData, bson.E{Key: "dateTimestamp", Value: brc20TickKline.DateTimestamp})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: brc20TickKline.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "timeType", Value: brc20TickKline.TimeType})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return brc20TickKline, nil
	} else {
		return createBrc20TickKlineModel(brc20TickKline)
	}
}

func FindNewestBrc20TickKlineModel(net, tick string) (*model.Brc20TickKlineModel, error) {
	collection, err := model.Brc20TickKlineModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"tick", tick},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.Brc20TickKlineModel{}
	sort := options.FindOne().SetSort(bson.M{"timestamp": -1})
	err = collection.FindOne(context.TODO(), queryBson, sort).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func FindBrc20TickKlineModelList(net, tick string, startTime, endTime, limit int64, timeType model.TimeType) ([]*model.Brc20TickKlineModel, error) {
	collection, err := model.Brc20TickKlineModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":   net,
		"tick":  tick,
		"state": model.STATE_EXIST,
	}
	if timeType != "" {
		find["timeType"] = timeType
	}

	between := bson.M{GTE_: startTime, LTE_: endTime}
	find["timestamp"] = between

	models := make([]*model.Brc20TickKlineModel, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(0)
	sort := options.Find().SetSort(bson.M{"timestamp": -1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.Brc20TickKlineModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get Brc20TickKlineModel Error")
	}
	return models, nil
}

func FindBrc20TickRecentlyInfoModelByTickId(tickId string) (*model.Brc20TickRecentlyInfoModel, error) {
	collection, err := model.Brc20TickRecentlyInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"tickId", tickId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.Brc20TickRecentlyInfoModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createBrc20TickRecentlyInfoModel(brc20TickRecentlyInfo *model.Brc20TickRecentlyInfoModel) (*model.Brc20TickRecentlyInfoModel, error) {
	collection, err := model.Brc20TickRecentlyInfoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "tickId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "volume")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "recentlyType")

	entity := &model.Brc20TickRecentlyInfoModel{
		Id:            util.GetUUIDInt64(),
		TickId:        brc20TickRecentlyInfo.TickId,
		Net:           brc20TickRecentlyInfo.Net,
		Tick:          brc20TickRecentlyInfo.Tick,
		Highest:       brc20TickRecentlyInfo.Highest,
		Lowest:        brc20TickRecentlyInfo.Lowest,
		Volume:        brc20TickRecentlyInfo.Volume,
		Percentage:    brc20TickRecentlyInfo.Percentage,
		RecentlyType:  brc20TickRecentlyInfo.RecentlyType,
		OrderLastTime: brc20TickRecentlyInfo.OrderLastTime,
		Timestamp:     brc20TickRecentlyInfo.Timestamp,
		CreateTime:    util.Time(),
		State:         model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetBrc20TickRecentlyInfoModel(brc20TickRecentlyInfo *model.Brc20TickRecentlyInfoModel) (*model.Brc20TickRecentlyInfoModel, error) {
	entity, err := FindBrc20TickRecentlyInfoModelByTickId(brc20TickRecentlyInfo.TickId)
	if err == nil && entity != nil {
		collection, err := model.Brc20TickRecentlyInfoModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"tickId", brc20TickRecentlyInfo.TickId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "tickId", Value: brc20TickRecentlyInfo.TickId})
		bsonData = append(bsonData, bson.E{Key: "net", Value: brc20TickRecentlyInfo.Net})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: brc20TickRecentlyInfo.Tick})
		bsonData = append(bsonData, bson.E{Key: "highest", Value: brc20TickRecentlyInfo.Highest})
		bsonData = append(bsonData, bson.E{Key: "lowest", Value: brc20TickRecentlyInfo.Lowest})
		bsonData = append(bsonData, bson.E{Key: "volume", Value: brc20TickRecentlyInfo.Volume})
		bsonData = append(bsonData, bson.E{Key: "percentage", Value: brc20TickRecentlyInfo.Percentage})
		bsonData = append(bsonData, bson.E{Key: "recentlyType", Value: brc20TickRecentlyInfo.RecentlyType})
		bsonData = append(bsonData, bson.E{Key: "orderLastTime", Value: brc20TickRecentlyInfo.OrderLastTime})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: brc20TickRecentlyInfo.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return brc20TickRecentlyInfo, nil
	} else {
		return createBrc20TickRecentlyInfoModel(brc20TickRecentlyInfo)
	}
}

func FindBlockInfoModelByBlockId(blockId string) (*model.BlockInfoModel, error) {
	collection, err := model.BlockInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"blockId", blockId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.BlockInfoModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createBlockInfoModel(blockInfo *model.BlockInfoModel) (*model.BlockInfoModel, error) {
	collection, err := model.BlockInfoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "blockId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "chain")
	CreateIndex(collection, "height")
	CreateIndex(collection, "blockTime")
	CreateIndex(collection, "timestamp")

	entity := &model.BlockInfoModel{
		Id:           util.GetUUIDInt64(),
		BlockId:      blockInfo.BlockId,
		Net:          blockInfo.Net,
		Chain:        blockInfo.Chain,
		Height:       blockInfo.Height,
		Hash:         blockInfo.Hash,
		BlockTime:    blockInfo.BlockTime,
		BlockTimeStr: blockInfo.BlockTimeStr,
		Timestamp:    blockInfo.Timestamp,
		CreateTime:   util.Time(),
		State:        model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetBlockInfoModel(blockInfo *model.BlockInfoModel) (*model.BlockInfoModel, error) {
	entity, err := FindBlockInfoModelByBlockId(blockInfo.BlockId)
	if err == nil && entity != nil {
		collection, err := model.BlockInfoModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"blockId", blockInfo.BlockId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "blockId", Value: blockInfo.BlockId})
		bsonData = append(bsonData, bson.E{Key: "net", Value: blockInfo.Net})
		bsonData = append(bsonData, bson.E{Key: "chain", Value: blockInfo.Chain})
		bsonData = append(bsonData, bson.E{Key: "height", Value: blockInfo.Height})
		bsonData = append(bsonData, bson.E{Key: "hash", Value: blockInfo.Hash})
		bsonData = append(bsonData, bson.E{Key: "blockTime", Value: blockInfo.BlockTime})
		bsonData = append(bsonData, bson.E{Key: "blockTimeStr", Value: blockInfo.BlockTimeStr})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: blockInfo.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return blockInfo, nil
	} else {
		return createBlockInfoModel(blockInfo)
	}
}

func FindNewestHeightBlockInfoModel(net, chain string) (*model.BlockInfoModel, error) {
	collection, err := model.BlockInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"chain", chain},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.BlockInfoModel{}
	sort := options.FindOne().SetSort(bson.M{"height": -1})
	err = collection.FindOne(context.TODO(), queryBson, sort).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func FindBlockInfoModelList(net, chain string, skip, limit int64) ([]*model.BlockInfoModel, error) {
	collection, err := model.BlockInfoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":   net,
		"chain": chain,
		//"state": model.STATE_EXIST,
	}

	models := make([]*model.BlockInfoModel, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"height": 1})
	if cursor, err := collection.Find(context.Background(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.BlockInfoModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get BlockInfoModel Error")
	}
	return models, nil
}

func FindNewestHeightBlockInfoModelByBlockTime(net, chain string, blockTime int64) (*model.BlockInfoModel, error) {
	collection, err := model.BlockInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.M{
		"net":   net,
		"chain": chain,
		//{"state", model.STATE_EXIST},
	}
	if blockTime != 0 {
		queryBson["blockTime"] = bson.M{
			LTE_: blockTime,
		}
	}

	entity := &model.BlockInfoModel{}
	sort := options.FindOne().SetSort(bson.M{"height": -1})
	err = collection.FindOne(context.TODO(), queryBson, sort).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}
