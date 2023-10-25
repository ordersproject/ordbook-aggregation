package mongo_service

import (
	"context"
	"errors"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/model"
	"strings"
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
	CreateIndex(collection, "btcPoolMode")
	CreateIndex(collection, "poolType")
	CreateIndex(collection, "poolState")
	CreateIndex(collection, "claimTxBlock")
	CreateIndex(collection, "dealCoinTxBlock")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "dealTime")
	CreateIndex(collection, "claimTime")
	CreateIndex(collection, "claimTxBlockState")
	CreateIndex(collection, "dealCoinTxBlockState")

	entity := &model.PoolBrc20Model{
		Id:                       util.GetUUIDInt64(),
		Net:                      poolBrc20.Net,
		OrderId:                  poolBrc20.OrderId,
		Tick:                     poolBrc20.Tick,
		Pair:                     poolBrc20.Pair,
		CoinAmount:               poolBrc20.CoinAmount,
		CoinDecimalNum:           poolBrc20.CoinDecimalNum,
		Amount:                   poolBrc20.Amount,
		DecimalNum:               poolBrc20.DecimalNum,
		CoinRatePrice:            poolBrc20.CoinRatePrice,
		CoinAddress:              poolBrc20.CoinAddress,
		CoinPublicKey:            poolBrc20.CoinPublicKey,
		CoinInputValue:           poolBrc20.CoinInputValue,
		Address:                  poolBrc20.Address,
		MultiSigScript:           poolBrc20.MultiSigScript,
		MultiSigScriptAddress:    poolBrc20.MultiSigScriptAddress,
		CoinPsbtRaw:              poolBrc20.CoinPsbtRaw,
		MultiSigScriptBtc:        poolBrc20.MultiSigScriptBtc,
		MultiSigScriptAddressBtc: poolBrc20.MultiSigScriptAddressBtc,
		PsbtRaw:                  poolBrc20.PsbtRaw,
		InscriptionId:            poolBrc20.InscriptionId,
		InscriptionNumber:        poolBrc20.InscriptionNumber,
		BtcPoolMode:              poolBrc20.BtcPoolMode,
		UtxoId:                   poolBrc20.UtxoId,
		RefundTx:                 poolBrc20.RefundTx,
		PoolType:                 poolBrc20.PoolType,
		PoolState:                poolBrc20.PoolState,
		DealTime:                 poolBrc20.DealTime,
		Ratio:                    poolBrc20.Ratio,
		RewardRatio:              poolBrc20.RewardRatio,
		Timestamp:                poolBrc20.Timestamp,
		CreateTime:               util.Time(),
		State:                    model.STATE_EXIST,
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
		bsonData = append(bsonData, bson.E{Key: "coinPublicKey", Value: poolBrc20.CoinPublicKey})
		bsonData = append(bsonData, bson.E{Key: "coinInputValue", Value: poolBrc20.CoinInputValue})
		bsonData = append(bsonData, bson.E{Key: "address", Value: poolBrc20.Address})
		bsonData = append(bsonData, bson.E{Key: "multiSigScript", Value: poolBrc20.MultiSigScript})
		bsonData = append(bsonData, bson.E{Key: "multiSigScriptAddress", Value: poolBrc20.MultiSigScriptAddress})
		bsonData = append(bsonData, bson.E{Key: "coinPsbtRaw", Value: poolBrc20.CoinPsbtRaw})
		bsonData = append(bsonData, bson.E{Key: "multiSigScriptBtc", Value: poolBrc20.MultiSigScriptBtc})
		bsonData = append(bsonData, bson.E{Key: "multiSigScriptAddressBtc", Value: poolBrc20.MultiSigScriptAddressBtc})
		bsonData = append(bsonData, bson.E{Key: "psbtRaw", Value: poolBrc20.PsbtRaw})
		bsonData = append(bsonData, bson.E{Key: "inscriptionId", Value: poolBrc20.InscriptionId})
		bsonData = append(bsonData, bson.E{Key: "inscriptionNumber", Value: poolBrc20.InscriptionNumber})
		bsonData = append(bsonData, bson.E{Key: "btcPoolMode", Value: poolBrc20.BtcPoolMode})
		bsonData = append(bsonData, bson.E{Key: "utxoId", Value: poolBrc20.UtxoId})
		bsonData = append(bsonData, bson.E{Key: "refundTx", Value: poolBrc20.RefundTx})
		bsonData = append(bsonData, bson.E{Key: "poolType", Value: poolBrc20.PoolType})
		bsonData = append(bsonData, bson.E{Key: "poolState", Value: poolBrc20.PoolState})
		bsonData = append(bsonData, bson.E{Key: "dealTime", Value: poolBrc20.DealTime})
		bsonData = append(bsonData, bson.E{Key: "ratio", Value: poolBrc20.Ratio})
		bsonData = append(bsonData, bson.E{Key: "rewardRatio", Value: poolBrc20.RewardRatio})
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

func SetPoolBrc20ModelForStatus(orderId string, status model.PoolState, dealTx string, dealTxIndex, dealTxOutValue, dealTime int64) error {
	entity, err := FindPoolBrc20ModelByOrderId(orderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", orderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "poolState", Value: status})
		bsonData = append(bsonData, bson.E{Key: "dealTx", Value: dealTx})
		bsonData = append(bsonData, bson.E{Key: "dealTxIndex", Value: dealTxIndex})
		bsonData = append(bsonData, bson.E{Key: "dealTxOutValue", Value: dealTxOutValue})
		bsonData = append(bsonData, bson.E{Key: "dealTime", Value: dealTime})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForCoinStatus(orderId string, coinStatus model.PoolState, dealCoinTx string, dealCoinTxIndex, dealCoinTxOutValue, dealCoinTime int64) error {
	entity, err := FindPoolBrc20ModelByOrderId(orderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", orderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "poolCoinState", Value: coinStatus})
		bsonData = append(bsonData, bson.E{Key: "dealCoinTx", Value: dealCoinTx})
		bsonData = append(bsonData, bson.E{Key: "dealCoinTxIndex", Value: dealCoinTxIndex})
		bsonData = append(bsonData, bson.E{Key: "dealCoinTxOutValue", Value: dealCoinTxOutValue})
		bsonData = append(bsonData, bson.E{Key: "dealCoinTime", Value: dealCoinTime})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForDealInscription(orderId string, dealInscriptionId, dealInscriptionTx string, dealInscriptionTxIndex, dealInscriptionTxOutValue, dealInscriptionTime int64) error {
	entity, err := FindPoolBrc20ModelByOrderId(orderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", orderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "dealInscriptionId", Value: dealInscriptionId})
		bsonData = append(bsonData, bson.E{Key: "dealInscriptionTx", Value: dealInscriptionTx})
		bsonData = append(bsonData, bson.E{Key: "dealInscriptionTxIndex", Value: dealInscriptionTxIndex})
		bsonData = append(bsonData, bson.E{Key: "dealInscriptionTxOutValue", Value: dealInscriptionTxOutValue})
		bsonData = append(bsonData, bson.E{Key: "dealInscriptionTime", Value: dealInscriptionTime})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForDealInscriptionInTool(orderId string, dealInscriptionId, dealInscriptionTx string) error {
	entity, err := FindPoolBrc20ModelByOrderId(orderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", orderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "dealInscriptionId", Value: dealInscriptionId})
		bsonData = append(bsonData, bson.E{Key: "dealInscriptionTx", Value: dealInscriptionTx})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForClaim(poolBrc20 *model.PoolBrc20Model) error {
	entity, err := FindPoolBrc20ModelByOrderId(poolBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", poolBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "claimTx", Value: poolBrc20.ClaimTx})
		bsonData = append(bsonData, bson.E{Key: "claimTxBlockState", Value: poolBrc20.ClaimTxBlockState})
		bsonData = append(bsonData, bson.E{Key: "claimTime", Value: poolBrc20.ClaimTime})
		bsonData = append(bsonData, bson.E{Key: "poolState", Value: poolBrc20.PoolState})
		bsonData = append(bsonData, bson.E{Key: "poolCoinState", Value: poolBrc20.PoolCoinState})
		bsonData = append(bsonData, bson.E{Key: "rewardRealAmount", Value: poolBrc20.RewardRealAmount})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForBlock(poolBrc20 *model.PoolBrc20Model) error {
	entity, err := FindPoolBrc20ModelByOrderId(poolBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", poolBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "claimTxBlock", Value: poolBrc20.ClaimTxBlock})
		bsonData = append(bsonData, bson.E{Key: "claimTxBlockState", Value: poolBrc20.ClaimTxBlockState})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForDealBlock(poolBrc20 *model.PoolBrc20Model) error {
	entity, err := FindPoolBrc20ModelByOrderId(poolBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", poolBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "dealCoinTxBlock", Value: poolBrc20.DealCoinTxBlock})
		bsonData = append(bsonData, bson.E{Key: "dealCoinTxBlockState", Value: poolBrc20.DealCoinTxBlockState})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForReward(poolBrc20 *model.PoolBrc20Model) error {
	entity, err := FindPoolBrc20ModelByOrderId(poolBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", poolBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "rewardAmount", Value: poolBrc20.RewardAmount})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForCalReward(poolBrc20 *model.PoolBrc20Model) error {
	entity, err := FindPoolBrc20ModelByOrderId(poolBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", poolBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "percentage", Value: poolBrc20.Percentage})
		bsonData = append(bsonData, bson.E{Key: "rewardAmount", Value: poolBrc20.RewardAmount})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetPoolBrc20ModelForCalExtraReward(poolBrc20 *model.PoolBrc20Model) error {
	entity, err := FindPoolBrc20ModelByOrderId(poolBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.PoolBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", poolBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "percentageExtra", Value: poolBrc20.PercentageExtra})
		bsonData = append(bsonData, bson.E{Key: "rewardExtraAmount", Value: poolBrc20.RewardExtraAmount})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
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
		find["coinAddress"] = address
	}
	//if poolType != 0 {
	//	find["poolType"] = poolType
	//}
	if poolType != 0 {
		if poolType == model.PoolTypeAll {
			find["poolType"] = bson.M{IN_: []model.PoolType{
				model.PoolTypeTick,
				model.PoolTypeBoth,
				model.PoolTypeBtc,
			}}
		} else {
			find["poolType"] = poolType
		}
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
	limit, flag, page int64, sortKey string, sortType int64) ([]*model.PoolBrc20Model, error) {
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
		find["coinAddress"] = address
	}
	//if poolType != 0 {
	//	find["poolType"] = poolType
	//}
	if poolType != 0 {
		if poolType == model.PoolTypeAll {
			find["poolType"] = bson.M{IN_: []model.PoolType{
				model.PoolTypeTick,
				model.PoolTypeBoth,
				model.PoolTypeBtc,
			}}
		} else {
			find["poolType"] = poolType
		}
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

func FindPoolBrc20ModelListByClaimTime(net, tick, pair, address string, poolState model.PoolState,
	limit, page int64, claimTxBlockState model.ClaimTxBlockState) ([]*model.PoolBrc20Model, error) {
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
		find["coinAddress"] = address
	}
	if poolState != 0 {
		find["poolState"] = poolState
	}
	if claimTxBlockState != 0 {
		find["claimTxBlockState"] = claimTxBlockState
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	}

	models := make([]*model.PoolBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"claimTime": 1})
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

func FindPoolBrc20ModelListByStartAndEndBlock(net, tick, pair, address string,
	poolType model.PoolType, poolState model.PoolState,
	limit, page int64, startBlock, endBlock int64) ([]*model.PoolBrc20Model, error) {
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
		find["coinAddress"] = address
	}
	if poolType != 0 {
		if poolType == model.PoolTypeAll {
			find["poolType"] = bson.M{IN_: []model.PoolType{
				model.PoolTypeTick,
				model.PoolTypeBoth,
				model.PoolTypeBtc,
			}}
		} else {
			find["poolType"] = poolType
		}
	}

	if startBlock != 0 && endBlock != 0 {
		between := bson.M{
			GTE_: startBlock,
			LTE_: endBlock,
		}
		find["claimTxBlock"] = between
	}

	if poolState != 0 {
		find["poolState"] = poolState
	}
	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	}

	models := make([]*model.PoolBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"claimTxBlock": 1})
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

func FindPoolBrc20ModelListByDealStartAndDealEndBlock(net, tick, pair, address string,
	poolType model.PoolType, poolState model.PoolState,
	limit, page int64, startBlock, endBlock int64) ([]*model.PoolBrc20Model, error) {
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
		find["coinAddress"] = address
	}
	if poolType != 0 {
		if poolType == model.PoolTypeAll {
			find["poolType"] = bson.M{IN_: []model.PoolType{
				model.PoolTypeTick,
				model.PoolTypeBoth,
				model.PoolTypeBtc,
			}}
		} else {
			find["poolType"] = poolType
		}
	}

	if startBlock != 0 && endBlock != 0 {
		between := bson.M{
			GTE_: startBlock,
			LTE_: endBlock,
		}
		find["dealCoinTxBlock"] = between
	}

	if poolState != 0 {
		find["poolState"] = poolState
	}
	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	}

	models := make([]*model.PoolBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"dealCoinTxBlock": 1})
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

func FindPoolBrc20ModelListByDealTime(net, tick, pair, address string, poolState model.PoolState,
	limit, page int64, dealCoinTxBlockState model.ClaimTxBlockState) ([]*model.PoolBrc20Model, error) {
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
		find["coinAddress"] = address
	}
	if poolState != 0 {
		find["poolState"] = poolState
	}
	if dealCoinTxBlockState != 0 {
		find["dealCoinTxBlockState"] = dealCoinTxBlockState
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	}

	models := make([]*model.PoolBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"dealTime": 1})
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

func FindPoolBrc20ModelListByEndTime(net, tick, pair, address string,
	poolType model.PoolType, poolState model.PoolState,
	limit, page int64, endTime int64) ([]*model.PoolBrc20Model, error) {
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
		find["coinAddress"] = address
	}
	if poolType != 0 {
		if poolType == model.PoolTypeAll {
			find["poolType"] = bson.M{IN_: []model.PoolType{
				model.PoolTypeTick,
				model.PoolTypeBoth,
				model.PoolTypeBtc,
			}}
		} else {
			find["poolType"] = poolType
		}
	}

	if endTime != 0 {
		between := bson.M{
			LTE_: endTime,
		}
		find["timestamp"] = between
	}

	if poolState != 0 {
		find["poolState"] = poolState
	}
	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	}

	models := make([]*model.PoolBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"timestamp": 1})
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

func CountOwnPoolPair(net, tick, pair, address string, poolType model.PoolType) (*model.PoolOrderCount, error) {
	collection, err := model.PoolBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	countInfo := &model.PoolOrderCount{
		Id:              address,
		CoinAmountTotal: 0,
		AmountTotal:     0,
	}
	countInfoList := make([]model.PoolOrderCount, 0)

	pipeline := mongo.Pipeline{
		{
			{"$match", bson.D{
				{"net", net},
				{"tick", tick},
				{"pair", pair},
				{"coinAddress", address},
				{"poolState", model.PoolStateAdd},
				{"poolType", poolType},
			}},
		},
		{
			{"$group", bson.D{
				{"_id", "$coinAddress"},
				{"coinAmountTotal", bson.D{
					{"$sum", "$coinAmount"},
				}},
				{"amountTotal", bson.D{
					{"$sum", "$amount"},
				}},
				{"orderCounts", bson.D{
					{"$sum", 1},
				}},
			}},
		},
	}
	if cursor, err := collection.Aggregate(context.Background(), pipeline); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			var entity model.PoolOrderCount
			if err = cursor.Decode(&entity); err == nil {
				countInfoList = append(countInfoList, entity)
			}
		}
		if countInfoList != nil && len(countInfoList) != 0 {
			for _, v := range countInfoList {
				if v.Id == address {
					countInfo = &v
					break
				}
			}
		}
		return countInfo, nil
	} else {
		return nil, errors.New("db get records error")
	}
}

func FindUsedInscriptionPool(inscriptionId string) (int64, error) {
	if strings.Contains(inscriptionId, "i") {
		inscriptionId = strings.ReplaceAll(inscriptionId, "i", ":")
	}

	collection, err := model.PoolBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"inscriptionId": inscriptionId,
		"poolState": bson.M{IN_: []model.PoolState{
			model.PoolStateAdd,
			model.PoolStateUsed,
			model.PoolStateClaim,
		}},
		"state": model.STATE_EXIST,
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindUsedInscriptionPoolFinish(inscriptionId string) (int64, error) {
	if strings.Contains(inscriptionId, "i") {
		inscriptionId = strings.ReplaceAll(inscriptionId, "i", ":")
	}

	collection, err := model.PoolBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"inscriptionId": inscriptionId,
		"poolState": bson.M{IN_: []model.PoolState{
			model.PoolStateUsed,
			model.PoolStateClaim,
		}},
		"state": model.STATE_EXIST,
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindPoolBrc20ModelListByInscriptionId(inscriptionId string, limit, flag, page int64, sortKey string, sortType int64) ([]*model.PoolBrc20Model, error) {
	collection, err := model.PoolBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"inscriptionId": inscriptionId,
		"state":         model.STATE_EXIST,
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

func CountOwnPoolReward(net, tick, pair, address string) (*model.PoolRewardCount, error) {
	collection, err := model.PoolBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	countInfo := &model.PoolRewardCount{
		Id:                     address,
		CoinAmountTotal:        0,
		AmountTotal:            0,
		RewardAmountTotal:      0,
		RewardExtraAmountTotal: 0,
	}
	countInfoList := make([]model.PoolRewardCount, 0)

	match := bson.D{
		{"net", net},
		{"coinAddress", address},
		{"poolState", model.PoolStateClaim},
	}
	if tick != "" {
		match = append(match, bson.E{Key: "tick", Value: tick})
	}
	if pair != "" {
		match = append(match, bson.E{Key: "pair", Value: pair})
	}

	pipeline := mongo.Pipeline{
		{
			{"$match", match},
		},
		{
			{"$group", bson.D{
				{"_id", "$coinAddress"},
				{"coinAmountTotal", bson.D{
					{"$sum", "$coinAmount"},
				}},
				{"amountTotal", bson.D{
					{"$sum", "$amount"},
				}},
				{"orderCounts", bson.D{
					{"$sum", 1},
				}},
				{"rewardAmountTotal", bson.D{
					{"$sum", "$rewardRealAmount"},
				}},
				{"rewardExtraAmountTotal", bson.D{
					{"$sum", "$rewardExtraAmount"},
				}},
			}},
		},
	}
	if cursor, err := collection.Aggregate(context.Background(), pipeline); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			var entity model.PoolRewardCount
			if err = cursor.Decode(&entity); err == nil {
				countInfoList = append(countInfoList, entity)
			}
		}
		if countInfoList != nil && len(countInfoList) != 0 {
			for _, v := range countInfoList {
				if v.Id == address {
					countInfo = &v
					break
				}
			}
		}
		return countInfo, nil
	} else {
		return nil, errors.New("db get records error")
	}
}

func FindPoolRewardOrderModelByOrderId(orderId string) (*model.PoolRewardOrderModel, error) {
	collection, err := model.PoolRewardOrderModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"orderId", orderId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.PoolRewardOrderModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createPoolRewardOrderModel(poolRewardOrder *model.PoolRewardOrderModel) (*model.PoolRewardOrderModel, error) {
	collection, err := model.PoolRewardOrderModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "orderId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "address")
	CreateIndex(collection, "rewardState")
	CreateIndex(collection, "timestamp")

	entity := &model.PoolRewardOrderModel{
		Id:                  util.GetUUIDInt64(),
		Net:                 poolRewardOrder.Net,
		OrderId:             poolRewardOrder.OrderId,
		Pair:                poolRewardOrder.Pair,
		Tick:                poolRewardOrder.Tick,
		RewardCoinAmount:    poolRewardOrder.RewardCoinAmount,
		Address:             poolRewardOrder.Address,
		RewardState:         poolRewardOrder.RewardState,
		InscriptionId:       poolRewardOrder.InscriptionId,
		InscriptionOutValue: poolRewardOrder.InscriptionOutValue,
		SendId:              poolRewardOrder.SendId,
		Timestamp:           poolRewardOrder.Timestamp,
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

func SetPoolRewardOrderModel(poolRewardOrder *model.PoolRewardOrderModel) (*model.PoolRewardOrderModel, error) {
	entity, err := FindPoolRewardOrderModelByOrderId(poolRewardOrder.OrderId)
	if err == nil && entity != nil {
		collection, err := model.PoolRewardOrderModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"orderId", poolRewardOrder.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: poolRewardOrder.Net})
		bsonData = append(bsonData, bson.E{Key: "orderId", Value: poolRewardOrder.OrderId})
		bsonData = append(bsonData, bson.E{Key: "pair", Value: poolRewardOrder.Pair})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: poolRewardOrder.Tick})
		bsonData = append(bsonData, bson.E{Key: "rewardCoinAmount", Value: poolRewardOrder.RewardCoinAmount})
		bsonData = append(bsonData, bson.E{Key: "address", Value: poolRewardOrder.Address})
		bsonData = append(bsonData, bson.E{Key: "rewardState", Value: poolRewardOrder.RewardState})
		bsonData = append(bsonData, bson.E{Key: "inscriptionId", Value: poolRewardOrder.InscriptionId})
		bsonData = append(bsonData, bson.E{Key: "inscriptionOutValue", Value: poolRewardOrder.InscriptionOutValue})
		bsonData = append(bsonData, bson.E{Key: "sendId", Value: poolRewardOrder.SendId})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: poolRewardOrder.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return poolRewardOrder, nil
	} else {
		return createPoolRewardOrderModel(poolRewardOrder)
	}
}

func CountPoolRewardOrderModelList(net, tick, pair, address string, rewardState model.RewardState) (int64, error) {
	collection, err := model.PoolRewardOrderModel{}.GetReadDB()
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
	//if poolType != 0 {
	//	find["poolType"] = poolType
	//}
	if rewardState != 0 {
		if rewardState == model.RewardStateAll {
			find["rewardState"] = bson.M{IN_: []model.RewardState{
				model.RewardStateCreate,
				model.RewardStateInscription,
				model.RewardStateSend,
			}}
		} else {
			find["rewardState"] = rewardState
		}
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindPoolRewardOrderModelList(net, tick, pair, address string,
	rewardState model.RewardState,
	limit, flag, page int64, sortKey string, sortType int64) ([]*model.PoolRewardOrderModel, error) {
	collection, err := model.PoolRewardOrderModel{}.GetReadDB()
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
	//if poolType != 0 {
	//	find["poolType"] = poolType
	//}
	if rewardState != 0 {
		if rewardState == model.RewardStateAll {
			find["rewardState"] = bson.M{IN_: []model.RewardState{
				model.RewardStateCreate,
				model.RewardStateInscription,
				model.RewardStateSend,
			}}
		} else {
			find["rewardState"] = rewardState
		}
	}

	switch sortKey {
	default:
		sortKey = "timestamp"
	}

	flagKey := GT_
	//if sortType >= 0 {
	//	sortType = 1
	//	flagKey = GT_
	//} else {
	//	sortType = -1
	//	flagKey = LT_
	//}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	} else if flag != 0 {
		find[sortKey] = bson.M{flagKey: flag}
	}

	models := make([]*model.PoolRewardOrderModel, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{sortKey: -1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.PoolRewardOrderModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get PoolRewardOrderModel Error")
	}
	return models, nil
}

func CountOwnPoolRewardOrder(net, tick, pair, address string) (*model.PoolRewardOrderCount, error) {
	collection, err := model.PoolRewardOrderModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	countInfo := &model.PoolRewardOrderCount{
		Id:                    address,
		RewardCoinAmountTotal: 0,
		RewardCoinOrderCount:  0,
	}
	countInfoList := make([]model.PoolRewardOrderCount, 0)

	match := bson.D{
		{"net", net},
		//{"address", address},
		//{"poolState", model.PoolStateClaim},
	}
	if tick != "" {
		match = append(match, bson.E{Key: "tick", Value: tick})
	}
	if pair != "" {
		match = append(match, bson.E{Key: "pair", Value: pair})
	}
	if address != "" {
		match = append(match, bson.E{Key: "address", Value: address})
	}

	pipeline := mongo.Pipeline{
		{
			{"$match", match},
		},
		{
			{"$group", bson.D{
				{"_id", "$address"},
				{"rewardCoinAmountTotal", bson.D{
					{"$sum", "$rewardCoinAmount"},
				}},
				{"rewardCoinOrderCount", bson.D{
					{"$sum", 1},
				}},
			}},
		},
	}
	if cursor, err := collection.Aggregate(context.Background(), pipeline); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			var entity model.PoolRewardOrderCount
			if err = cursor.Decode(&entity); err == nil {
				countInfoList = append(countInfoList, entity)
			}
		}
		if countInfoList != nil && len(countInfoList) != 0 {
			for _, v := range countInfoList {
				if v.Id == address {
					countInfo = &v
					break
				}
			}
		}
		return countInfo, nil
	} else {
		return nil, errors.New("db get PoolRewardOrderCount error")
	}
}

func CountPoolRewardOrder(net, tick, pair, address string, poolState model.PoolState) (*model.PoolRewardOrderCount, error) {
	collection, err := model.PoolRewardOrderModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	countInfo := &model.PoolRewardOrderCount{
		Id:                    address,
		CoinAmountTotal:       0,
		AmountTotal:           0,
		RewardCoinAmountTotal: 0,
		RewardCoinOrderCount:  0,
	}
	countInfoList := make([]model.PoolRewardOrderCount, 0)

	match := bson.D{
		{"net", net},
		//{"address", address},
		//{"poolState", model.PoolStateClaim},
	}
	if tick != "" {
		match = append(match, bson.E{Key: "tick", Value: tick})
	}
	if pair != "" {
		match = append(match, bson.E{Key: "pair", Value: pair})
	}
	if address != "" {
		match = append(match, bson.E{Key: "address", Value: address})
	}

	if poolState != 0 {
		match = append(match, bson.E{Key: "poolState", Value: poolState})
	}

	pipeline := mongo.Pipeline{
		{
			{"$match", match},
		},
		{
			{"$group", bson.D{
				{"_id", "$coinAddress"},
				{"coinAmountTotal", bson.D{
					{"$sum", "$coinAmount"},
				}},
				{"amountTotal", bson.D{
					{"$sum", "$amount"},
				}},
				{"rewardCoinAmountTotal", bson.D{
					{"$sum", "$rewardCoinAmount"},
				}},
				{"rewardCoinOrderCount", bson.D{
					{"$sum", 1},
				}},
			}},
		},
	}
	if cursor, err := collection.Aggregate(context.Background(), pipeline); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			var entity model.PoolRewardOrderCount
			if err = cursor.Decode(&entity); err == nil {
				countInfoList = append(countInfoList, entity)
			}
		}
		if countInfoList != nil && len(countInfoList) != 0 {
			for _, v := range countInfoList {
				if v.Id == address {
					countInfo = &v
					break
				}
			}
		}
		return countInfo, nil
	} else {
		return nil, errors.New("db get PoolRewardOrderCount error")
	}
}

func FindPoolRewardOrderModelListByTimestamp(net, tick, pair string, limit, timestamp int64, rewardState model.RewardState) ([]*model.PoolRewardOrderModel, error) {
	collection, err := model.PoolRewardOrderModel{}.GetReadDB()
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
	if rewardState != 0 {
		find["rewardState"] = rewardState
	}

	skip := int64(0)

	models := make([]*model.PoolRewardOrderModel, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"timestamp": 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.PoolRewardOrderModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get PoolRewardOrderModel Error")
	}
	return models, nil
}

func FindPoolBlockUserInfoModelByBlockUserId(blockUserId string) (*model.PoolBlockUserInfoModel, error) {
	collection, err := model.PoolBlockUserInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"blockUserId", blockUserId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.PoolBlockUserInfoModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createPoolBlockUserInfoModel(poolBlockUserInfo *model.PoolBlockUserInfoModel) (*model.PoolBlockUserInfoModel, error) {
	collection, err := model.PoolBlockUserInfoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "blockUserId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "infoType")
	CreateIndex(collection, "address")
	CreateIndex(collection, "bigBlock")
	CreateIndex(collection, "startBlock")
	CreateIndex(collection, "cycleBlock")
	CreateIndex(collection, "percentage")
	CreateIndex(collection, "rewardAmount")
	CreateIndex(collection, "timestamp")

	entity := &model.PoolBlockUserInfoModel{
		Id:             util.GetUUIDInt64(),
		BlockUserId:    poolBlockUserInfo.BlockUserId,
		Net:            poolBlockUserInfo.Net,
		InfoType:       poolBlockUserInfo.InfoType,
		HasNoUsed:      poolBlockUserInfo.HasNoUsed,
		Address:        poolBlockUserInfo.Address,
		BigBlock:       poolBlockUserInfo.BigBlock,
		StartBlock:     poolBlockUserInfo.StartBlock,
		CycleBlock:     poolBlockUserInfo.CycleBlock,
		CoinPrice:      poolBlockUserInfo.CoinPrice,
		CoinAmount:     poolBlockUserInfo.CoinAmount,
		Amount:         poolBlockUserInfo.Amount,
		UserTotalValue: poolBlockUserInfo.UserTotalValue,
		AllTotalValue:  poolBlockUserInfo.AllTotalValue,
		Percentage:     poolBlockUserInfo.Percentage,
		RewardAmount:   poolBlockUserInfo.RewardAmount,
		Timestamp:      poolBlockUserInfo.Timestamp,
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

func SetPoolBlockUserInfoModel(poolBlockUserInfo *model.PoolBlockUserInfoModel) (*model.PoolBlockUserInfoModel, error) {
	entity, err := FindPoolBlockUserInfoModelByBlockUserId(poolBlockUserInfo.BlockUserId)
	if err == nil && entity != nil {
		collection, err := model.PoolBlockUserInfoModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"blockUserId", poolBlockUserInfo.BlockUserId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "blockUserId", Value: poolBlockUserInfo.BlockUserId})
		bsonData = append(bsonData, bson.E{Key: "net", Value: poolBlockUserInfo.Net})
		bsonData = append(bsonData, bson.E{Key: "infoType", Value: poolBlockUserInfo.InfoType})
		bsonData = append(bsonData, bson.E{Key: "hasNoUsed", Value: poolBlockUserInfo.HasNoUsed})
		bsonData = append(bsonData, bson.E{Key: "address", Value: poolBlockUserInfo.Address})
		bsonData = append(bsonData, bson.E{Key: "bigBlock", Value: poolBlockUserInfo.BigBlock})
		bsonData = append(bsonData, bson.E{Key: "startBlock", Value: poolBlockUserInfo.StartBlock})
		bsonData = append(bsonData, bson.E{Key: "cycleBlock", Value: poolBlockUserInfo.CycleBlock})
		bsonData = append(bsonData, bson.E{Key: "coinPrice", Value: poolBlockUserInfo.CoinPrice})
		bsonData = append(bsonData, bson.E{Key: "coinAmount", Value: poolBlockUserInfo.CoinAmount})
		bsonData = append(bsonData, bson.E{Key: "amount", Value: poolBlockUserInfo.Amount})
		bsonData = append(bsonData, bson.E{Key: "userTotalValue", Value: poolBlockUserInfo.UserTotalValue})
		bsonData = append(bsonData, bson.E{Key: "allTotalValue", Value: poolBlockUserInfo.AllTotalValue})
		bsonData = append(bsonData, bson.E{Key: "percentage", Value: poolBlockUserInfo.Percentage})
		bsonData = append(bsonData, bson.E{Key: "rewardAmount", Value: poolBlockUserInfo.RewardAmount})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: poolBlockUserInfo.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return poolBlockUserInfo, nil
	} else {
		return createPoolBlockUserInfoModel(poolBlockUserInfo)
	}
}

func CountPoolRewardBlockUser(net, address string) (*model.PoolRewardBlockUserCount, error) {
	collection, err := model.PoolBlockUserInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	countInfo := &model.PoolRewardBlockUserCount{
		Id:                    address,
		RewardCoinAmountTotal: 0,
	}
	countInfoList := make([]model.PoolRewardBlockUserCount, 0)

	match := bson.D{
		{"net", net},
		//{"address", address},
		//{"poolState", model.PoolStateClaim},
	}
	if address != "" {
		match = append(match, bson.E{Key: "address", Value: address})
	}

	pipeline := mongo.Pipeline{
		{
			{"$match", match},
		},
		{
			{"$group", bson.D{
				{"_id", "$address"},
				{"rewardCoinAmountTotal", bson.D{
					{"$sum", "$rewardAmount"},
				}},
			}},
		},
	}
	if cursor, err := collection.Aggregate(context.Background(), pipeline); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			var entity model.PoolRewardBlockUserCount
			if err = cursor.Decode(&entity); err == nil {
				countInfoList = append(countInfoList, entity)
			}
		}
		if countInfoList != nil && len(countInfoList) != 0 {
			for _, v := range countInfoList {
				if v.Id == address {
					countInfo = &v
					break
				}
			}
		}
		return countInfo, nil
	} else {
		return nil, errors.New("db get PoolRewardBlockUserCount error")
	}
}

func FindPoolBlockInfoModelByBigBlockId(bigBlockId string) (*model.PoolBlockInfoModel, error) {
	collection, err := model.PoolBlockInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"bigBlockId", bigBlockId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.PoolBlockInfoModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createPoolBlockInfoModel(poolBlockInfo *model.PoolBlockInfoModel) (*model.PoolBlockInfoModel, error) {
	collection, err := model.PoolBlockInfoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "bigBlockId")
	CreateIndex(collection, "bigBlock")
	CreateIndex(collection, "cycleBlock")
	CreateIndex(collection, "timestamp")

	entity := &model.PoolBlockInfoModel{
		Id:         util.GetUUIDInt64(),
		BigBlockId: poolBlockInfo.BigBlockId,
		BigBlock:   poolBlockInfo.BigBlock,
		StartBlock: poolBlockInfo.StartBlock,
		CycleBlock: poolBlockInfo.CycleBlock,
		Timestamp:  poolBlockInfo.Timestamp,
		CreateTime: util.Time(),
		State:      model.STATE_EXIST,
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

func SetPoolBlockInfoModel(poolBlockInfo *model.PoolBlockInfoModel) (*model.PoolBlockInfoModel, error) {
	entity, err := FindPoolBlockInfoModelByBigBlockId(poolBlockInfo.BigBlockId)
	if err == nil && entity != nil {
		collection, err := model.PoolBlockInfoModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"bigBlockId", poolBlockInfo.BigBlockId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "bigBlockId", Value: poolBlockInfo.BigBlockId})
		bsonData = append(bsonData, bson.E{Key: "bigBlock", Value: poolBlockInfo.BigBlock})
		bsonData = append(bsonData, bson.E{Key: "startBlock", Value: poolBlockInfo.StartBlock})
		bsonData = append(bsonData, bson.E{Key: "cycleBlock", Value: poolBlockInfo.CycleBlock})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: poolBlockInfo.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return poolBlockInfo, nil
	} else {
		return createPoolBlockInfoModel(poolBlockInfo)
	}
}
func FindNewestPoolBlockInfoModelByCycleBlock(cycleBlock int64) (*model.PoolBlockInfoModel, error) {
	collection, err := model.PoolBlockInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"cycleBlock", cycleBlock},
		//{"state", model.STATE_EXIST},
	}
	sort := options.FindOne().SetSort(bson.M{"bigBlock": -1})

	entity := &model.PoolBlockInfoModel{}
	err = collection.FindOne(context.TODO(), queryBson, sort).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}
