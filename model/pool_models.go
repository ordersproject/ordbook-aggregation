package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type PoolType int
type PoolState int
type PoolMode int
type ClaimTxBlockState int
type MultiSigScriptAddressTickAvailableState int

const (
	PoolTypeTick                PoolType = 1
	PoolTypeBtc                 PoolType = 2
	PoolTypeBoth                PoolType = 3
	PoolTypeMultiSigInscription PoolType = 4
	PoolTypeAll                 PoolType = 100

	PoolStateAdd    PoolState = 1
	PoolStateRemove PoolState = 2
	PoolStateUsed   PoolState = 3
	PoolStateClaim  PoolState = 4
	PoolStateErr    PoolState = 5

	PoolModeNone    PoolMode = 0
	PoolModePsbt    PoolMode = 1
	PoolModeCustody PoolMode = 2

	ClaimTxBlockStateUnconfirmed ClaimTxBlockState = 1
	ClaimTxBlockStateConfirmed   ClaimTxBlockState = 2

	MultiSigScriptAddressTickAvailableStateNo  = 0
	MultiSigScriptAddressTickAvailableStateYes = 1
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
	CoinPrice             int64  `json:"coinPrice" bson:"coinPrice"`                     //MAX-9223372036854775807
	CoinPriceDecimalNum   int32  `json:"coinPriceDecimalNum" bson:"coinPriceDecimalNum"` //8
	CoinAddress           string `json:"coinAddress" bson:"coinAddress"`
	CoinPublicKey         string `json:"coinPublicKey" bson:"coinPublicKey"`
	CoinInputValue        uint64 `json:"coinInputValue" bson:"coinInputValue"`
	Address               string `json:"address" bson:"address"`
	MultiSigScript        string `json:"multiSigScript" bson:"multiSigScript"`
	MultiSigScriptAddress string `json:"multiSigScriptAddress" bson:"multiSigScriptAddress"`
	CoinPsbtRaw           string `json:"coinPsbtRaw" bson:"coinPsbtRaw"`
	InscriptionId         string `json:"inscriptionId" bson:"inscriptionId"`         //InscriptionId
	InscriptionNumber     string `json:"inscriptionNumber" bson:"inscriptionNumber"` //InscriptionId

	Amount                   uint64   `json:"amount" bson:"amount"`
	DecimalNum               int      `json:"decimalNum" bson:"decimalNum"` //decimal num
	MultiSigScriptBtc        string   `json:"multiSigScriptBtc" bson:"multiSigScriptBtc"`
	MultiSigScriptAddressBtc string   `json:"multiSigScriptAddressBtc" bson:"multiSigScriptAddressBtc"`
	PsbtRaw                  string   `json:"psbtRaw" bson:"psbtRaw"`
	BtcPoolMode              PoolMode `json:"btcPoolMode" bson:"btcPoolMode"` //PoolMode for btc
	UtxoId                   string   `json:"utxoId" bson:"utxoId"`           //UtxoId
	RefundTx                 string   `json:"refundTx" bson:"refundTx"`       //UtxoId
	PoolType                 PoolType `json:"poolType" bson:"poolType"`

	PoolState      PoolState `json:"poolState" bson:"poolState"`
	DealTx         string    `json:"dealTx" bson:"dealTx"`
	DealTxIndex    int64     `json:"dealTxIndex" bson:"dealTxIndex"`
	DealTxOutValue int64     `json:"dealTxOutValue" bson:"dealTxOutValue"`
	DealTime       int64     `json:"dealTime" bson:"dealTime"`

	PoolCoinState        PoolState         `json:"poolCoinState" bson:"poolCoinState"`
	DealCoinTxBlockState ClaimTxBlockState `json:"dealCoinTxBlockState" bson:"dealCoinTxBlockState"`
	DealCoinTxBlock      int64             `json:"dealCoinTxBlock" bson:"dealCoinTxBlock"`
	DealCoinTx           string            `json:"dealCoinTx" bson:"dealCoinTx"`
	DealCoinTxIndex      int64             `json:"dealCoinTxIndex" bson:"dealCoinTxIndex"`
	DealCoinTxOutValue   int64             `json:"dealCoinTxOutValue" bson:"dealCoinTxOutValue"`
	DealCoinTime         int64             `json:"dealCoinTime" bson:"dealCoinTime"`

	DealInscriptionId                       string                                  `json:"dealInscriptionId" bson:"dealInscriptionId"` //InscriptionId
	DealInscriptionTx                       string                                  `json:"dealInscriptionTx" bson:"dealInscriptionTx"`
	DealInscriptionTxIndex                  int64                                   `json:"dealInscriptionTxIndex" bson:"dealInscriptionTxIndex"`
	DealInscriptionTxOutValue               int64                                   `json:"dealInscriptionTxOutValue" bson:"dealInscriptionTxOutValue"`
	DealInscriptionTime                     int64                                   `json:"dealInscriptionTime" bson:"dealInscriptionTime"`
	MultiSigScriptAddressTickAvailableState MultiSigScriptAddressTickAvailableState `json:"multiSigScriptAddressTickAvailableState" bson:"multiSigScriptAddressTickAvailableState"`

	ClaimTx           string            `json:"claimTx" bson:"claimTx"`
	ClaimTime         int64             `json:"claimTime" bson:"claimTime"`
	ClaimTxBlock      int64             `json:"claimTxBlock" bson:"claimTxBlock"`
	ClaimTxBlockState ClaimTxBlockState `json:"claimTxBlockState" bson:"claimTxBlockState"`
	Percentage        int64             `json:"percentage" bson:"percentage"`
	CalValue          int64             `json:"calValue" bson:"calValue"`
	CalTotalValue     int64             `json:"calTotalValue" bson:"calTotalValue"`
	CalStartBlock     int64             `json:"calStartBlock" bson:"calStartBlock"`
	CalEndBlock       int64             `json:"calEndBlock" bson:"calEndBlock"`
	RewardAmount      int64             `json:"rewardAmount" bson:"rewardAmount"`
	RewardRealAmount  int64             `json:"rewardRealAmount" bson:"rewardRealAmount"`
	PercentageExtra   int64             `json:"percentageExtra" bson:"percentageExtra"`
	RewardExtraAmount int64             `json:"rewardExtraAmount" bson:"rewardExtraAmount"`
	Decreasing        int64             `json:"decreasing" bson:"decreasing"`
	Ratio             int64             `json:"ratio" bson:"ratio"`
	RewardRatio       int64             `json:"rewardRatio" bson:"rewardRatio"`
	Timestamp         int64             `json:"timestamp" bson:"timestamp"`
	Version           int64             `json:"version" bson:"version"`
	CreateTime        int64             `json:"createTime" bson:"createTime"`
	UpdateTime        int64             `json:"updateTime" bson:"updateTime"`
	State             int64             `json:"state" bson:"state"`
}

