package mongo_service

import (
	"context"
	"errors"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/model"
)

func FindPoolBrc20ModelByOrderId(orderId string) (*model.PoolBrc20Model, error) {
	collection, err := model.PoolBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"orderId", orderId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.PoolBrc20Model{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createPoolBrc20Model(poolBrc20 *model.PoolBrc20Model) (*model.PoolBrc20Model, error) {
	collection, err := model.PoolBrc20Model{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "orderId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "Pair")
	CreateIndex(collection, "coinRatePrice")
	CreateIndex(collection, "coinAddress")
	CreateIndex(collection, "address")
	CreateIndex(collection, "inscriptionId")
	CreateIndex(collection, "utxoId")
	CreateIndex(collection, "poolType")
	CreateIndex(collection, "poolState")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "dealTime")

	entity := &model.PoolBrc20Model{
		Id:             util.GetUUIDInt64(),
		Net:            poolBrc20.Net,
		OrderId:        poolBrc20.OrderId,
		Tick:           poolBrc20.Tick,
		Pair:           poolBrc20.Pair,
		CoinAmount:     poolBrc20.CoinAmount,
		CoinDecimalNum: poolBrc20.CoinDecimalNum,
		Amount:         poolBrc20.Amount,
		DecimalNum:     poolBrc20.DecimalNum,
		CoinRatePrice:  poolBrc20.CoinRatePrice,
		CoinAddress:    poolBrc20.CoinAddress,
		Address:        poolBrc20.Address,
		CoinPsbtRaw:    poolBrc20.CoinPsbtRaw,
		PsbtRaw:        poolBrc20.PsbtRaw,
		InscriptionId:  poolBrc20.InscriptionId,
		UtxoId:         poolBrc20.UtxoId,
		PoolType:       poolBrc20.PoolType,
		PoolState:      poolBrc20.PoolState,
		DealTime:       poolBrc20.DealTime,
		Timestamp:      poolBrc20.Timestamp,
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

func SetPoolBrc20Model(poolBrc20 *model.PoolBrc20Model) (*model.PoolBrc20Model, error) {
	entity, err := FindPoolBrc20ModelByOrderId(poolBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"orderId", poolBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: poolBrc20.Net})
		bsonData = append(bsonData, bson.E{Key: "orderId", Value: poolBrc20.OrderId})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: poolBrc20.Tick})
		bsonData = append(bsonData, bson.E{Key: "pair", Value: poolBrc20.Pair})
		bsonData = append(bsonData, bson.E{Key: "coinAmount", Value: poolBrc20.CoinAmount})
		bsonData = append(bsonData, bson.E{Key: "coinDecimalNum", Value: poolBrc20.CoinDecimalNum})
		bsonData = append(bsonData, bson.E{Key: "amount", Value: poolBrc20.Amount})
		bsonData = append(bsonData, bson.E{Key: "decimalNum", Value: poolBrc20.DecimalNum})
		bsonData = append(bsonData, bson.E{Key: "coinRatePrice", Value: poolBrc20.CoinRatePrice})
		bsonData = append(bsonData, bson.E{Key: "coinAddress", Value: poolBrc20.CoinAddress})
		bsonData = append(bsonData, bson.E{Key: "address", Value: poolBrc20.Address})
		bsonData = append(bsonData, bson.E{Key: "coinPsbtRaw", Value: poolBrc20.CoinPsbtRaw})
		bsonData = append(bsonData, bson.E{Key: "psbtRaw", Value: poolBrc20.PsbtRaw})
		bsonData = append(bsonData, bson.E{Key: "inscriptionId", Value: poolBrc20.InscriptionId})
		bsonData = append(bsonData, bson.E{Key: "utxoId", Value: poolBrc20.UtxoId})
		bsonData = append(bsonData, bson.E{Key: "poolType", Value: poolBrc20.PoolType})
		bsonData = append(bsonData, bson.E{Key: "poolState", Value: poolBrc20.PoolState})
		bsonData = append(bsonData, bson.E{Key: "dealTime", Value: poolBrc20.DealTime})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: poolBrc20.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return poolBrc20, nil
	} else {
		return createPoolBrc20Model(poolBrc20)
	}
}

