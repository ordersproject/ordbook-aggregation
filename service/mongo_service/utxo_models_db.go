package mongo_service

import (
	"context"
	"errors"
	"fmt"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"time"
)

func FindOrderUtxoModelByUtxorId(utxoId string) (*model.OrderUtxoModel, error) {
	collection, err := model.OrderUtxoModel{}.GetReadDB()
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

func createOrderUtxoModel(orderUtxo *model.OrderUtxoModel) (*model.OrderUtxoModel, error) {
	collection, err := model.OrderUtxoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "utxoId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "utxoType")
	CreateIndex(collection, "txId")
	CreateIndex(collection, "amount")
	CreateIndex(collection, "index")
	CreateIndex(collection, "used")
	CreateIndex(collection, "useTx")
	CreateIndex(collection, "sortIndex")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "confirmStatus")
	CreateIndex(collection, "fromOrderId")

	entity := &model.OrderUtxoModel{
		Id:             util.GetUUIDInt64(),
		Net:            orderUtxo.Net,
		UtxoId:         orderUtxo.UtxoId,
		UtxoType:       orderUtxo.UtxoType,
		Amount:         orderUtxo.Amount,
		Address:        orderUtxo.Address,
		PrivateKeyHex:  orderUtxo.PrivateKeyHex,
		TxId:           orderUtxo.TxId,
		Index:          orderUtxo.Index,
		PkScript:       orderUtxo.PkScript,
		UsedState:      orderUtxo.UsedState,
		UseTx:          orderUtxo.UseTx,
		SortIndex:      orderUtxo.SortIndex,
		Timestamp:      orderUtxo.Timestamp,
		ConfirmStatus:  orderUtxo.ConfirmStatus,
		FromOrderId:    orderUtxo.FromOrderId,
		NetworkFeeRate: orderUtxo.NetworkFeeRate,
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

func SetOrderUtxoModel(orderUtxo *model.OrderUtxoModel) (*model.OrderUtxoModel, error) {
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
		bsonData = append(bsonData, bson.E{Key: "confirmStatus", Value: orderUtxo.ConfirmStatus})
		bsonData = append(bsonData, bson.E{Key: "fromOrderId", Value: orderUtxo.FromOrderId})
		bsonData = append(bsonData, bson.E{Key: "networkFeeRate", Value: orderUtxo.NetworkFeeRate})
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

func UpdateOrderUtxoModelForUsed(utxoId, useTx string, UsedState model.UsedState) error {
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

func UpdateOrderUtxoModelForOccupied(utxoId, orderId string, UsedState model.UsedState) error {
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
		bsonData = append(bsonData, bson.E{Key: "orderId", Value: orderId})
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

func UpdateOrderUtxoModelForConfirm(utxoId string, confirmState model.ConfirmStatus) error {
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
		bsonData = append(bsonData, bson.E{Key: "confirmStatus", Value: confirmState})
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

func CountUtxoList(net string, perAmount int64, utxoType model.UtxoType) (int64, error) {
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"net":      net,
		"utxoType": utxoType,
		"used":     model.UsedNo,
		"state":    model.STATE_EXIST,
	}

	if perAmount != 0 {
		find["amount"] = perAmount
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindUtxoList(net string, startIndex, limit, perAmount int64, utxoType model.UtxoType, confirmStatus model.ConfirmStatus, fromOrderId string, networkFeeRate int64) ([]*model.OrderUtxoModel, error) {
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":      net,
		"utxoType": utxoType,
		"used":     model.UsedNo,
		"state":    model.STATE_EXIST,
	}
	start := bson.M{GT_: startIndex}
	find["sortIndex"] = start

	if perAmount != 0 {
		find["amount"] = perAmount
	}
	if confirmStatus != -1 {
		find["confirmStatus"] = confirmStatus
	}
	if fromOrderId != "" {
		find["fromOrderId"] = fromOrderId
	}
	if networkFeeRate != 0 {
		find["networkFeeRate"] = bson.M{
			GT_: networkFeeRate - 50,
			LT_: networkFeeRate + 200,
		}
	}

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

func FindAllTypeUtxoList(net string, startIndex, limit, perAmount int64, confirmStatus model.ConfirmStatus) ([]*model.OrderUtxoModel, error) {
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":   net,
		"used":  model.UsedNo,
		"state": model.STATE_EXIST,
	}
	//start := bson.M{GT_: startIndex}
	//find["sortIndex"] = start

	if perAmount != 0 {
		find["amount"] = perAmount
	}
	if confirmStatus != -1 {
		find["confirmStatus"] = confirmStatus
	}

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

func GetLatestStartIndexUtxo(net string, utxoType model.UtxoType, perAmount int64) (*model.OrderUtxoModel, error) {
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}
	find := bson.M{
		"net":      net,
		"utxoType": utxoType,
		"used":     model.UsedNo,
		"state":    model.STATE_EXIST,
	}
	if perAmount != 0 {
		find["amount"] = perAmount
	}
	sort := options.FindOne().SetSort(bson.M{"sortIndex": -1})
	entity := &model.OrderUtxoModel{}
	err = collection.FindOne(context.TODO(), find, sort).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func SetManyUtxoInSession(utxoList []*model.OrderUtxoModel, jop func() error) error {
	mongoDB, err := major.GetOrderbookDb()
	if err != nil {
		return err
	}

	timeout := time.Duration(3000) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err = mongoDB.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		if err := sessionContext.StartTransaction(); err != nil {
			return err
		}
		collection := mongoDB.Database(model.OrderUtxoModel{}.GetDB()).Collection(model.OrderUtxoModel{}.GetCollection())

		for _, orderUtxo := range utxoList {
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
			if _, err = collection.InsertOne(context.Background(), entity); err != nil {
				if err := sessionContext.AbortTransaction(context.Background()); err != nil {
					fmt.Printf("mongo transaction rollback failed, %s\n", err.Error())
					return err
				}
				return err
			}
			fmt.Printf("InsertOne in mongo transaction success\n")
		}

		//if _, err := collection.InsertMany(sessionContext, utxoList); err != nil {
		//	if err := sessionContext.AbortTransaction(context.Background()); err != nil {
		//		fmt.Printf("mongo transaction rollback failed, %s\n", err.Error())
		//		return err
		//	}
		//	return err
		//}

		if err := jop(); err != nil {
			if err := sessionContext.AbortTransaction(context.Background()); err != nil {
				fmt.Printf("mongo transaction rollback failed, %s\n", err.Error())
				return err
			}
			return err
		}

		if err := sessionContext.CommitTransaction(context.Background()); err != nil {
			fmt.Printf("mongo transaction commit failed, %s\n", err.Error())
			return err
		}
		return nil
	}); err != nil {
		fmt.Printf("insert failed, err:%s\n", err.Error())
		return err
	}
	return nil
}

func FindAllUtxoList(net string, limit int64, utxoType model.UtxoType, useState model.UsedState, confirmStatus model.ConfirmStatus) ([]*model.OrderUtxoModel, error) {
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":      net,
		"utxoType": utxoType,
		"state":    model.STATE_EXIST,
	}
	find["used"] = useState
	if confirmStatus != -1 {
		find["confirmStatus"] = confirmStatus
	}

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

func FindOccupiedUtxoListByOrderId(net, orderId string, limit int64, useState model.UsedState) ([]*model.OrderUtxoModel, error) {
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":     net,
		"orderId": orderId,
		"state":   model.STATE_EXIST,
	}
	find["used"] = useState

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

func FindUtxoListByTxId(net, txId string, limit int64, utxoType model.UtxoType, useState model.UsedState, confirmStatus model.ConfirmStatus) ([]*model.OrderUtxoModel, error) {
	collection, err := model.OrderUtxoModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"net":      net,
		"txId":     txId,
		"utxoType": utxoType,
		"state":    model.STATE_EXIST,
	}
	find["used"] = useState
	if confirmStatus != -1 {
		find["confirmStatus"] = confirmStatus
	}

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
