package mongo_service

import (
	"context"
	"errors"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/model"
)

func FindOrderUtxoModelByUtxorId(utxoId string) (*model.OrderUtxoModel, error) {
	collection, err :=  model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"utxoId", utxoId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderUtxoModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}


func createOrderUtxoModel(orderUtxo *model.OrderUtxoModel) (*model.OrderUtxoModel, error)  {
	collection, err := model.OrderUtxoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "utxoId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "utxoType")
	CreateIndex(collection, "txId")
	CreateIndex(collection, "index")
	CreateIndex(collection, "used")
	CreateIndex(collection, "useTx")
	CreateIndex(collection, "sortIndex")
	CreateIndex(collection, "timestamp")

	entity := &model.OrderUtxoModel{
		Id:            util.GetUUIDInt64(),
		Net:           orderUtxo.Net,
		UtxoId:        orderUtxo.UtxoId,
		UtxoType:      orderUtxo.UtxoType,
		Amount:        orderUtxo.Amount,
		Address:       orderUtxo.Address,
		PrivateKeyHex: orderUtxo.PrivateKeyHex,
		TxId:          orderUtxo.TxId,
		Index:         orderUtxo.Index,
		PkScript:      orderUtxo.PkScript,
		UsedState:     orderUtxo.UsedState,
		UseTx:         orderUtxo.UseTx,
		SortIndex:     orderUtxo.SortIndex,
		Timestamp:     orderUtxo.Timestamp,
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

func SetOrderUtxoModel(orderUtxo *model.OrderUtxoModel) (*model.OrderUtxoModel, error)  {
	entity, err := FindOrderUtxoModelByUtxorId(orderUtxo.UtxoId)
	if err == nil && entity != nil {
		collection, err := model.OrderUtxoModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"utxoId", orderUtxo.UtxoId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: orderUtxo.Net})
		bsonData = append(bsonData, bson.E{Key: "utxoId", Value: orderUtxo.UtxoId})
		bsonData = append(bsonData, bson.E{Key: "utxoType", Value: orderUtxo.UtxoType})
		bsonData = append(bsonData, bson.E{Key: "amount", Value: orderUtxo.Amount})
		bsonData = append(bsonData, bson.E{Key: "address", Value: orderUtxo.Address})
		bsonData = append(bsonData, bson.E{Key: "privateKeyHex", Value: orderUtxo.PrivateKeyHex})
		bsonData = append(bsonData, bson.E{Key: "txId", Value: orderUtxo.TxId})
		bsonData = append(bsonData, bson.E{Key: "index", Value: orderUtxo.Index})
		bsonData = append(bsonData, bson.E{Key: "pkScript", Value: orderUtxo.PkScript})
		bsonData = append(bsonData, bson.E{Key: "used", Value: orderUtxo.UsedState})
		bsonData = append(bsonData, bson.E{Key: "useTx", Value: orderUtxo.UseTx})
		bsonData = append(bsonData, bson.E{Key: "sortIndex", Value: orderUtxo.SortIndex})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: orderUtxo.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return orderUtxo, nil
	} else {
		return createOrderUtxoModel(orderUtxo)
	}
}


func UpdateOrderUtxoModelForUsed(utxoId, useTx string, UsedState  model.UsedState) error {
	entity, err := FindOrderUtxoModelByUtxorId(utxoId)
	if err == nil && entity != nil {
		collection, err := model.OrderUtxoModel{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"utxoId", utxoId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "used", Value: UsedState})
		bsonData = append(bsonData, bson.E{Key: "useTx", Value: useTx})
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



func FindUtxoList(net string, startIndex, limit int64, utxoType model.UtxoType) ([]*model.OrderUtxoModel, error){
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":  net,
		"utxoType":  utxoType,
		"used":  model.UsedNo,
		"state": model.STATE_EXIST,
	}
	start := bson.M{GT_:startIndex}
	find["sortIndex"] = start

	models := make([]*model.OrderUtxoModel, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(0)
	sort := options.Find().SetSort(bson.M{"sortIndex": 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderUtxoModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderUtxoModel Error")
	}
	return models, nil
}

func GetLatestStartIndexUtxo(net string, utxoType model.UtxoType) (*model.OrderUtxoModel, error){
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}
	find := bson.M{
		"net":  net,
		"utxoType":  utxoType,
		"used":  model.UsedNo,
		"state": model.STATE_EXIST,
	}
	sort := options.FindOne().SetSort(bson.M{"sortIndex": -1})
	entity := &model.OrderUtxoModel{}
	err = collection.FindOne(context.TODO(), find, sort).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}