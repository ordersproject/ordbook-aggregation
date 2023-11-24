package model

import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"ordbook-aggregation/major"
)

type OrderBrc20Model struct {
	Id                  int64            `json:"id" bson:"_id" tb:"order_brc20_model" mg:"true"`
	Net                 string           `json:"net" bson:"net"`
	OrderId             string           `json:"orderId" bson:"orderId"`
	Tick                string           `json:"tick" bson:"tick"`
	Amount              uint64           `json:"amount" bson:"amount"`
	DecimalNum          int              `json:"decimalNum" bson:"decimalNum"`
	CoinAmount          uint64           `json:"coinAmount" bson:"coinAmount"`
	CoinDecimalNum      int              `json:"coinDecimalNum" bson:"coinDecimalNum"`
	CoinRatePrice       uint64           `json:"coinRatePrice" bson:"coinRatePrice"`
	CoinPrice           int64            `json:"coinPrice" bson:"coinPrice"`                     //MAX-9223372036854775807
	CoinPriceDecimalNum int32            `json:"coinPriceDecimalNum" bson:"coinPriceDecimalNum"` //8
	OrderState          OrderState       `json:"orderState" bson:"orderState"`                   //1-create,2-finish,3-cancel
	OrderType           OrderType        `json:"orderType" bson:"orderType"`                     //1-sell,2-buy
	SellerAddress       string           `json:"sellerAddress" bson:"sellerAddress"`
	BuyerAddress        string           `json:"buyerAddress" bson:"buyerAddress"`
	BuyerIp             string           `json:"buyerIp" bson:"buyerIp"`
	MarketAmount        uint64           `json:"marketAmount" bson:"marketAmount"`
	PlatformFee         uint64           `json:"platformFee" bson:"platformFee"`
	ChangeAmount        uint64           `json:"changeAmount" bson:"changeAmount"`
	Fee                 uint64           `json:"fee" bson:"fee"`
	FeeRate             int              `json:"feeRate" bson:"feeRate"`
	SupplementaryAmount uint64           `json:"supplementaryAmount" bson:"supplementaryAmount"`
	PlatformTx          string           `json:"platformTx" bson:"platformTx"`
	InscriptionId       string           `json:"inscriptionId" bson:"inscriptionId"`
	InscriptionNumber   string           `json:"inscriptionNumber" bson:"inscriptionNumber"`
	SellInscriptionId   string           `json:"sellInscriptionId" bson:"sellInscriptionId"`
	PsbtRawPreAsk       string           `json:"psbtRawPreAsk" bson:"psbtRawPreAsk"`
	PsbtRawFinalAsk     string           `json:"psbtRawFinalAsk" bson:"psbtRawFinalAsk"`
	PsbtAskTxId         string           `json:"psbtAskTxId" bson:"psbtAskTxId"`
	BidValueToXUtxoId   string           `json:"bidValueToXUtxoId" bson:"bidValueToXUtxoId"` //PsbtAskTxId_bidYOffsetIndex
	PsbtRawPreBid       string           `json:"psbtRawPreBid" bson:"psbtRawPreBid"`
	PsbtRawMidBid       string           `json:"psbtRawMidBid" bson:"psbtRawMidBid"`
	PsbtRawFinalBid     string           `json:"psbtRawFinalBid" bson:"psbtRawFinalBid"`
	PsbtBidTxId         string           `json:"psbtBidTxId" bson:"psbtBidTxId"`
	PoolOrderId         string           `json:"poolOrderId" bson:"poolOrderId"`
	Integral            int64            `json:"integral" bson:"integral"`
	FreeState           FreeState        `json:"freeState" bson:"freeState"`
	PlatformDummy       PlatformDummy    `json:"platformDummy" bson:"platformDummy"` //0-no 1-yes
	DealTime            int64            `json:"dealTime" bson:"dealTime"`
	Timestamp           int64            `json:"timestamp" bson:"timestamp"`
	InscriptionState    InscriptionState `json:"inscriptionState" bson:"inscriptionState"`

	DealTxBlockState ClaimTxBlockState `json:"dealTxBlockState" bson:"dealTxBlockState"` //psbtBidTxId
	DealTxBlock      int64             `json:"dealTxBlock" bson:"dealTxBlock"`
	Percentage       int64             `json:"percentage" bson:"percentage"`
	CalValue         int64             `json:"calValue" bson:"calValue"`
	CalTotalValue    int64             `json:"calTotalValue" bson:"calTotalValue"`
	CalStartBlock    int64             `json:"calStartBlock" bson:"calStartBlock"`
	CalEndBlock      int64             `json:"calEndBlock" bson:"calEndBlock"`
	RewardAmount     int64             `json:"rewardAmount" bson:"rewardAmount"`
	RewardRealAmount int64             `json:"rewardRealAmount" bson:"rewardRealAmount"`
	Version          int               `json:"version" bson:"version"`

	CreateTime int64 `json:"createTime" bson:"createTime"`
	UpdateTime int64 `json:"updateTime" bson:"updateTime"`
	State      int64 `json:"state" bson:"state"`
}

