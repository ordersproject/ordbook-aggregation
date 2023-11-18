package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type ConfirmStatus int

const (
	Unconfirmed ConfirmStatus = 0
	Confirmed   ConfirmStatus = 1
)

type LpBrc20Model struct {
	Id                  int64         `json:"id" bson:"_id" tb:"lp_brc20_model" mg:"true"`
	Net                 string        `json:"net" bson:"net"`
	Tick                string        `json:"tick" bson:"tick"`
	OrderId             string        `json:"orderId" bson:"orderId"`
	Address             string        `json:"address" bson:"address"`
	FeeAddress          string        `json:"feeAddress" bson:"feeAddress"`
	PoolOrderId         string        `json:"poolOrderId" bson:"poolOrderId"`
	Brc20InscriptionId  string        `json:"brc20InscriptionId" bson:"brc20InscriptionId"`
	Brc20CoinAmount     int64         `json:"brc20CoinAmount" bson:"brc20CoinAmount"`
	Brc20InValue        int64         `json:"brc20InValue" bson:"brc20InValue"`
	Brc20ConfirmStatus  ConfirmStatus `json:"brc20ConfirmStatus" bson:"brc20ConfirmStatus"`
	BtcUtxoId           string        `json:"btcUtxoId" bson:"btcUtxoId"`
	BtcAmount           int64         `json:"btcAmount" bson:"btcAmount"`
	BtcOutValue         int64         `json:"btcOutValue" bson:"btcOutValue"`
	BtcConfirmStatus    ConfirmStatus `json:"btcConfirmStatus" bson:"btcConfirmStatus"`
	PoolOrderState      PoolState     `json:"poolOrderState" bson:"poolOrderState"`
	PoolOrderCoinState  PoolState     `json:"poolOrderCoinState" bson:"poolOrderCoinState"`
	PoolCoinRatePrice   uint64        `json:"poolCoinRatePrice" bson:"poolCoinRatePrice"`     //
	CoinRatePrice       uint64        `json:"coinRatePrice" bson:"coinRatePrice"`             //
	CoinPrice           int64         `json:"coinPrice" bson:"coinPrice"`                     //MAX-9223372036854775807
	CoinPriceDecimalNum int32         `json:"coinPriceDecimalNum" bson:"coinPriceDecimalNum"` //8
	Ratio               int64         `json:"ratio" bson:"ratio"`                             // ratio: 12/15/18
	Timestamp           int64         `json:"timestamp" bson:"timestamp"`
	CreateTime          int64         `json:"createTime" bson:"createTime"`
	UpdateTime          int64         `json:"updateTime" bson:"updateTime"`
	State               int64         `json:"state" bson:"state"`
}

func (s LpBrc20Model) GetCollection() string {
	return "lp_brc20_model"
}

func (s LpBrc20Model) GetDB() string {
	return major.DsOrdbook
}

func (s LpBrc20Model) GetReadDB() (*mongo.Collection, error) {
	mongoDB, err := major.GetOrderbookDb()
	if err != nil {
		return nil, err
	}
	collection := mongoDB.Database(s.GetDB()).Collection(s.GetCollection())
	if collection == nil {
		return nil, errors.New("db connect error")
	}
	return collection, nil
}

func (s LpBrc20Model) GetWriteDB() (*mongo.Collection, error) {
	mongoDB, err := major.GetOrderbookDb()
	if err != nil {
		return nil, err
	}
	collection := mongoDB.Database(s.GetDB()).Collection(s.GetCollection())
	if collection == nil {
		return nil, errors.New("db connect error")
	}
	return collection, nil
}
