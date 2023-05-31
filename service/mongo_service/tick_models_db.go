package mongo_service

import (
	"context"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"ordbook-aggregation/model"
)

func FindBrc20TickModelByPair(pair string) (*model.Brc20TickModel, error) {
	collection, err :=  model.Brc20TickModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
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


func createBrc20TickModel(brc20Tick *model.Brc20TickModel) (*model.Brc20TickModel, error)  {
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

func SetBrc20TickModel(brc20Tick *model.Brc20TickModel) (*model.Brc20TickModel, error)  {
	entity, err := FindBrc20TickModelByPair(brc20Tick.Pair)
	if err == nil && entity != nil {
		collection, err := model.Brc20TickModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
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