func CountPoolBrc20ModelList(net, tick, pair, address string, poolType model.PoolType, poolState model.PoolState) (int64, error) {
	collection, err := model.PoolBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if pair != "" {
		find["pair"] = pair
	}
	if address != "" {
		find["address"] = address
	}
	if poolType != 0 {
		find["poolType"] = poolType
	}
	if poolState != 0 {
		find["poolState"] = poolState
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindPoolBrc20ModelList(net, tick, pair, address string,
	poolType model.PoolType, poolState model.PoolState,
	limit int64, flag, page int64, sortKey string, sortType int64) ([]*model.PoolBrc20Model, error) {
	collection, err := model.PoolBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if pair != "" {
		find["pair"] = pair
	}
	if address != "" {
		find["address"] = address
	}
	if poolType != 0 {
		find["poolType"] = poolType
	}
	if poolState != 0 {
		find["poolState"] = poolState
	}

	switch sortKey {
	case "coinRatePrice":
		sortKey = "coinRatePrice"
	default:
		sortKey = "timestamp"
	}

	flagKey := GT_
	if sortType >= 0 {
		sortType = 1
		flagKey = GT_
	} else {
		sortType = -1
		flagKey = LT_
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	} else if flag != 0 {
		find[sortKey] = bson.M{flagKey: flag}
	}

	models := make([]*model.PoolBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{sortKey: sortType})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.PoolBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get PoolBrc20Model Error")
	}
	return models, nil
}

func FindPoolInfoModelByPair(net, pair string) (*model.PoolInfoModel, error) {
	collection, err := model.PoolInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"pair", pair},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.PoolInfoModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createPoolInfoModel(poolInfo *model.PoolInfoModel) (*model.PoolInfoModel, error) {
	collection, err := model.PoolInfoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateIndex(collection, "net")
	CreateIndex(collection, "pair")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "timestamp")

	entity := &model.PoolInfoModel{
		Id:             util.GetUUIDInt64(),
		Net:            poolInfo.Net,
		Pair:           poolInfo.Pair,
		Tick:           poolInfo.Tick,
		CoinAmount:     poolInfo.CoinAmount,
		CoinDecimalNum: poolInfo.CoinDecimalNum,
		Amount:         poolInfo.Amount,
		DecimalNum:     poolInfo.DecimalNum,
		Timestamp:      poolInfo.Timestamp,
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

func SetPoolInfoModel(poolInfo *model.PoolInfoModel) (*model.PoolInfoModel, error) {
	entity, err := FindPoolInfoModelByPair(poolInfo.Net, poolInfo.Pair)
	if err == nil && entity != nil {
		collection, err := model.PoolInfoModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"net", poolInfo.Net},
			{"pair", poolInfo.Pair},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: poolInfo.Net})
		bsonData = append(bsonData, bson.E{Key: "pair", Value: poolInfo.Pair})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: poolInfo.Tick})
		bsonData = append(bsonData, bson.E{Key: "coinAmount", Value: poolInfo.CoinAmount})
		bsonData = append(bsonData, bson.E{Key: "coinDecimalNum", Value: poolInfo.CoinDecimalNum})
		bsonData = append(bsonData, bson.E{Key: "amount", Value: poolInfo.Amount})
		bsonData = append(bsonData, bson.E{Key: "decimalNum", Value: poolInfo.DecimalNum})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return poolInfo, nil
	} else {
		return createPoolInfoModel(poolInfo)
	}
}

func CountPoolInfoModelList(net, tick, pair string) (int64, error) {
	collection, err := model.PoolInfoModel{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if pair != "" {
		find["pair"] = pair
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindPoolInfoModelList(net, tick, pair string) ([]*model.PoolInfoModel, error) {
	collection, err := model.PoolInfoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if pair != "" {
		find["pair"] = pair
	}

	models := make([]*model.PoolInfoModel, 0)
	pagination := options.Find().SetLimit(100).SetSkip(0)
	if cursor, err := collection.Find(context.TODO(), find, pagination); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.PoolInfoModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get PoolInfoModel Error")
	}
	return models, nil
}