type PoolOrderCount struct {
	Id              string `json:"id" bson:"_id"`
	CoinAmountTotal int64  `json:"coinAmountTotal" bson:"coinAmountTotal"`
	AmountTotal     int64  `json:"amountTotal" bson:"amountTotal"`
	OrderCounts     int64  `json:"orderCounts" bson:"orderCounts"`
}

type PoolRewardCount struct {
	Id                     string `json:"id" bson:"_id"`
	CoinAmountTotal        int64  `json:"coinAmountTotal" bson:"coinAmountTotal"`
	AmountTotal            int64  `json:"amountTotal" bson:"amountTotal"`
	RewardAmountTotal      int64  `json:"rewardAmountTotal" bson:"rewardAmountTotal"`
	RewardExtraAmountTotal int64  `json:"rewardExtraAmountTotal" bson:"rewardExtraAmountTotal"`
	OrderCounts            int64  `json:"orderCounts" bson:"orderCounts"`
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

type RewardState int

const (
	RewardStateNo          RewardState = 0
	RewardStateCreate      RewardState = 1
	RewardStateInscription RewardState = 2
	RewardStateSend        RewardState = 3
	RewardStateAll         RewardState = 100
)

type RewardType int

const (
	RewardTypeNormal             RewardType = 1
	RewardTypeExtra              RewardType = 2
	RewardTypeEventOneLp         RewardType = 11
	RewardTypeEventOneBid        RewardType = 12
	RewardTypeEventOneLpUnused   RewardType = 13
	RewardTypeEventOneLpUnusedV2 RewardType = 15
)

type PoolRewardOrderModel struct {
	Id                  int64       `json:"id" bson:"_id" tb:"pool_reward_order_model" mg:"true"`
	Net                 string      `json:"net" bson:"net"`
	Tick                string      `json:"tick" bson:"tick"`
	OrderId             string      `json:"orderId" bson:"orderId"`
	Pair                string      `json:"pair" bson:"pair"`
	RewardCoinAmount    int64       `json:"rewardCoinAmount" bson:"rewardCoinAmount"`
	Address             string      `json:"address" bson:"address"`
	RewardType          RewardType  `json:"rewardType" bson:"rewardType"`
	RewardState         RewardState `json:"rewardState" bson:"rewardState"`
	InscriptionId       string      `json:"inscriptionId" bson:"inscriptionId"`
	InscriptionOutValue int64       `json:"inscriptionOutValue" bson:"inscriptionOutValue"`
	SendId              string      `json:"sendId" bson:"sendId"`
	FeeRawTx            string      `json:"feeRawTx" bson:"feeRawTx"`
	FeeUtxoTxId         string      `json:"feeUtxoTxId" bson:"feeUtxoTxId"`
	FeeInscription      int64       `json:"feeInscription" bson:"feeInscription"`
	FeeSend             int64       `json:"feeSend" bson:"feeSend"`
	NetworkFeeRate      int64       `json:"networkFeeRate" bson:"networkFeeRate"`
	Version             int         `json:"version" bson:"version"`
	Timestamp           int64       `json:"timestamp" bson:"timestamp"`
	CreateTime          int64       `json:"createTime" bson:"createTime"`
	UpdateTime          int64       `json:"updateTime" bson:"updateTime"`
	State               int64       `json:"state" bson:"state"`
}

type PoolRewardOrderCount struct {
	Id                    string `json:"id" bson:"_id"`
	CoinAmountTotal       int64  `json:"coinAmountTotal" bson:"coinAmountTotal"`
	AmountTotal           int64  `json:"amountTotal" bson:"amountTotal"`
	RewardCoinAmountTotal int64  `json:"rewardCoinAmountTotal" bson:"rewardCoinAmountTotal"`
	RewardCoinOrderCount  int64  `json:"rewardCoinOrderCount" bson:"rewardCoinOrderCount"`
}

func (s PoolRewardOrderModel) getCollection() string {
	return "pool_reward_order_model"
}

func (s PoolRewardOrderModel) getDB() string {
	return major.DsOrdbook
}

func (s PoolRewardOrderModel) GetReadDB() (*mongo.Collection, error) {
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

func (s PoolRewardOrderModel) GetWriteDB() (*mongo.Collection, error) {
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

type InfoType int

const (
	InfoTypeBlock  InfoType = 1
	InfoTypeNoUsed InfoType = 2
)

type PoolBlockUserInfoModel struct {
	Id             int64    `json:"id" bson:"_id" tb:"pool_block_user_info_model" mg:"true"`
	BlockUserId    string   `json:"blockUserId" bson:"blockUserId"`
	Net            string   `json:"net" bson:"net"`
	InfoType       InfoType `json:"infoType" bson:"infoType"`   //
	HasNoUsed      bool     `json:"hasNoUsed" bson:"hasNoUsed"` //
	Address        string   `json:"address" bson:"address"`
	BigBlock       int64    `json:"bigBlock" bson:"bigBlock"`
	StartBlock     int64    `json:"startBlock" bson:"startBlock"`
	CycleBlock     int64    `json:"cycleBlock" bson:"cycleBlock"`
	CoinPrice      int64    `json:"coinPrice" bson:"coinPrice"`
	CoinAmount     int64    `json:"coinAmount" bson:"coinAmount"`
	Amount         int64    `json:"amount" bson:"amount"`
	UserTotalValue int64    `json:"userTotalValue" bson:"userTotalValue"`
	AllTotalValue  int64    `json:"allTotalValue" bson:"allTotalValue"`
	Percentage     int64    `json:"percentage" bson:"percentage"`     //10000
	RewardAmount   int64    `json:"rewardAmount" bson:"rewardAmount"` //
	Timestamp      int64    `json:"timestamp" bson:"timestamp"`
	CreateTime     int64    `json:"createTime" bson:"createTime"`
	UpdateTime     int64    `json:"updateTime" bson:"updateTime"`
	State          int64    `json:"state" bson:"state"`
}

type PoolRewardBlockUserCount struct {
	Id                    string `json:"id" bson:"_id"`
	RewardCoinAmountTotal int64  `json:"rewardCoinAmountTotal" bson:"rewardCoinAmountTotal"`
}

func (s PoolBlockUserInfoModel) getCollection() string {
	return "pool_block_user_info_model"
}

func (s PoolBlockUserInfoModel) getDB() string {
	return major.DsOrdbook
}

func (s PoolBlockUserInfoModel) GetReadDB() (*mongo.Collection, error) {
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

func (s PoolBlockUserInfoModel) GetWriteDB() (*mongo.Collection, error) {
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

type CalType int

const (
	CalTypePlatform CalType = 1
	CalTypeEventOne CalType = 11
)

type PoolBlockInfoModel struct {
	Id                                   int64             `json:"id" bson:"_id" tb:"pool_block_info_model" mg:"true"`
	BigBlockId                           string            `json:"bigBlockId" bson:"bigBlockId"` //bigBlock_cycleBlock_calType
	BigBlock                             int64             `json:"bigBlock" bson:"bigBlock"`
	StartBlock                           int64             `json:"startBlock" bson:"startBlock"`
	EndBlock                             int64             `json:"endBlock" bson:"endBlock"`
	CycleBlock                           int64             `json:"cycleBlock" bson:"cycleBlock"`
	CalPoolRewardInfo                    map[string]string `json:"calPoolRewardInfo" bson:"calPoolRewardInfo"` //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
	CalPoolRewardTotalValue              int64             `json:"calPoolRewardTotalValue" bson:"calPoolRewardTotalValue"`
	CalPoolExtraRewardInfo               map[string]string `json:"calPoolExtraRewardInfo" bson:"calPoolExtraRewardInfo"` //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
	CalPoolExtraRewardTotalValue         int64             `json:"calPoolExtraRewardTotalValue" bson:"calPoolExtraRewardTotalValue"`
	CalEventBidDealExtraRewardInfo       map[string]string `json:"calEventBidDealExtraRewardInfo" bson:"calEventBidDealExtraRewardInfo"` //{"brc20OrderId":"value:percentage:dealAmount"}
	CalEventBidDealExtraRewardTotalValue int64             `json:"calEventBidDealExtraRewardTotalValue" bson:"calEventBidDealExtraRewardTotalValue"`
	CalType                              CalType           `json:"calType" bson:"calType"`
	Timestamp                            int64             `json:"timestamp" bson:"timestamp"`
	CreateTime                           int64             `json:"createTime" bson:"createTime"`
	UpdateTime                           int64             `json:"updateTime" bson:"updateTime"`
	State                                int64             `json:"state" bson:"state"`
}

func (s PoolBlockInfoModel) getCollection() string {
	return "pool_block_info_model"
}

func (s PoolBlockInfoModel) getDB() string {
	return major.DsOrdbook
}

func (s PoolBlockInfoModel) GetReadDB() (*mongo.Collection, error) {
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

func (s PoolBlockInfoModel) GetWriteDB() (*mongo.Collection, error) {
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

type PoolTimeInfoModel struct {
	Id                           int64             `json:"id" bson:"_id" tb:"pool_time_info_model" mg:"true"`
	BigTimeId                    string            `json:"bigTimeId" bson:"bigTimeId"` //bigTime_cycleBlock_calType
	BigTime                      int64             `json:"bigTime" bson:"bigTime"`
	StartTime                    int64             `json:"startTime" bson:"startTime"`
	EndTime                      int64             `json:"endTime" bson:"endTime"`
	StartTimeStr                 string            `json:"startTimeStr" bson:"startTime"`
	EndTimeStr                   string            `json:"endTimeStr" bson:"endTimeStr"`
	CycleTime                    int64             `json:"cycleTime" bson:"cycleTime"`
	CalPoolExtraRewardInfo       map[string]string `json:"calPoolExtraRewardInfo" bson:"calPoolExtraRewardInfo"` //{"poolOrderId":"value:percentage:amount:coinAmount:price"}
	CalPoolExtraRewardTotalValue int64             `json:"calPoolExtraRewardTotalValue" bson:"calPoolExtraRewardTotalValue"`
	CalType                      CalType           `json:"calType" bson:"calType"`
	Timestamp                    int64             `json:"timestamp" bson:"timestamp"`
	CreateTime                   int64             `json:"createTime" bson:"createTime"`
	UpdateTime                   int64             `json:"updateTime" bson:"updateTime"`
	State                        int64             `json:"state" bson:"state"`
}

func (s PoolTimeInfoModel) getCollection() string {
	return "pool_time_info_model"
}

func (s PoolTimeInfoModel) getDB() string {
	return major.DsOrdbook
}

func (s PoolTimeInfoModel) GetReadDB() (*mongo.Collection, error) {
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

func (s PoolTimeInfoModel) GetWriteDB() (*mongo.Collection, error) {
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

type RewardRecordModel struct {
	Id                  int64      `json:"id" bson:"_id" tb:"reward_record_model" mg:"true"`
	Net                 string     `json:"net" bson:"net"`
	Tick                string     `json:"tick" bson:"tick"` //net_tick_calDay_fromOrderId
	OrderId             string     `json:"orderId" bson:"orderId"`
	Pair                string     `json:"pair" bson:"pair"`
	FromOrderId         string     `json:"fromOrderId" bson:"fromOrderId"`
	FromOrderRole       string     `json:"fromOrderRole" bson:"fromOrderRole"`
	FromOrderTotalValue int64      `json:"fromOrderTotalValue" bson:"fromOrderTotalValue"`
	FromOrderOwnValue   int64      `json:"fromOrderOwnValue" bson:"fromOrderOwnValue"`
	Address             string     `json:"address" bson:"address"`
	TotalValue          int64      `json:"totalValue" bson:"totalValue"`
	OwnValue            int64      `json:"ownValue" bson:"ownValue"`
	Percentage          int64      `json:"percentage" bson:"percentage"`
	RewardAmount        int64      `json:"rewardAmount" bson:"rewardAmount"`
	RewardType          RewardType `json:"rewardType" bson:"rewardType"`
	CalBigBlock         int64      `json:"calBigBlock" bson:"calBigBlock"`
	CalDayIndex         int64      `json:"calDayIndex" bson:"calDayIndex"`
	CalDay              int64      `json:"calDay" bson:"calDay"`
	CalStartTime        int64      `json:"calStartTime" bson:"calStartTime"`
	CalEndTime          int64      `json:"calEndTime" bson:"calEndTime"`
	CalStartBlock       int64      `json:"calStartBlock" bson:"calStartBlock"`
	CalEndBlock         int64      `json:"calEndBlock" bson:"calEndBlock"`
	Version             int        `json:"version" bson:"version"`
	Timestamp           int64      `json:"timestamp" bson:"timestamp"`
	CreateTime          int64      `json:"createTime" bson:"createTime"`
	UpdateTime          int64      `json:"updateTime" bson:"updateTime"`
	State               int64      `json:"state" bson:"state"`
}

type RewardCount struct {
	Id                string `json:"id" bson:"_id"`
	RewardAmountTotal int64  `json:"rewardAmountTotal" bson:"rewardAmountTotal"`
	OrderCounts       int64  `json:"orderCounts" bson:"orderCounts"`
}

func (s RewardRecordModel) getCollection() string {
	return "reward_record_model"
}

func (s RewardRecordModel) getDB() string {
	return major.DsOrdbook
}

func (s RewardRecordModel) GetReadDB() (*mongo.Collection, error) {
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

func (s RewardRecordModel) GetWriteDB() (*mongo.Collection, error) {
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
