package mongo_service

import (
	"context"
	"errors"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/model"
)

func FindLpBrc20ModelByOrderId(orderId string) (*model.LpBrc20Model, error) {
	collection, err := model.LpBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"orderId", orderId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.LpBrc20Model{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createLpBrc20Model(lpBrc20 *model.LpBrc20Model) (*model.LpBrc20Model, error) {
	collection, err := model.LpBrc20Model{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "orderId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "poolOrderId")
	CreateIndex(collection, "feeAddress")
	CreateIndex(collection, "address")
	CreateIndex(collection, "inscriptionId")
	CreateIndex(collection, "btcUtxoId")
	CreateIndex(collection, "coinPrice")
	CreateIndex(collection, "coinRatePrice")
	CreateIndex(collection, "timestamp")

	entity := &model.LpBrc20Model{
		Id:                  util.GetUUIDInt64(),
		Net:                 lpBrc20.Net,
		OrderId:             lpBrc20.OrderId,
		Tick:                lpBrc20.Tick,
		Address:             lpBrc20.Address,
		FeeAddress:          lpBrc20.FeeAddress,
		PoolOrderId:         lpBrc20.PoolOrderId,
		Brc20InscriptionId:  lpBrc20.Brc20InscriptionId,
		Brc20CoinAmount:     lpBrc20.Brc20CoinAmount,
		Brc20InValue:        lpBrc20.Brc20InValue,
		Brc20ConfirmStatus:  lpBrc20.Brc20ConfirmStatus,
		BtcUtxoId:           lpBrc20.BtcUtxoId,
		BtcAmount:           lpBrc20.BtcAmount,
		BtcOutValue:         lpBrc20.BtcOutValue,
		BtcConfirmStatus:    lpBrc20.BtcConfirmStatus,
		PoolOrderState:      lpBrc20.PoolOrderState,
		PoolOrderCoinState:  lpBrc20.PoolOrderCoinState,
		PoolCoinRatePrice:   lpBrc20.PoolCoinRatePrice,
		CoinRatePrice:       lpBrc20.CoinRatePrice,
		CoinPrice:           lpBrc20.CoinPrice,
		CoinPriceDecimalNum: lpBrc20.CoinPriceDecimalNum,
		Ratio:               lpBrc20.Ratio,
		Timestamp:           lpBrc20.Timestamp,
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

func SetLpBrc20Model(lpBrc20 *model.LpBrc20Model) (*model.LpBrc20Model, error) {
	entity, err := FindLpBrc20ModelByOrderId(lpBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.LpBrc20Model{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"orderId", lpBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: lpBrc20.Net})
		bsonData = append(bsonData, bson.E{Key: "orderId", Value: lpBrc20.OrderId})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: lpBrc20.Tick})
		bsonData = append(bsonData, bson.E{Key: "address", Value: lpBrc20.Address})
		bsonData = append(bsonData, bson.E{Key: "feeAddress", Value: lpBrc20.FeeAddress})
		bsonData = append(bsonData, bson.E{Key: "poolOrderId", Value: lpBrc20.PoolOrderId})
		bsonData = append(bsonData, bson.E{Key: "brc20InscriptionId", Value: lpBrc20.Brc20InscriptionId})
		bsonData = append(bsonData, bson.E{Key: "brc20CoinAmount", Value: lpBrc20.Brc20CoinAmount})
		bsonData = append(bsonData, bson.E{Key: "brc20InValue", Value: lpBrc20.Brc20InValue})
		bsonData = append(bsonData, bson.E{Key: "brc20ConfirmStatus", Value: lpBrc20.Brc20ConfirmStatus})
		bsonData = append(bsonData, bson.E{Key: "btcUtxoId", Value: lpBrc20.BtcUtxoId})
		bsonData = append(bsonData, bson.E{Key: "btcAmount", Value: lpBrc20.BtcAmount})
		bsonData = append(bsonData, bson.E{Key: "btcOutValue", Value: lpBrc20.BtcOutValue})
		bsonData = append(bsonData, bson.E{Key: "btcConfirmStatus", Value: lpBrc20.BtcConfirmStatus})
		bsonData = append(bsonData, bson.E{Key: "poolOrderState", Value: lpBrc20.PoolOrderState})
		bsonData = append(bsonData, bson.E{Key: "poolOrderCoinState", Value: lpBrc20.PoolOrderCoinState})
		bsonData = append(bsonData, bson.E{Key: "poolCoinRatePrice", Value: lpBrc20.PoolCoinRatePrice})
		bsonData = append(bsonData, bson.E{Key: "coinRatePrice", Value: lpBrc20.CoinRatePrice})
		bsonData = append(bsonData, bson.E{Key: "coinPrice", Value: lpBrc20.CoinPrice})
		bsonData = append(bsonData, bson.E{Key: "coinPriceDecimalNum", Value: lpBrc20.CoinPriceDecimalNum})
		bsonData = append(bsonData, bson.E{Key: "ratio", Value: lpBrc20.Ratio})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: lpBrc20.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return lpBrc20, nil
	} else {
		return createLpBrc20Model(lpBrc20)
	}
}

func FindLpBrc20ModelList(limit, timestamp int64, poolState model.PoolState) ([]*model.LpBrc20Model, error) {
	collection, err := model.LpBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"poolOrderState": poolState,
		"state":          model.STATE_EXIST,
		"timestamp":      bson.M{GT_: timestamp},
	}
	skip := int64(0)

	models := make([]*model.LpBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"timestamp": 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.LpBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get LpBrc20Model Error")
	}
	return models, nil
}
