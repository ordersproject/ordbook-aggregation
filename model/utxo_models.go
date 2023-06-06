package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type UtxoType int

const (
	UtxoTypeDummy UtxoType = 1
	UtxoTypeBidY  UtxoType = 2
	UtxoTypeFakerInscription  UtxoType = 6
)

type UsedState int

const (
	UsedNo  UsedState = 1
	UsedYes UsedState = 2
	UsedErr UsedState = 3
	UsedDel UsedState = 4
)

type OrderUtxoModel struct {
	Id         int64     `json:"id" bson:"_id" tb:"order_utxo_model" mg:"true"`
	UtxoId     string    `json:"utxoId" bson:"utxoId"` //txId_index
	Net        string    `json:"net" bson:"net"`
	UtxoType   UtxoType  `json:"utxoType" bson:"utxoType"`
	Amount     uint64    `json:"amount" bson:"amount"`
	Address string `json:"address" bson:"address"`
	PrivateKeyHex string `json:"privateKeyHex" bson:"privateKeyHex"`
	TxId       string    `json:"txId" bson:"txId"`
	Index      int64     `json:"index" bson:"index"`
	PkScript   string    `json:"pkScript" bson:"pkScript"`
	UsedState  UsedState `json:"used" bson:"used"`
	UseTx      string    `json:"useTx" bson:"useTx"`
	SortIndex  int64     `json:"sortIndex" bson:"sortIndex"`
	Timestamp  int64     `json:"timestamp" bson:"timestamp"`
	CreateTime int64     `json:"createTime" bson:"createTime"`
	UpdateTime int64     `json:"updateTime" bson:"updateTime"`
	State      int64     `json:"state" bson:"state"`
}

func (s OrderUtxoModel) getCollection() string {
	return "order_utxo_model"
}

func (s OrderUtxoModel) getDB() string {
	return major.DsOrdbook
}

func (s OrderUtxoModel) GetReadDB() (*mongo.Collection, error) {
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

func (s OrderUtxoModel) GetWriteDB() (*mongo.Collection, error) {
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