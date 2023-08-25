package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type PoolType int
type PoolState int

const (
	PoolTypeTick                PoolType = 1
	PoolTypeBtc                 PoolType = 2
	PoolTypeBoth                PoolType = 3
	PoolTypeMultiSigInscription PoolType = 4

	PoolStateAdd    PoolState = 1
	PoolStateRemove PoolState = 2
	PoolStateUsed   PoolState = 3
	PoolStateClaim  PoolState = 4
)

type PoolBrc20Model struct {
	Id                    int64  `json:"id" bson:"_id" tb:"pool_brc20_model" mg:"true"`
	Net                   string `json:"net" bson:"net"`
	OrderId               string `json:"orderId" bson:"orderId"`
	OrderPairId           string `json:"orderPairId" bson:"orderPairId"`
	Tick                  string `json:"tick" bson:"tick"`
	Pair                  string `json:"pair" bson:"pair"`
	CoinAmount            uint64 `json:"coinAmount" bson:"coinAmount"`
	CoinDecimalNum        int    `json:"coinDecimalNum" bson:"coinDecimalNum"`
	CoinRatePrice         uint64 `json:"coinRatePrice" bson:"coinRatePrice"`
	CoinAddress           string `json:"coinAddress" bson:"coinAddress"`
	CoinPublicKey         string `json:"coinPublicKey" bson:"coinPublicKey"`
	CoinInputValue        uint64 `json:"coinInputValue" bson:"coinInputValue"`
	Address               string `json:"address" bson:"address"`
	MultiSigScript        string `json:"multiSigScript" bson:"multiSigScript"`
	MultiSigScriptAddress string `json:"multiSigScriptAddress" bson:"multiSigScriptAddress"`
	CoinPsbtRaw           string `json:"coinPsbtRaw" bson:"coinPsbtRaw"`
	InscriptionId         string `json:"inscriptionId" bson:"inscriptionId"`         //InscriptionId
	InscriptionNumber     string `json:"inscriptionNumber" bson:"inscriptionNumber"` //InscriptionId

	Amount     uint64   `json:"amount" bson:"amount"`
	DecimalNum int      `json:"decimalNum" bson:"decimalNum"` //decimal num
	PsbtRaw    string   `json:"psbtRaw" bson:"psbtRaw"`
	UtxoId     string   `json:"utxoId" bson:"utxoId"` //UtxoId
	PoolType   PoolType `json:"poolType" bson:"poolType"`

	PoolState      PoolState `json:"poolState" bson:"poolState"`
	DealTx         string    `json:"dealTx" bson:"dealTx"`
	DealTxIndex    int64     `json:"dealTxIndex" bson:"dealTxIndex"`
	DealTxOutValue int64     `json:"dealTxOutValue" bson:"dealTxOutValue"`
	DealTime       int64     `json:"dealTime" bson:"dealTime"`

	PoolCoinState      PoolState `json:"poolCoinState" bson:"poolCoinState"`
	DealCoinTx         string    `json:"dealCoinTx" bson:"dealCoinTx"`
	DealCoinTxIndex    int64     `json:"dealCoinTxIndex" bson:"dealCoinTxIndex"`
	DealCoinTxOutValue int64     `json:"dealCoinTxOutValue" bson:"dealCoinTxOutValue"`
	DealCoinTime       int64     `json:"dealCoinTime" bson:"dealCoinTime"`

	DealInscriptionId         string `json:"dealInscriptionId" bson:"dealInscriptionId"` //InscriptionId
	DealInscriptionTx         string `json:"dealInscriptionTx" bson:"dealInscriptionTx"`
	DealInscriptionTxIndex    int64  `json:"dealInscriptionTxIndex" bson:"dealInscriptionTxIndex"`
	DealInscriptionTxOutValue int64  `json:"dealInscriptionTxOutValue" bson:"dealInscriptionTxOutValue"`
	DealInscriptionTime       int64  `json:"dealInscriptionTime" bson:"dealInscriptionTime"`

	ClaimTx    string `json:"claimTx" bson:"claimTx"`
	Timestamp  int64  `json:"timestamp" bson:"timestamp"`
	CreateTime int64  `json:"createTime" bson:"createTime"`
	UpdateTime int64  `json:"updateTime" bson:"updateTime"`
	State      int64  `json:"state" bson:"state"`
}

type PoolOrderCount struct {
	Id              string `json:"id" bson:"_id"`
	CoinAmountTotal int64  `json:"coinAmountTotal" bson:"coinAmountTotal"`
	AmountTotal     int64  `json:"amountTotal" bson:"amountTotal"`
	OrderCounts     int64  `json:"orderCounts" bson:"orderCounts"`
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
