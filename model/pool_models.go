package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type PoolType int
type PoolState int

const (
	PoolTypeTick PoolType = 1
	PoolTypeBtc  PoolType = 2
	PoolTypeBoth PoolType = 3

	PoolStateAdd    PoolState = 1
	PoolStateRemove PoolState = 2
	PoolStateUsed   PoolState = 3
	PoolStateClaim  PoolState = 4
)

type PoolBrc20Model struct {
	Id             int64     `json:"id" bson:"_id" tb:"pool_brc20_model" mg:"true"`
	Net            string    `json:"net" bson:"net"`
	OrderId        string    `json:"orderId" bson:"orderId"`
	Tick           string    `json:"tick" bson:"tick"`
	Pair           string    `json:"pair" bson:"pair"`
	CoinAmount     uint64    `json:"coinAmount" bson:"coinAmount"`
	CoinDecimalNum int       `json:"coinDecimalNum" bson:"coinDecimalNum"`
	Amount         uint64    `json:"amount" bson:"amount"`
	DecimalNum     int       `json:"decimalNum" bson:"decimalNum"` //decimal num
	CoinRatePrice  uint64    `json:"coinRatePrice" bson:"coinRatePrice"`
	CoinAddress    string    `json:"coinAddress" bson:"coinAddress"`
	Address        string    `json:"address" bson:"address"`
	CoinPsbtRaw    string    `json:"coinPsbtRaw" bson:"coinPsbtRaw"`
	PsbtRaw        string    `json:"psbtRaw" bson:"psbtRaw"`
	InscriptionId  string    `json:"inscriptionId" bson:"inscriptionId"` //InscriptionId
	UtxoId         string    `json:"utxoId" bson:"utxoId"`               //UtxoId
	PoolType       PoolType  `json:"poolType" bson:"poolType"`
	PoolState      PoolState `json:"poolState" bson:"poolState"`
	DealTime       int64     `json:"dealTime" bson:"dealTime"`
	Timestamp      int64     `json:"timestamp" bson:"timestamp"`
	CreateTime     int64     `json:"createTime" bson:"createTime"`
	UpdateTime     int64     `json:"updateTime" bson:"updateTime"`
	State          int64     `json:"state" bson:"state"`
}

func (s PoolBrc20Model) getCollection() string {
	return "pool_brc20_model"
}

func (s PoolBrc20Model) getDB() string {
	return major.DsOrdbook
}

func (s PoolBrc20Model) GetReadDB() (*mongo.Collection, error) {
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

func (s PoolBrc20Model) GetWriteDB() (*mongo.Collection, error) {
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

type PoolInfoModel struct {
	Id             int64  `json:"id" bson:"_id" tb:"pool_info_model" mg:"true"`
	Net            string `json:"net" bson:"net"`
	Tick           string `json:"tick" bson:"tick"`
	Pair           string `json:"pair" bson:"pair"`
	CoinAmount     uint64 `json:"coinAmount" bson:"coinAmount"`
	CoinDecimalNum int    `json:"coinDecimalNum" bson:"coinDecimalNum"` //omitempty
	Amount         uint64 `json:"amount" bson:"amount"`
	DecimalNum     int    `json:"decimalNum" bson:"decimalNum"`
	Timestamp      int64  `json:"timestamp" bson:"timestamp"`
	CreateTime     int64  `json:"createTime" bson:"createTime"`
	UpdateTime     int64  `json:"updateTime" bson:"updateTime"`
	State          int64  `json:"state" bson:"state"`
}

func (s PoolInfoModel) getCollection() string {
	return "pool_info_model"
}

func (s PoolInfoModel) getDB() string {
	return major.DsOrdbook
}

func (s PoolInfoModel) GetReadDB() (*mongo.Collection, error) {
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

func (s PoolInfoModel) GetWriteDB() (*mongo.Collection, error) {
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
