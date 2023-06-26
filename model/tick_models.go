package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type Brc20TickModel struct {
	Id                 int64   `json:"id" bson:"_id" tb:"brc20_tick_model" mg:"true"`
	Net                string  `json:"net" bson:"net"`
	Tick               string  `json:"tick" bson:"tick"`
	Pair               string  `json:"pair" bson:"pair"`                             //
	Buy                uint64  `json:"buy" bson:"buy"`                               //
	Sell               uint64  `json:"sell" bson:"sell"`                             //
	Low                uint64  `json:"low" bson:"low"`                               //
	High               uint64  `json:"high" bson:"high"`                             //
	Open               uint64  `json:"open" bson:"open"`                             //
	Last               uint64  `json:"last" bson:"last"`                             //
	Volume             uint64  `json:"volume" bson:"volume"`                         //
	Amount             uint64  `json:"amount" bson:"amount"`                         //
	Vol                uint64  `json:"vol" bson:"vol"`                               //
	AvgPrice           uint64  `json:"avgPrice" bson:"avgPrice"`                     //
	QuoteSymbol        string  `json:"quoteSymbol" bson:"quoteSymbol"`               //
	PriceChangePercent float64 `json:"priceChangePercent" bson:"priceChangePercent"` //
	CreateTime         int64   `json:"createTime" bson:"createTime"`
	UpdateTime         int64   `json:"updateTime" bson:"updateTime"`
	State              int64   `json:"state" bson:"state"`
}

func (s Brc20TickModel) getCollection() string {
	return "brc20_tick_model"
}

func (s Brc20TickModel) getDB() string {
	return major.DsOrdbook
}

func (s Brc20TickModel) GetReadDB() (*mongo.Collection, error) {
	mongoDB, err := major.GetOrderbookDb()
	if err != nil {
		return nil, err
	}
	collection := mongoDB.Database(s.getDB()).Collection(s.getCollection())
	if collection == nil {
		return nil, errors.New("db connect error")
	}
	return collection, nil
}

func (s Brc20TickModel) GetWriteDB() (*mongo.Collection, error) {
	mongoDB, err := major.GetOrderbookDb()
	if err != nil {
		return nil, err
	}
	collection := mongoDB.Database(s.getDB()).Collection(s.getCollection())
	if collection == nil {
		return nil, errors.New("db connect error")
	}
	return collection, nil
}

type Brc20TickInfoModel struct {
	Id             int64  `json:"id" bson:"_id" tb:"brc20_tick_info_model" mg:"true"`
	Net            string `json:"net" bson:"net"`
	Tick           string `json:"tick" bson:"tick"`
	Name           string `json:"name" bson:"name"`                     //
	Decimal        string `json:"decimal" bson:"decimal"`               //
	Supply         string `json:"supply" bson:"supply"`                 //
	Icon           string `json:"icon" bson:"icon"`                     //
	DefaultLimit   string `json:"defaultLimit" bson:"defaultLimit"`     //
	Deployer       string `json:"deployer" bson:"deployer"`             //
	DeployTime     string `json:"deployTime" bson:"deployTime"`         //
	DeployContract string `json:"deployContract" bson:"deployContract"` //
	Description    string `json:"description" bson:"description"`       //
	CreateTime     int64  `json:"createTime" bson:"createTime"`
	UpdateTime     int64  `json:"updateTime" bson:"updateTime"`
	State          int64  `json:"state" bson:"state"`
}

func (s Brc20TickInfoModel) getCollection() string {
	return "brc20_tick_info_model"
}

func (s Brc20TickInfoModel) getDB() string {
	return major.DsOrdbook
}

func (s Brc20TickInfoModel) GetReadDB() (*mongo.Collection, error) {
	mongoDB, err := major.GetOrderbookDb()
	if err != nil {
		return nil, err
	}
	collection := mongoDB.Database(s.getDB()).Collection(s.getCollection())
	if collection == nil {
		return nil, errors.New("db connect error")
	}
	return collection, nil
}

func (s Brc20TickInfoModel) GetWriteDB() (*mongo.Collection, error) {
	mongoDB, err := major.GetOrderbookDb()
	if err != nil {
		return nil, err
	}
	collection := mongoDB.Database(s.getDB()).Collection(s.getCollection())
	if collection == nil {
		return nil, errors.New("db connect error")
	}
	return collection, nil
}
