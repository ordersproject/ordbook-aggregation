package mongo_service

import (
	"context"
	"errors"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/model"
)

func FindBrc20TickModelByPair(net, pair string) (*model.Brc20TickModel, error) {
	collection, err := model.Brc20TickModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"pair", pair},
		//{"state", model.STATE_EXIST},
	}
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

	CreateUniqueIndex(collection, "pair")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "volume")
	CreateIndex(collection, "timestamp")

	entity := &model.Brc20TickModel{
		Id:                 util.GetUUIDInt64(),
		Net:                brc20Tick.Net,
		Tick:               brc20Tick.Tick,
		Pair:               brc20Tick.Pair,
		Buy:                brc20Tick.Buy,
		Sell:               brc20Tick.Sell,
		Low:                brc20Tick.Low,
		High:               brc20Tick.High,
		Open:               brc20Tick.Open,
		Last:               brc20Tick.Last,
		Volume:             brc20Tick.Volume,
		Amount:             brc20Tick.Amount,
		Vol:                brc20Tick.Vol,
		AvgPrice:           brc20Tick.AvgPrice,
		QuoteSymbol:        brc20Tick.QuoteSymbol,
		PriceChangePercent: brc20Tick.PriceChangePercent,
		CreateTime:         util.Time(),
		State:              model.STATE_EXIST,
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
	entity, err := FindBrc20TickModelByPair(brc20Tick.Net, brc20Tick.Pair)
	if err == nil && entity != nil {
		collection, err := model.Brc20TickModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"net", brc20Tick.Net},
			{"pair", brc20Tick.Pair},
			//{"state", model.STATE_EXIST},
		}
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
		bsonData = append(bsonData, bson.E{Key: "volume", Value: brc20Tick.Volume})
		bsonData = append(bsonData, bson.E{Key: "amount", Value: brc20Tick.Amount})
		bsonData = append(bsonData, bson.E{Key: "vol", Value: brc20Tick.Vol})
		bsonData = append(bsonData, bson.E{Key: "avgPrice", Value: brc20Tick.AvgPrice})
		bsonData = append(bsonData, bson.E{Key: "quoteSymbol", Value: brc20Tick.QuoteSymbol})
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
