package mongo_service

import (
	"context"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
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
	CreateIndex(collection, "address")
	CreateIndex(collection, "poolType")
	CreateIndex(collection, "poolState")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "dealTime")

	entity := &model.PoolBrc20Model{
		Id:         util.GetUUIDInt64(),
		Net:        poolBrc20.Net,
		OrderId:    poolBrc20.OrderId,
		Tick:       poolBrc20.Tick,
		Pair:       poolBrc20.Pair,
		Amount:     poolBrc20.Amount,
		Address:    poolBrc20.Address,
		PsbtRaw:    poolBrc20.PsbtRaw,
		PoolType:   poolBrc20.PoolType,
		PoolState:  poolBrc20.PoolState,
		DealTime:   poolBrc20.DealTime,
		Timestamp:  poolBrc20.Timestamp,
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
		bsonData = append(bsonData, bson.E{Key: "amount", Value: poolBrc20.Amount})
		bsonData = append(bsonData, bson.E{Key: "address", Value: poolBrc20.Address})
		bsonData = append(bsonData, bson.E{Key: "psbtRaw", Value: poolBrc20.PsbtRaw})
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