type EventRewardCount struct {
	Id                string `json:"id" bson:"_id"`
	AmountTotal       int64  `json:"amountTotal" bson:"amountTotal"`
	RewardAmountTotal int64  `json:"rewardAmountTotal" bson:"rewardAmountTotal"`
	OrderCounts       int64  `json:"orderCounts" bson:"orderCounts"`
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
	DummyStateLive   DummyState = 1
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

// marketPrice  model
type OrderBrc20MarketPriceModel struct {
	Id         int64  `json:"id" bson:"_id" tb:"order_brc20_market_price_model" mg:"true"`
	Net        string `json:"net" bson:"net"`
	Pair       string `json:"pair" bson:"pair"`
	Tick       string `json:"tick" bson:"tick"`
	Price      int64  `json:"price" bson:"price"` //1 brc20 = xxx sats
	Timestamp  int64  `json:"timestamp" bson:"timestamp"`
	CreateTime int64  `json:"createTime" bson:"createTime"`
	UpdateTime int64  `json:"updateTime" bson:"updateTime"`
	State      int64  `json:"state" bson:"state"`
}

func (s OrderBrc20MarketPriceModel) getCollection() string {
	return "order_brc20_market_price_model"
}

func (s OrderBrc20MarketPriceModel) getDB() string {
	return major.DsOrdbook
}

func (s OrderBrc20MarketPriceModel) GetReadDB() (*mongo.Collection, error) {
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

func (s OrderBrc20MarketPriceModel) GetWriteDB() (*mongo.Collection, error) {
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

// market info  model
type OrderBrc20MarketInfoModel struct {
	Id         int64  `json:"id" bson:"_id" tb:"order_brc20_market_info_model" mg:"true"`
	NetDate    string `json:"netDate" bson:"netDate"`
	Net        string `json:"net" bson:"net"`
	Date       string `json:"date" bson:"date"`
	AskVolume  int64  `json:"askVolume" bson:"askVolume"`
	BidVolume  int64  `json:"bidVolume" bson:"bidVolume"`
	AskFees    int64  `json:"askFees" bson:"askFees"`
	BidFees    int64  `json:"bidFees" bson:"bidFees"`
	Between    string `json:"between" bson:"between"`
	Timestamp  int64  `json:"timestamp" bson:"timestamp"`
	CreateTime int64  `json:"createTime" bson:"createTime"`
	UpdateTime int64  `json:"updateTime" bson:"updateTime"`
	State      int64  `json:"state" bson:"state"`
}

func (s OrderBrc20MarketInfoModel) getCollection() string {
	return "order_brc20_market_info_model"
}

func (s OrderBrc20MarketInfoModel) getDB() string {
	return major.DsOrdbook
}

func (s OrderBrc20MarketInfoModel) GetReadDB() (*mongo.Collection, error) {
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

func (s OrderBrc20MarketInfoModel) GetWriteDB() (*mongo.Collection, error) {
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

type NotificationType int64

const (
	NotificationTypePoolUsed    NotificationType = 1
	NotificationTypeBidInvalid  NotificationType = 2
	NotificationTypeOrderFinish NotificationType = 3
)

// notification model
type OrderNotificationModel struct {
	Id                int64            `json:"id" bson:"_id" tb:"order_notification_model" mg:"true"`
	Address           string           `json:"address" bson:"address"`
	NotificationType  NotificationType `json:"notificationType" bson:"notificationType"`
	NotificationCount int64            `json:"notificationCount" bson:"notificationCount"`
	Timestamp         int64            `json:"timestamp" bson:"timestamp"`
	CreateTime        int64            `json:"createTime" bson:"createTime"`
	UpdateTime        int64            `json:"updateTime" bson:"updateTime"`
	State             int64            `json:"state" bson:"state"`
}

func (s OrderNotificationModel) getCollection() string {
	return "order_notification_model"
}

func (s OrderNotificationModel) getDB() string {
	return major.DsOrdbook
}

func (s OrderNotificationModel) GetReadDB() (*mongo.Collection, error) {
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

func (s OrderNotificationModel) GetWriteDB() (*mongo.Collection, error) {
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
