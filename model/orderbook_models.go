package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type OrderBrc20Model struct {
	Id             int64      `json:"id" bson:"_id" tb:"order_brc20_model" mg:"true"`
	OrderId        string     `json:"orderId" bson:"orderId"`
	Tick           string     `json:"tick" bson:"tick"`
	Amount         uint64     `json:"amount" bson:"amount"`
	DecimalNum     int        `json:"decimalNum" bson:"decimalNum"`
	CoinAmount     uint64     `json:"coinAmount" bson:"coinAmount"`
	CoinDecimalNum int        `json:"coinDecimalNum" bson:"coinDecimalNum"`
	CoinRatePrice  uint64     `json:"coinRatePrice" bson:"coinRatePrice"`
	OrderState     OrderState `json:"orderState" bson:"orderState"` //1-create,2-finish,3-cancel
	OrderType      OrderType  `json:"orderType" bson:"orderType"`   //1-sell,2-buy
	SellerAddress  string     `json:"sellerAddress" bson:"sellerAddress"`
	BuyerAddress   string     `json:"buyerAddress" bson:"buyerAddress"`
	PsbtRaw        string     `json:"psbtRaw" bson:"psbtRaw"`
	Timestamp      int64      `json:"timestamp" bson:"timestamp"`
	CreateTime     int64      `json:"createTime" bson:"createTime"`
	UpdateTime     int64      `json:"updateTime" bson:"updateTime"`
	State          int64      `json:"state" bson:"state"`
}

func (s OrderBrc20Model) getCollection() string {
	return "order_brc20_model"
}

func (s OrderBrc20Model) getDB() string {
	return major.DsOrderbook
}

func (s OrderBrc20Model) GetReadDB() (*mongo.Collection, error) {
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

func (s OrderBrc20Model) GetWriteDB() (*mongo.Collection, error) {
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