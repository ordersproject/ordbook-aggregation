package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type OrderBrc20Model struct {
	Id                int64      `json:"id" bson:"_id" tb:"order_brc20_model" mg:"true"`
	Net               string     `json:"net" bson:"net"`
	OrderId           string     `json:"orderId" bson:"orderId"`
	Tick              string     `json:"tick" bson:"tick"`
	Amount            uint64     `json:"amount" bson:"amount"`
	DecimalNum        int        `json:"decimalNum" bson:"decimalNum"`
	CoinAmount        uint64     `json:"coinAmount" bson:"coinAmount"`
	CoinDecimalNum    int        `json:"coinDecimalNum" bson:"coinDecimalNum"`
	CoinRatePrice     uint64     `json:"coinRatePrice" bson:"coinRatePrice"`
	OrderState        OrderState `json:"orderState" bson:"orderState"` //1-create,2-finish,3-cancel
	OrderType         OrderType  `json:"orderType" bson:"orderType"`   //1-sell,2-buy
	SellerAddress     string     `json:"sellerAddress" bson:"sellerAddress"`
	BuyerAddress      string     `json:"buyerAddress" bson:"buyerAddress"`
	MarketAmount      uint64     `json:"marketAmount" bson:"marketAmount"`
	PlatformTx        string     `json:"platformTx" bson:"platformTx"`
	InscriptionId     string     `json:"inscriptionId" bson:"inscriptionId"`
	InscriptionNumber string     `json:"inscriptionNumber" bson:"inscriptionNumber"`
	PsbtRawPreAsk     string     `json:"psbtRawPreAsk" bson:"psbtRawPreAsk"`
	PsbtRawFinalAsk   string     `json:"psbtRawFinalAsk" bson:"psbtRawFinalAsk"`
	PsbtAskTxId       string     `json:"psbtAskTxId" bson:"psbtAskTxId"`
	PsbtRawPreBid     string     `json:"psbtRawPreBid" bson:"psbtRawPreBid"`
	PsbtRawMidBid     string     `json:"psbtRawMidBid" bson:"psbtRawMidBid"`
	PsbtRawFinalBid   string     `json:"psbtRawBidFinalBid" bson:"psbtRawBidFinalBid"`
	PsbtBidTxId       string     `json:"psbtBidTxId" bson:"psbtBidTxId"`
	Integral          int64      `json:"integral" bson:"integral"`
	Timestamp         int64      `json:"timestamp" bson:"timestamp"`
	CreateTime        int64      `json:"createTime" bson:"createTime"`
	UpdateTime        int64      `json:"updateTime" bson:"updateTime"`
	State             int64      `json:"state" bson:"state"`
}

func (s OrderBrc20Model) getCollection() string {
	return "order_brc20_model"
}

func (s OrderBrc20Model) getDB() string {
	return major.DsOrdbook
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

type DummyState int

const (
	DummyStateLive DummyState = 1
	DummyStateCancel DummyState = 2
	DummyStateFinish DummyState = 3
)

type OrderBrc20BidDummyModel struct {
	Id         int64      `json:"id" bson:"_id" tb:"order_brc20_bid_dummy_model" mg:"true"`
	Net        string     `json:"net" bson:"net"`
	DummyId    string     `json:"dummyId" bson:"dummyId"` //txId:index
	OrderId    string     `json:"orderId" bson:"orderId"`
	Tick       string     `json:"tick" bson:"tick"`
	Address    string     `json:"address" bson:"address"`
	DummyState DummyState `json:"dummyState" bson:"dummyState"`
	Timestamp  int64      `json:"timestamp" bson:"timestamp"`
	CreateTime int64      `json:"createTime" bson:"createTime"`
	UpdateTime int64      `json:"updateTime" bson:"updateTime"`
	State      int64      `json:"state" bson:"state"`
}

func (s OrderBrc20BidDummyModel) getCollection() string {
	return "order_brc20_bid_dummy_model"
}

func (s OrderBrc20BidDummyModel) getDB() string {
	return major.DsOrdbook
}

func (s OrderBrc20BidDummyModel) GetReadDB() (*mongo.Collection, error) {
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

func (s OrderBrc20BidDummyModel) GetWriteDB() (*mongo.Collection, error) {
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

