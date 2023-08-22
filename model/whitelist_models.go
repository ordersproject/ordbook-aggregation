package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type WhitelistType int

const (
	WhitelistTypeClaim   WhitelistType = 1
	WhitelistTypeClaim1w WhitelistType = 2
)

type WhiteUseState int

const (
	WhiteUseStateNo  WhiteUseState = 0
	WhiteUseStateYes WhiteUseState = 1
)

type WhitelistModel struct {
	Id            int64         `json:"id" bson:"_id" tb:"whitelist_model" mg:"true"`
	AddressId     string        `json:"addressId" bson:"addressId"` //Address_whitelistType
	Address       string        `json:"address" bson:"address"`     //
	IP            string        `json:"ip" bson:"ip"`
	WhitelistType WhitelistType `json:"whitelistType" bson:"whitelistType"`
	WhiteUseState WhiteUseState `json:"whiteUseState" bson:"whiteUseState"`
	Limit         int64         `json:"limit" bson:"limit"`
	Timestamp     int64         `json:"timestamp" bson:"timestamp"`
	CreateTime    int64         `json:"createTime" bson:"createTime"`
	UpdateTime    int64         `json:"updateTime" bson:"updateTime"`
	State         int64         `json:"state" bson:"state"`
}

func (s WhitelistModel) GetCollection() string {
	return "whitelist_model"
}

func (s WhitelistModel) GetDB() string {
	return major.DsOrdbook
}

func (s WhitelistModel) GetReadDB() (*mongo.Collection, error) {
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

func (s WhitelistModel) GetWriteDB() (*mongo.Collection, error) {
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
