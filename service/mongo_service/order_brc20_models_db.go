package mongo_service

import (
	"context"
	"errors"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/model"
	"strings"
)

func FindOrderBrc20ModelByOrderId(orderId string) (*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"orderId", orderId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderBrc20Model{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createOrderBrc20Model(orderBrc20 *model.OrderBrc20Model) (*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "orderId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "coinRatePrice")
	CreateIndex(collection, "coinPrice")
	CreateIndex(collection, "metaBlockHeight")
	CreateIndex(collection, "orderState")
	CreateIndex(collection, "orderType")
	CreateIndex(collection, "sellerAddress")
	CreateIndex(collection, "buyerAddress")
	CreateIndex(collection, "buyerIp")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "dealTime")
	CreateIndex(collection, "poolOrderId")
	CreateIndex(collection, "poolOrderMode")
	CreateIndex(collection, "inscriptionId")
	CreateIndex(collection, "sellInscriptionId")
	CreateIndex(collection, "version")
	CreateIndex(collection, "dealTxBlock")
	CreateIndex(collection, "dealTxBlockState")

	entity := &model.OrderBrc20Model{
		Id:                  util.GetUUIDInt64(),
		Net:                 orderBrc20.Net,
		OrderId:             orderBrc20.OrderId,
		Tick:                orderBrc20.Tick,
		Amount:              orderBrc20.Amount,
		DecimalNum:          orderBrc20.DecimalNum,
		CoinAmount:          orderBrc20.CoinAmount,
		CoinDecimalNum:      orderBrc20.CoinDecimalNum,
		CoinRatePrice:       orderBrc20.CoinRatePrice,
		CoinPrice:           orderBrc20.CoinPrice,
		CoinPriceDecimalNum: orderBrc20.CoinPriceDecimalNum,
		OrderState:          orderBrc20.OrderState,
		OrderType:           orderBrc20.OrderType,
		SellerAddress:       orderBrc20.SellerAddress,
		BuyerAddress:        orderBrc20.BuyerAddress,
		BuyerIp:             orderBrc20.BuyerIp,
		MarketAmount:        orderBrc20.MarketAmount,
		PlatformFee:         orderBrc20.PlatformFee,
		PlatformSellFee:     orderBrc20.PlatformSellFee,
		ChangeAmount:        orderBrc20.ChangeAmount,
		Fee:                 orderBrc20.Fee,
		FeeRate:             orderBrc20.FeeRate,
		SupplementaryAmount: orderBrc20.SupplementaryAmount,
		PlatformTx:          orderBrc20.PlatformTx,
		InscriptionId:       orderBrc20.InscriptionId,
		InscriptionNumber:   orderBrc20.InscriptionNumber,
		SellInscriptionId:   orderBrc20.SellInscriptionId,
		PsbtRawPreAsk:       orderBrc20.PsbtRawPreAsk,
		PsbtRawFinalAsk:     orderBrc20.PsbtRawFinalAsk,
		PsbtAskTxId:         orderBrc20.PsbtAskTxId,
		PsbtRawPreBid:       orderBrc20.PsbtRawPreBid,
		PsbtRawMidBid:       orderBrc20.PsbtRawMidBid,
		PsbtRawFinalBid:     orderBrc20.PsbtRawFinalBid,
		PsbtBidTxId:         orderBrc20.PsbtBidTxId,
		PoolOrderId:         orderBrc20.PoolOrderId,
		PoolCoinAddress:     orderBrc20.PoolCoinAddress,
		PoolOrderMode:       orderBrc20.PoolOrderMode,
		PoolPreUtxoRaw:      orderBrc20.PoolPreUtxoRaw,
		PoolUtxoId:          orderBrc20.PoolUtxoId,
		Integral:            orderBrc20.Integral,
		FreeState:           orderBrc20.FreeState,
		SellerTotalFee:      orderBrc20.SellerTotalFee,
		BuyerTotalFee:       orderBrc20.BuyerTotalFee,
		DealTime:            orderBrc20.DealTime,
		DealTxBlockState:    orderBrc20.DealTxBlockState,
		DealTxBlock:         orderBrc20.DealTxBlock,
		Percentage:          orderBrc20.Percentage,
		CalValue:            orderBrc20.CalValue,
		CalTotalValue:       orderBrc20.CalTotalValue,
		CalStartBlock:       orderBrc20.CalStartBlock,
		CalEndBlock:         orderBrc20.CalEndBlock,
		RewardAmount:        orderBrc20.RewardAmount,
		RewardRealAmount:    orderBrc20.RewardRealAmount,
		Version:             orderBrc20.Version,
		Timestamp:           orderBrc20.Timestamp,
		PlatformDummy:       orderBrc20.PlatformDummy,
		CreateTime:          util.Time(),
		State:               model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetOrderBrc20Model(orderBrc20 *model.OrderBrc20Model) (*model.OrderBrc20Model, error) {
	entity, err := FindOrderBrc20ModelByOrderId(orderBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.OrderBrc20Model{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"orderId", orderBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: orderBrc20.Net})
		bsonData = append(bsonData, bson.E{Key: "orderId", Value: orderBrc20.OrderId})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: orderBrc20.Tick})
		bsonData = append(bsonData, bson.E{Key: "amount", Value: orderBrc20.Amount})
		bsonData = append(bsonData, bson.E{Key: "decimalNum", Value: orderBrc20.DecimalNum})
		bsonData = append(bsonData, bson.E{Key: "coinAmount", Value: orderBrc20.CoinAmount})
		bsonData = append(bsonData, bson.E{Key: "coinDecimalNum", Value: orderBrc20.CoinDecimalNum})
		bsonData = append(bsonData, bson.E{Key: "coinRatePrice", Value: orderBrc20.CoinRatePrice})
		bsonData = append(bsonData, bson.E{Key: "coinPrice", Value: orderBrc20.CoinPrice})
		bsonData = append(bsonData, bson.E{Key: "coinPriceDecimalNum", Value: orderBrc20.CoinPriceDecimalNum})
		bsonData = append(bsonData, bson.E{Key: "orderState", Value: orderBrc20.OrderState})
		bsonData = append(bsonData, bson.E{Key: "orderType", Value: orderBrc20.OrderType})
		bsonData = append(bsonData, bson.E{Key: "sellerAddress", Value: orderBrc20.SellerAddress})
		bsonData = append(bsonData, bson.E{Key: "buyerAddress", Value: orderBrc20.BuyerAddress})
		bsonData = append(bsonData, bson.E{Key: "buyerIp", Value: orderBrc20.BuyerIp})
		bsonData = append(bsonData, bson.E{Key: "marketAmount", Value: orderBrc20.MarketAmount})
		bsonData = append(bsonData, bson.E{Key: "platformFee", Value: orderBrc20.PlatformFee})
		bsonData = append(bsonData, bson.E{Key: "platformSellFee", Value: orderBrc20.PlatformSellFee})
		bsonData = append(bsonData, bson.E{Key: "changeAmount", Value: orderBrc20.ChangeAmount})
		bsonData = append(bsonData, bson.E{Key: "fee", Value: orderBrc20.Fee})
		bsonData = append(bsonData, bson.E{Key: "feeRate", Value: orderBrc20.FeeRate})
		bsonData = append(bsonData, bson.E{Key: "supplementaryAmount", Value: orderBrc20.SupplementaryAmount})
		bsonData = append(bsonData, bson.E{Key: "platformTx", Value: orderBrc20.PlatformTx})
		bsonData = append(bsonData, bson.E{Key: "inscriptionId", Value: orderBrc20.InscriptionId})
		bsonData = append(bsonData, bson.E{Key: "inscriptionNumber", Value: orderBrc20.InscriptionNumber})
		bsonData = append(bsonData, bson.E{Key: "sellInscriptionId", Value: orderBrc20.SellInscriptionId})
		bsonData = append(bsonData, bson.E{Key: "bidValueToXUtxoId", Value: orderBrc20.BidValueToXUtxoId})
		bsonData = append(bsonData, bson.E{Key: "psbtRawPreAsk", Value: orderBrc20.PsbtRawPreAsk})
		bsonData = append(bsonData, bson.E{Key: "psbtRawFinalAsk", Value: orderBrc20.PsbtRawFinalAsk})
		bsonData = append(bsonData, bson.E{Key: "psbtAskTxId", Value: orderBrc20.PsbtAskTxId})
		bsonData = append(bsonData, bson.E{Key: "psbtRawPreBid", Value: orderBrc20.PsbtRawPreBid})
		bsonData = append(bsonData, bson.E{Key: "psbtRawMidBid", Value: orderBrc20.PsbtRawMidBid})
		bsonData = append(bsonData, bson.E{Key: "psbtRawFinalBid", Value: orderBrc20.PsbtRawFinalBid})
		bsonData = append(bsonData, bson.E{Key: "psbtBidTxId", Value: orderBrc20.PsbtBidTxId})
		bsonData = append(bsonData, bson.E{Key: "poolOrderId", Value: orderBrc20.PoolOrderId})
		bsonData = append(bsonData, bson.E{Key: "poolCoinAddress", Value: orderBrc20.PoolCoinAddress})
		bsonData = append(bsonData, bson.E{Key: "poolOrderMode", Value: orderBrc20.PoolOrderMode})
		bsonData = append(bsonData, bson.E{Key: "poolPreUtxoRaw", Value: orderBrc20.PoolPreUtxoRaw})
		bsonData = append(bsonData, bson.E{Key: "poolUtxoId", Value: orderBrc20.PoolUtxoId})
		bsonData = append(bsonData, bson.E{Key: "integral", Value: orderBrc20.Integral})
		bsonData = append(bsonData, bson.E{Key: "freeState", Value: orderBrc20.FreeState})
		bsonData = append(bsonData, bson.E{Key: "dealTime", Value: orderBrc20.DealTime})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: orderBrc20.Timestamp})

		bsonData = append(bsonData, bson.E{Key: "sellerTotalFee", Value: orderBrc20.SellerTotalFee})
		bsonData = append(bsonData, bson.E{Key: "buyerTotalFee", Value: orderBrc20.BuyerTotalFee})

		bsonData = append(bsonData, bson.E{Key: "dealTxBlockState", Value: orderBrc20.DealTxBlockState})
		bsonData = append(bsonData, bson.E{Key: "dealTxBlock", Value: orderBrc20.DealTxBlock})
		bsonData = append(bsonData, bson.E{Key: "percentage", Value: orderBrc20.Percentage})
		bsonData = append(bsonData, bson.E{Key: "calValue", Value: orderBrc20.CalValue})
		bsonData = append(bsonData, bson.E{Key: "calTotalValue", Value: orderBrc20.CalTotalValue})
		bsonData = append(bsonData, bson.E{Key: "calStartBlock", Value: orderBrc20.CalStartBlock})
		bsonData = append(bsonData, bson.E{Key: "calEndBlock", Value: orderBrc20.CalEndBlock})
		bsonData = append(bsonData, bson.E{Key: "rewardAmount", Value: orderBrc20.RewardAmount})
		bsonData = append(bsonData, bson.E{Key: "rewardRealAmount", Value: orderBrc20.RewardRealAmount})
		bsonData = append(bsonData, bson.E{Key: "version", Value: orderBrc20.Version})

		bsonData = append(bsonData, bson.E{Key: "platformDummy", Value: orderBrc20.PlatformDummy})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return orderBrc20, nil
	} else {
		return createOrderBrc20Model(orderBrc20)
	}
}

func FindOrderBrc20ModelByInscriptionId(inscriptionId string, orderState model.OrderState) (*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"inscriptionId", inscriptionId},
		{"orderState", orderState},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderBrc20Model{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func CountOrderBrc20ModelListForClaim(net, tick string, orderState model.OrderState, coinAmount int64) (int64, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"freeState": model.FreeStateClaim,
		"state":     model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if orderState != 0 {
		find["orderState"] = orderState
	}
	if coinAmount != 0 {
		find["coinAmount"] = coinAmount
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func CountOrderBrc20ModelList(net, tick, sellerAddress, buyerAddress string, orderType model.OrderType, orderState model.OrderState, poolOrderModel model.PoolMode) (int64, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if sellerAddress != "" {
		find["sellerAddress"] = sellerAddress
	}
	if buyerAddress != "" {
		find["buyerAddress"] = buyerAddress
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if orderState != 0 {
		find["orderState"] = orderState
	}
	if poolOrderModel != model.PoolModeDefault {
		find["poolOrderMode"] = poolOrderModel
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindOrderBrc20ModelList(net, tick, sellerAddress, buyerAddress string,
	orderType model.OrderType, orderState model.OrderState,
	limit int64, flag, page int64, sortKey string, sortType int64, freeState model.FreeState, coinAmount int64,
	poolOrderModel model.PoolMode) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if sellerAddress != "" {
		find["sellerAddress"] = sellerAddress
	}
	if buyerAddress != "" {
		find["buyerAddress"] = buyerAddress
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if coinAmount != 0 {
		find["coinAmount"] = coinAmount
	}
	if orderState != 0 {
		if orderState == model.OrderStateAll {
			find["orderState"] = bson.M{IN_: []model.OrderState{
				model.OrderStateCreate,
				model.OrderStateFinish,
				model.OrderStateCancel,
				model.OrderStateErr,
			}}
		} else {
			find["orderState"] = orderState
		}
	}

	if freeState != 0 {
		find["freeState"] = freeState
	}
	if poolOrderModel != model.PoolModeDefault {
		find["poolOrderMode"] = poolOrderModel
	}

	switch sortKey {
	case "coinRatePrice":
		sortKey = "coinRatePrice"
	case "coinPrice":
		sortKey = "coinPrice"
	default:
		sortKey = "timestamp"
	}

	flagKey := GT_
	if sortType >= 0 {
		sortType = 1
		flagKey = GT_
	} else {
		sortType = -1
		flagKey = LT_
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	} else if flag != 0 {
		find[sortKey] = bson.M{flagKey: flag}
	}

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{sortKey: sortType})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func CountAddressOrderBrc20ModelList(net, tick, address string, orderType model.OrderType, orderState model.OrderState) (int64, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}

	buyer := bson.M{"buyerAddress": address}
	seller := bson.M{"sellerAddress": address}

	find := bson.M{
		OR_:     []bson.M{buyer, seller},
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if orderState != 0 {
		find["orderState"] = orderState
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindAddressOrderBrc20ModelList(net, tick, address string,
	orderType model.OrderType, orderState model.OrderState,
	limit int64, flag, page int64, sortKey string, sortType int64) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	buyer := bson.M{"buyerAddress": address}
	seller := bson.M{"sellerAddress": address}

	find := bson.M{
		OR_:     []bson.M{buyer, seller},
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if orderState != 0 {
		if orderState == model.OrderStateAll {
			find["orderState"] = bson.M{IN_: []model.OrderState{
				model.OrderStateCreate,
				model.OrderStateFinish,
				model.OrderStateCancel,
				model.OrderStateErr,
			}}
		} else {
			find["orderState"] = orderState
		}
	}

	switch sortKey {
	case "coinRatePrice":
		sortKey = "coinRatePrice"
	default:
		sortKey = "timestamp"
	}

	flagKey := GT_
	if sortType >= 0 {
		sortType = 1
		flagKey = GT_
	} else {
		sortType = -1
		flagKey = LT_
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	} else if flag != 0 {
		find[sortKey] = bson.M{flagKey: flag}
	}

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{sortKey: sortType})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func FindOrderBrc20ModelListByTimestamp(net, tick string,
	orderType model.OrderType, orderState model.OrderState, limit, startTIme, endTime int64) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}

	if orderType != 0 {
		find["orderType"] = orderType
	}
	if orderState != 0 {
		find["orderState"] = orderState
	}

	if startTIme != 0 {

		between := bson.M{GTE_: startTIme}
		if endTime != 0 {
			between[LTE_] = endTime
		}
		//or := make([]bson.M, 0)
		//or = append(or, bson.M{"timestamp": between})
		//or = append(or, bson.M{"dealTime": between})
		//find[OR_] = or

		between["timestamp"] = between
	}

	sortKey := "timestamp"

	skip := int64(0)

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{sortKey: 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func FindLastOrderBrc20ModelFinish(net, tick string, orderType model.OrderType, orderState model.OrderState) (*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"tick", tick},
		//{"orderType", orderType},
		{"orderState", orderState},
		//{"state", model.STATE_EXIST},
	}
	if orderType != 0 {
		queryBson = append(queryBson, bson.E{Key: "orderType", Value: orderType})
	}
	sort := options.FindOne().SetSort(bson.M{"dealTime": -1})
	entity := &model.OrderBrc20Model{}
	err = collection.FindOne(context.TODO(), queryBson, sort).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func FindLastOrderBrc20ModelFinishList(net, tick string, limit int64, orderType model.OrderType, orderState model.OrderState) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"tick", tick},
		//{"orderType", orderType},
		{"orderState", orderState},
		//{"state", model.STATE_EXIST},
	}
	if orderType != 0 {
		queryBson = append(queryBson, bson.E{Key: "orderType", Value: orderType})
	}
	sort := options.Find().SetSort(bson.M{"dealTime": -1})
	pagination := options.Find().SetLimit(limit).SetSkip(0)

	models := make([]*model.OrderBrc20Model, 0)
	if cursor, err := collection.Find(context.TODO(), queryBson, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func FindOrderBrc20BidDummyModelByDummyId(dummyId string) (*model.OrderBrc20BidDummyModel, error) {
	collection, err := model.OrderBrc20BidDummyModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"dummyId", dummyId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderBrc20BidDummyModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createOrderBrc20BidDummyModel(orderBrc20BidDummy *model.OrderBrc20BidDummyModel) (*model.OrderBrc20BidDummyModel, error) {
	collection, err := model.OrderBrc20BidDummyModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "dummyId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "orderId")
	CreateIndex(collection, "address")
	CreateIndex(collection, "dummyState")
	CreateIndex(collection, "timestamp")

	entity := &model.OrderBrc20BidDummyModel{
		Id:         util.GetUUIDInt64(),
		Net:        orderBrc20BidDummy.Net,
		DummyId:    orderBrc20BidDummy.DummyId,
		OrderId:    orderBrc20BidDummy.OrderId,
		Tick:       orderBrc20BidDummy.Tick,
		Address:    orderBrc20BidDummy.Address,
		DummyState: orderBrc20BidDummy.DummyState,
		Timestamp:  orderBrc20BidDummy.Timestamp,
		CreateTime: util.Time(),
		State:      model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetOrderBrc20BidDummyModel(orderBrc20BidDummy *model.OrderBrc20BidDummyModel) (*model.OrderBrc20BidDummyModel, error) {
	entity, err := FindOrderBrc20BidDummyModelByDummyId(orderBrc20BidDummy.DummyId)
	if err == nil && entity != nil {
		collection, err := model.OrderBrc20BidDummyModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"dummyId", orderBrc20BidDummy.DummyId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: orderBrc20BidDummy.Net})
		bsonData = append(bsonData, bson.E{Key: "dummyId", Value: orderBrc20BidDummy.DummyId})
		bsonData = append(bsonData, bson.E{Key: "orderId", Value: orderBrc20BidDummy.OrderId})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: orderBrc20BidDummy.Tick})
		bsonData = append(bsonData, bson.E{Key: "address", Value: orderBrc20BidDummy.Address})
		bsonData = append(bsonData, bson.E{Key: "dummyState", Value: orderBrc20BidDummy.DummyState})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return orderBrc20BidDummy, nil
	} else {
		return createOrderBrc20BidDummyModel(orderBrc20BidDummy)
	}
}

func CountOrderBrc20BidDummyModelList(orderId, buyerAddress string, dummyState model.DummyState) (int64, error) {
	collection, err := model.OrderBrc20BidDummyModel{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if orderId != "" {
		find["orderId"] = orderId
	}
	if buyerAddress != "" {
		find["buyerAddress"] = buyerAddress
	}
	if dummyState != 0 {
		find["dummyState"] = dummyState
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindOrderBrc20BidDummyModelList(orderId, buyerAddress string, dummyState model.DummyState, skip, limit int64) ([]*model.OrderBrc20BidDummyModel, error) {
	collection, err := model.OrderBrc20BidDummyModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if orderId != "" {
		find["orderId"] = orderId
	}
	if buyerAddress != "" {
		find["address"] = buyerAddress
	}
	if dummyState != 0 {
		find["dummyState"] = dummyState
	}

	models := make([]*model.OrderBrc20BidDummyModel, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(0)
	sort := options.Find().SetSort(bson.M{"updateTime": -1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20BidDummyModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20BidDummyModel Error")
	}
	return models, nil
}

func FindOrderBrc20MarketPriceModelByPair(net, pair string) (*model.OrderBrc20MarketPriceModel, error) {
	collection, err := model.OrderBrc20MarketPriceModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"pair", pair},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderBrc20MarketPriceModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createOrderBrc20MarketPriceModel(orderBrc20MarketPrice *model.OrderBrc20MarketPriceModel) (*model.OrderBrc20MarketPriceModel, error) {
	collection, err := model.OrderBrc20MarketPriceModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateIndex(collection, "net")
	CreateIndex(collection, "pair")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "timestamp")

	entity := &model.OrderBrc20MarketPriceModel{
		Id:         util.GetUUIDInt64(),
		Net:        orderBrc20MarketPrice.Net,
		Pair:       orderBrc20MarketPrice.Pair,
		Tick:       orderBrc20MarketPrice.Tick,
		Price:      orderBrc20MarketPrice.Price,
		Timestamp:  orderBrc20MarketPrice.Timestamp,
		CreateTime: util.Time(),
		State:      model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetOrderBrc20MarketPriceModel(orderBrc20MarketPrice *model.OrderBrc20MarketPriceModel) (*model.OrderBrc20MarketPriceModel, error) {
	entity, err := FindOrderBrc20MarketPriceModelByPair(orderBrc20MarketPrice.Net, orderBrc20MarketPrice.Pair)
	if err == nil && entity != nil {
		collection, err := model.OrderBrc20MarketPriceModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"net", orderBrc20MarketPrice.Net},
			{"pair", orderBrc20MarketPrice.Pair},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: orderBrc20MarketPrice.Net})
		bsonData = append(bsonData, bson.E{Key: "pair", Value: orderBrc20MarketPrice.Pair})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: orderBrc20MarketPrice.Tick})
		bsonData = append(bsonData, bson.E{Key: "price", Value: orderBrc20MarketPrice.Price})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return orderBrc20MarketPrice, nil
	} else {
		return createOrderBrc20MarketPriceModel(orderBrc20MarketPrice)
	}
}

func CountBuyerOrderBrc20ModelList(net, tick, buyerAddress, buyerIp string, orderType model.OrderType, orderState model.OrderState, startTime, endTime int64, coinAmount int64) (int64, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if buyerAddress != "" {
		find["buyerAddress"] = buyerAddress
	}
	if buyerIp != "" {
		find["buyerIp"] = buyerIp
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if orderState != 0 {
		find["orderState"] = orderState
	}
	if coinAmount != 0 {
		find["coinAmount"] = coinAmount
	}

	between := bson.M{}
	if startTime != 0 {
		between[GTE_] = startTime
	}
	if endTime != 0 {
		between[LT_] = endTime
	}
	if startTime != 0 || endTime != 0 {
		find["dealTime"] = between
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindWhitelistModelByAddressId(addressId string) (*model.WhitelistModel, error) {
	collection, err := model.WhitelistModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"addressId", addressId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.WhitelistModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createWhitelistModel(whitelist *model.WhitelistModel) (*model.WhitelistModel, error) {
	collection, err := model.WhitelistModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "addressId")
	CreateIndex(collection, "ip")
	CreateIndex(collection, "address")
	CreateIndex(collection, "whitelistType")
	CreateIndex(collection, "whiteUseState")

	entity := &model.WhitelistModel{
		Id:            util.GetUUIDInt64(),
		AddressId:     whitelist.AddressId,
		IP:            whitelist.IP,
		Address:       whitelist.Address,
		WhitelistType: whitelist.WhitelistType,
		WhiteUseState: whitelist.WhiteUseState,
		Limit:         whitelist.Limit,
		Timestamp:     whitelist.Timestamp,
		CreateTime:    util.Time(),
		State:         model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetWhitelistModel(whitelist *model.WhitelistModel) (*model.WhitelistModel, error) {
	entity, err := FindWhitelistModelByAddressId(whitelist.AddressId)
	if err == nil && entity != nil {
		collection, err := model.WhitelistModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"addressId", whitelist.AddressId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "addressId", Value: whitelist.AddressId})
		bsonData = append(bsonData, bson.E{Key: "address", Value: whitelist.Address})
		bsonData = append(bsonData, bson.E{Key: "ip", Value: whitelist.IP})
		bsonData = append(bsonData, bson.E{Key: "whitelistType", Value: whitelist.WhitelistType})
		bsonData = append(bsonData, bson.E{Key: "whiteUseState", Value: whitelist.WhiteUseState})
		bsonData = append(bsonData, bson.E{Key: "limit", Value: whitelist.Limit})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: whitelist.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return whitelist, nil
	} else {
		return createWhitelistModel(whitelist)
	}
}

func FindWhitelistModelByAddressAndType(address string, whitelistType model.WhitelistType) (*model.WhitelistModel, error) {
	collection, err := model.WhitelistModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"address", address},
		{"whitelistType", whitelistType},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.WhitelistModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func FindWhitelistModelByIpAndType(ip string, whitelistType model.WhitelistType) (*model.WhitelistModel, error) {
	collection, err := model.WhitelistModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"ip", ip},
		{"whitelistType", whitelistType},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.WhitelistModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func FindOrderBrc20ModelToolList(net, tick, sellerAddress, buyerAddress, buyerIp, inscriptionId string,
	orderType model.OrderType, orderState model.OrderState,
	limit int64, flag, page int64, sortKey string, sortType int64, freeState model.FreeState, coinAmount int64) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if inscriptionId != "" {
		find["inscriptionId"] = inscriptionId
	}
	if sellerAddress != "" {
		find["sellerAddress"] = sellerAddress
	}
	if buyerAddress != "" {
		find["buyerAddress"] = buyerAddress
	}
	if buyerIp != "" {
		find["buyerIp"] = buyerIp
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if coinAmount != 0 {
		find["coinAmount"] = coinAmount
	}
	if orderState != 0 {
		if orderState == model.OrderStateAll {
			find["orderState"] = bson.M{IN_: []model.OrderState{
				model.OrderStateCreate,
				model.OrderStateFinish,
				model.OrderStateCancel,
				model.OrderStateErr,
			}}
		} else {
			find["orderState"] = orderState
		}
	}

	if freeState != 0 {
		find["freeState"] = freeState
	}

	switch sortKey {
	case "coinRatePrice":
		sortKey = "coinRatePrice"
	case "dealTime":
		sortKey = "dealTime"
	default:
		sortKey = "timestamp"
	}

	flagKey := GT_
	if sortType >= 0 {
		sortType = 1
		flagKey = GT_
	} else {
		sortType = -1
		flagKey = LT_
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	} else if flag != 0 {
		find[sortKey] = bson.M{flagKey: flag}
	}

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{sortKey: sortType})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func FindOrderBrc20ModelListByDealTimestamp(net, tick string,
	orderType model.OrderType, orderState model.OrderState, limit, startTIme, endTime int64) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}

	if orderType != 0 {
		find["orderType"] = orderType
	}
	if orderState != 0 {
		find["orderState"] = orderState
	}

	if startTIme != 0 {

		between := bson.M{GTE_: startTIme}
		if endTime != 0 {
			between[LTE_] = endTime
		}

		between["dealTime"] = between
	}

	sortKey := "dealTime"

	skip := int64(0)

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{sortKey: 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func SetOrderBrc20ModelForInscriptionState(orderBrc20 *model.OrderBrc20Model) error {
	entity, err := FindOrderBrc20ModelByOrderId(orderBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.OrderBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", orderBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "inscriptionState", Value: orderBrc20.InscriptionState})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func SetOrderBrc20ModelForOrderState(orderBrc20 *model.OrderBrc20Model) error {
	entity, err := FindOrderBrc20ModelByOrderId(orderBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.OrderBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", orderBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "orderState", Value: orderBrc20.OrderState})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func FindOrderBrc20MarketInfoModelByPair(net, date string) (*model.OrderBrc20MarketInfoModel, error) {
	collection, err := model.OrderBrc20MarketInfoModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"date", date},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderBrc20MarketInfoModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createOrderBrc20MarketInfoModel(orderBrc20MarketInfo *model.OrderBrc20MarketInfoModel) (*model.OrderBrc20MarketInfoModel, error) {
	collection, err := model.OrderBrc20MarketInfoModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "netDate")
	CreateIndex(collection, "net")
	CreateIndex(collection, "date")
	CreateIndex(collection, "bidVolume")
	CreateIndex(collection, "timestamp")

	entity := &model.OrderBrc20MarketInfoModel{
		Id:         util.GetUUIDInt64(),
		NetDate:    orderBrc20MarketInfo.NetDate,
		Net:        orderBrc20MarketInfo.Net,
		Date:       orderBrc20MarketInfo.Date,
		AskVolume:  orderBrc20MarketInfo.AskVolume,
		BidVolume:  orderBrc20MarketInfo.BidVolume,
		AskFees:    orderBrc20MarketInfo.AskFees,
		BidFees:    orderBrc20MarketInfo.BidFees,
		Between:    orderBrc20MarketInfo.Between,
		Timestamp:  orderBrc20MarketInfo.Timestamp,
		CreateTime: util.Time(),
		State:      model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetOrderBrc20MarketInfoModel(orderBrc20MarketInfo *model.OrderBrc20MarketInfoModel) (*model.OrderBrc20MarketInfoModel, error) {
	entity, err := FindOrderBrc20MarketInfoModelByPair(orderBrc20MarketInfo.Net, orderBrc20MarketInfo.Date)
	if err == nil && entity != nil {
		collection, err := model.OrderBrc20MarketInfoModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"net", orderBrc20MarketInfo.Net},
			{"date", orderBrc20MarketInfo.Date},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "netDate", Value: orderBrc20MarketInfo.NetDate})
		bsonData = append(bsonData, bson.E{Key: "net", Value: orderBrc20MarketInfo.Net})
		bsonData = append(bsonData, bson.E{Key: "date", Value: orderBrc20MarketInfo.Date})
		bsonData = append(bsonData, bson.E{Key: "askVolume", Value: orderBrc20MarketInfo.AskVolume})
		bsonData = append(bsonData, bson.E{Key: "bidVolume", Value: orderBrc20MarketInfo.BidVolume})
		bsonData = append(bsonData, bson.E{Key: "askFees", Value: orderBrc20MarketInfo.AskFees})
		bsonData = append(bsonData, bson.E{Key: "bidFees", Value: orderBrc20MarketInfo.BidFees})
		bsonData = append(bsonData, bson.E{Key: "between", Value: orderBrc20MarketInfo.Between})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: orderBrc20MarketInfo.Timestamp})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return orderBrc20MarketInfo, nil
	} else {
		return createOrderBrc20MarketInfoModel(orderBrc20MarketInfo)
	}
}

func CountOrderBrc20ModelListForPoolOrderId(poolOrderId, buyerAddress string) (int64, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"orderState": model.OrderStateCreate,
		"state":      model.STATE_EXIST,
	}
	//if net != "" {
	//	find["net"] = net
	//}
	//if tick != "" {
	//	find["tick"] = tick
	//}
	if poolOrderId != "" {
		find["poolOrderId"] = poolOrderId
	}
	if buyerAddress != "" {
		find["buyerAddress"] = buyerAddress
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindOrderNotificationModelByAddressAndNotificationType(address string, notificationType model.NotificationType) (*model.OrderNotificationModel, error) {
	collection, err := model.OrderNotificationModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"address", address},
		{"notificationType", notificationType},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderNotificationModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createOrderNotificationModel(orderNotification *model.OrderNotificationModel) (*model.OrderNotificationModel, error) {
	collection, err := model.OrderNotificationModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateIndex(collection, "address")
	CreateIndex(collection, "notificationType")
	CreateIndex(collection, "timestamp")

	entity := &model.OrderNotificationModel{
		Id:                util.GetUUIDInt64(),
		Address:           orderNotification.Address,
		NotificationType:  orderNotification.NotificationType,
		NotificationCount: orderNotification.NotificationCount,
		Timestamp:         util.Time(),
		CreateTime:        util.Time(),
		State:             model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetOrderNotificationModel(orderNotification *model.OrderNotificationModel) (*model.OrderNotificationModel, error) {
	entity, err := FindOrderNotificationModelByAddressAndNotificationType(orderNotification.Address, orderNotification.NotificationType)
	if err == nil && entity != nil {
		collection, err := model.OrderNotificationModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"address", orderNotification.Address},
			{"notificationType", orderNotification.NotificationType},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "address", Value: orderNotification.Address})
		bsonData = append(bsonData, bson.E{Key: "notificationType", Value: orderNotification.NotificationType})
		bsonData = append(bsonData, bson.E{Key: "notificationCount", Value: orderNotification.NotificationCount})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: util.Time()})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return orderNotification, nil
	} else {
		return createOrderNotificationModel(orderNotification)
	}
}

func CountOrderNotificationModelList(address string) (int64, error) {
	collection, err := model.OrderNotificationModel{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"address": address,
		"state":   model.STATE_EXIST,
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindOrderNotificationModelList(address string) ([]*model.OrderNotificationModel, error) {
	collection, err := model.OrderNotificationModel{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"address": address,
		"state":   model.STATE_EXIST,
	}

	models := make([]*model.OrderNotificationModel, 0)
	pagination := options.Find().SetLimit(10).SetSkip(0)
	sort := options.Find().SetSort(bson.M{"timestamp": -1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderNotificationModel{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderNotificationModel Error")
	}
	return models, nil
}

func FindOrderBrc20ModelListByPoolOrderId(poolOrderId string) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"orderState": model.OrderStateCreate,
		"state":      model.STATE_EXIST,
	}
	//if net != "" {
	//	find["net"] = net
	//}
	//if tick != "" {
	//	find["tick"] = tick
	//}
	if poolOrderId != "" {
		find["poolOrderId"] = poolOrderId
	}
	sortKey := "timestamp"

	skip := int64(0)

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(1000).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{sortKey: 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func FindUsedInscriptionOrder(inscriptionId string) (int64, error) {
	if strings.Contains(inscriptionId, "i") {
		inscriptionId = strings.ReplaceAll(inscriptionId, "i", ":")
	}

	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"inscriptionId": inscriptionId,
		"orderState":    model.OrderStateCreate,
		"state":         model.STATE_EXIST,
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}
func FindUsedInscriptionOrderV2(inscriptionId string) (int64, error) {
	if strings.Contains(inscriptionId, ":") {
		inscriptionId = strings.ReplaceAll(inscriptionId, ":", "i")
	}

	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"inscriptionId": inscriptionId,
		"orderState":    model.OrderStateCreate,
		"state":         model.STATE_EXIST,
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindSoldInscriptionOrder(inscriptionId string) (int64, error) {
	if strings.Contains(inscriptionId, ":") {
		inscriptionId = strings.ReplaceAll(inscriptionId, ":", "i")
	}

	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"sellInscriptionId": inscriptionId,
		"orderState":        model.OrderStateFinish,
		"state":             model.STATE_EXIST,
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindDealOrderBrc20ModelListByDealStartAndDealEndBlock(net, tick string,
	orderType model.OrderType, orderState model.OrderState,
	limit, page int64, startBlock, endBlock int64, version int) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if version != 0 {
		find["version"] = version
	}
	if orderState != 0 {
		if orderState == model.OrderStateAll {
			find["orderState"] = bson.M{IN_: []model.OrderState{
				model.OrderStateCreate,
				model.OrderStateFinish,
				model.OrderStateCancel,
				model.OrderStateErr,
			}}
		} else {
			find["orderState"] = orderState
		}
	}

	if startBlock != 0 && endBlock != 0 {
		between := bson.M{
			GTE_: startBlock,
			LTE_: endBlock,
		}
		find["dealTxBlock"] = between
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	}

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"dealTxBlock": 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func SetOrderBrc20ModelForCalReward(orderBrc20 *model.OrderBrc20Model) error {
	entity, err := FindOrderBrc20ModelByOrderId(orderBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.OrderBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", orderBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "calValue", Value: orderBrc20.CalValue})
		bsonData = append(bsonData, bson.E{Key: "calTotalValue", Value: orderBrc20.CalTotalValue})
		bsonData = append(bsonData, bson.E{Key: "calStartBlock", Value: orderBrc20.CalStartBlock})
		bsonData = append(bsonData, bson.E{Key: "calEndBlock", Value: orderBrc20.CalEndBlock})
		bsonData = append(bsonData, bson.E{Key: "percentage", Value: orderBrc20.Percentage})
		bsonData = append(bsonData, bson.E{Key: "rewardAmount", Value: orderBrc20.RewardAmount})
		bsonData = append(bsonData, bson.E{Key: "rewardRealAmount", Value: orderBrc20.RewardRealAmount})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func FindOrderBrc20ModelListByDealTime(net, tick string,
	orderType model.OrderType, orderState model.OrderState,
	limit, page int64, dealTxBlockState model.ClaimTxBlockState, version int) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if dealTxBlockState != 0 {
		find["dealTxBlockState"] = dealTxBlockState
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if version != 0 {
		find["version"] = version
	}
	if orderState != 0 {
		if orderState == model.OrderStateAll {
			find["orderState"] = bson.M{IN_: []model.OrderState{
				model.OrderStateCreate,
				model.OrderStateFinish,
				model.OrderStateCancel,
				model.OrderStateErr,
			}}
		} else {
			find["orderState"] = orderState
		}
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	}

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"dealTime": 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func CountEventOrderBrc20ModelList(net, tick, address string,
	orderType model.OrderType, orderState model.OrderState,
	version int, eventTime int64) (int64, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if orderState != 0 {
		find["orderState"] = orderState
	}
	if version != 0 {
		find["version"] = version
	}
	if address != "" {
		find["$or"] = []bson.M{
			{"sellerAddress": address},
			{"buyerAddress": address},
		}
	}
	if eventTime != 0 {
		find["dealTime"] = bson.M{
			GTE_: eventTime,
		}
	}

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindEventOrderBrc20ModelList(net, tick, address string,
	orderType model.OrderType, orderState model.OrderState,
	limit, page int64, version int, eventTime int64) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state": model.STATE_EXIST,
	}
	if net != "" {
		find["net"] = net
	}
	if tick != "" {
		find["tick"] = tick
	}
	if orderType != 0 {
		find["orderType"] = orderType
	}
	if version != 0 {
		find["version"] = version
	}
	if orderState != 0 {
		if orderState == model.OrderStateAll {
			find["orderState"] = bson.M{IN_: []model.OrderState{
				model.OrderStateCreate,
				model.OrderStateFinish,
				model.OrderStateCancel,
				model.OrderStateErr,
			}}
		} else {
			find["orderState"] = orderState
		}
	}
	if address != "" {
		find["$or"] = []bson.M{
			{"sellerAddress": address},
			{"buyerAddress": address},
		}
	}
	if eventTime != 0 {
		find["dealTime"] = bson.M{
			GTE_: eventTime,
		}
	}

	skip := int64(0)
	if page != 0 {
		skip = (page - 1) * limit
	}

	models := make([]*model.OrderBrc20Model, 0)
	pagination := options.Find().SetLimit(limit).SetSkip(skip)
	sort := options.Find().SetSort(bson.M{"dealTime": 1})
	if cursor, err := collection.Find(context.TODO(), find, pagination, sort); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			entity := &model.OrderBrc20Model{}
			if err = cursor.Decode(entity); err == nil {
				models = append(models, entity)
			}
		}
	} else {
		return nil, errors.New("Get OrderBrc20Model Error")
	}
	return models, nil
}

func SetOrderBrc20ModelForDealBlock(orderBrc20 *model.OrderBrc20Model) error {
	entity, err := FindOrderBrc20ModelByOrderId(orderBrc20.OrderId)
	if err == nil && entity != nil {
		collection, err := model.OrderBrc20Model{}.GetWriteDB()
		if err != nil {
			return err
		}
		filter := bson.D{
			{"orderId", orderBrc20.OrderId},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "dealTxBlock", Value: orderBrc20.DealTxBlock})
		bsonData = append(bsonData, bson.E{Key: "dealTxBlockState", Value: orderBrc20.DealTxBlockState})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func CountOwnEventOrderBrc20RewardBySeller(net, tick, sellerAddress string, eventTime int64) (*model.EventRewardCount, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	countInfo := &model.EventRewardCount{
		Id:                sellerAddress,
		AmountTotal:       0,
		RewardAmountTotal: 0,
	}
	countInfoList := make([]model.EventRewardCount, 0)

	match := bson.M{
		"net":              net,
		"dealTxBlockState": model.ClaimTxBlockStateConfirmed,
		"orderType":        model.OrderTypeBuy,
		"orderState":       model.OrderStateFinish,
		"version":          2,
		//"dealCoinTxBlock":   bson.M{GTE_: config.PlatformRewardCalStartBlock},
	}
	if tick != "" {
		match["tick"] = tick
	}
	if sellerAddress != "" {
		match["sellerAddress"] = sellerAddress
	}
	if eventTime != 0 {
		match["dealTime"] = bson.M{
			GTE_: eventTime,
		}
	}

	pipeline := mongo.Pipeline{
		{
			{"$match", match},
		},
		{
			{"$group", bson.D{
				{"_id", "$sellerAddress"},
				{"amountTotal", bson.D{
					{"$sum", "$amount"},
				}},
				{"orderCounts", bson.D{
					{"$sum", 1},
				}},
				{"rewardAmountTotal", bson.D{
					{"$sum", "$rewardRealAmount"},
				}},
			}},
		},
	}
	if cursor, err := collection.Aggregate(context.Background(), pipeline); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			var entity model.EventRewardCount
			if err = cursor.Decode(&entity); err == nil {
				countInfoList = append(countInfoList, entity)
			}
		}
		if countInfoList != nil && len(countInfoList) != 0 {
			for _, v := range countInfoList {
				if v.Id == sellerAddress {
					countInfo = &v
					break
				}
			}
		}
		return countInfo, nil
	} else {
		return nil, errors.New("db get EventRewardCount error")
	}
}

func CountOwnEventOrderBrc20RewardByBuyer(net, tick, pair, buyerAddress string, eventTime int64) (*model.EventRewardCount, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	countInfo := &model.EventRewardCount{
		Id:                buyerAddress,
		AmountTotal:       0,
		RewardAmountTotal: 0,
	}
	countInfoList := make([]model.EventRewardCount, 0)

	match := bson.M{
		"net":              net,
		"dealTxBlockState": model.ClaimTxBlockStateConfirmed,
		"orderType":        model.OrderTypeBuy,
		"orderState":       model.OrderStateFinish,
		"version":          2,
		//"dealCoinTxBlock":   bson.M{GTE_: config.PlatformRewardCalStartBlock},
	}
	if tick != "" {
		match["tick"] = tick
	}
	if pair != "" {
		match["pair"] = pair
	}
	if buyerAddress != "" {
		match["buyerAddress"] = buyerAddress
	}
	if eventTime != 0 {
		match["dealTime"] = bson.M{
			GTE_: eventTime,
		}
	}

	pipeline := mongo.Pipeline{
		{
			{"$match", match},
		},
		{
			{"$group", bson.D{
				{"_id", "$buyerAddress"},
				{"amountTotal", bson.D{
					{"$sum", "$amount"},
				}},
				{"orderCounts", bson.D{
					{"$sum", 1},
				}},
				{"rewardAmountTotal", bson.D{
					{"$sum", "$rewardRealAmount"},
				}},
			}},
		},
	}
	if cursor, err := collection.Aggregate(context.Background(), pipeline); err == nil {
		defer cursor.Close(context.Background())
		for cursor.Next(context.Background()) {
			var entity model.EventRewardCount
			if err = cursor.Decode(&entity); err == nil {
				countInfoList = append(countInfoList, entity)
			}
		}
		if countInfoList != nil && len(countInfoList) != 0 {
			for _, v := range countInfoList {
				if v.Id == buyerAddress {
					countInfo = &v
					break
				}
			}
		}
		return countInfo, nil
	} else {
		return nil, errors.New("db get EventRewardCount error")
	}
}

func FindOrderBrc20ModelTickAndAmountByOrderId(orderId string) (string, uint64, uint64, int64, int64, int64, int64) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return "", 0, 0, 0, 0, 0, 0
	}
	queryBson := bson.D{
		{"orderId", orderId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderBrc20Model{}
	projection := options.FindOne().SetProjection(bson.M{
		"id":               1,
		"_id":              1,
		"orderId":          1,
		"tick":             1,
		"coinAmount":       1,
		"amount":           1,
		"dealTxBlock":      1,
		"dealTime":         1,
		"percentage":       1,
		"rewardAmount":     1,
		"rewardRealAmount": 1,
		"state":            1,
	})
	err = collection.FindOne(context.TODO(), queryBson, projection).Decode(entity)
	if err != nil {
		return "", 0, 0, 0, 0, 0, 0
	}
	return entity.Tick, entity.CoinAmount, entity.Amount, entity.RewardRealAmount, entity.Percentage, entity.DealTxBlock, entity.DealTime
}

func FindOrderBrc20ModelPoolModeByOrderId(orderId string) model.PoolMode {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return model.PoolModeDefault
	}
	queryBson := bson.D{
		{"orderId", orderId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderBrc20Model{}
	projection := options.FindOne().SetProjection(bson.M{
		"id":            1,
		"_id":           1,
		"orderId":       1,
		"tick":          1,
		"poolOrderMode": 1,
		"state":         1,
	})
	err = collection.FindOne(context.TODO(), queryBson, projection).Decode(entity)
	if err != nil {
		return model.PoolModeDefault
	}
	return entity.PoolOrderMode
}

func FindOrderBrc20ModelAmountByOrderId(orderId string) uint64 {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0
	}
	queryBson := bson.D{
		{"orderId", orderId},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderBrc20Model{}
	projection := options.FindOne().SetProjection(bson.M{
		"id":      1,
		"_id":     1,
		"orderId": 1,
		"tick":    1,
		"amount":  1,
		"state":   1,
	})
	err = collection.FindOne(context.TODO(), queryBson, projection).Decode(entity)
	if err != nil {
		return 0
	}
	return entity.Amount
}

func FindOrderCirculationModelByTick(net, tick string) (*model.OrderCirculationModel, error) {
	collection, err := model.OrderCirculationModel{}.GetReadDB()
	if err != nil {
		return nil, err
	}
	queryBson := bson.D{
		{"net", net},
		{"tick", tick},
		//{"state", model.STATE_EXIST},
	}
	entity := &model.OrderCirculationModel{}
	err = collection.FindOne(context.TODO(), queryBson).Decode(entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func createOrderCirculationModel(orderCirculation *model.OrderCirculationModel) (*model.OrderCirculationModel, error) {
	collection, err := model.OrderCirculationModel{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")

	entity := &model.OrderCirculationModel{
		Id:                util.GetUUIDInt64(),
		Net:               orderCirculation.Net,
		Tick:              orderCirculation.Tick,
		CirculationSupply: orderCirculation.CirculationSupply,
		TotalSupply:       orderCirculation.TotalSupply,
		CreateTime:        util.Time(),
		State:             model.STATE_EXIST,
	}

	_, err = collection.InsertOne(context.TODO(), entity)
	if err != nil {
		return nil, err
	} else {
		//id := res.InsertedID
		//fmt.Println("insert id :", id)
		return entity, nil
	}
}

func SetOrderCirculationModel(orderCirculation *model.OrderCirculationModel) (*model.OrderCirculationModel, error) {
	entity, err := FindOrderCirculationModelByTick(orderCirculation.Net, orderCirculation.Tick)
	if err == nil && entity != nil {
		collection, err := model.OrderCirculationModel{}.GetWriteDB()
		if err != nil {
			return nil, err
		}
		filter := bson.D{
			{"net", orderCirculation.Net},
			{"tick", orderCirculation.Tick},
			//{"state", model.STATE_EXIST},
		}
		bsonData := bson.D{}
		bsonData = append(bsonData, bson.E{Key: "net", Value: orderCirculation.Net})
		bsonData = append(bsonData, bson.E{Key: "tick", Value: orderCirculation.Tick})
		bsonData = append(bsonData, bson.E{Key: "circulationSupply", Value: orderCirculation.CirculationSupply})
		bsonData = append(bsonData, bson.E{Key: "totalSupply", Value: orderCirculation.TotalSupply})
		bsonData = append(bsonData, bson.E{Key: "updateTime", Value: util.Time()})
		update := bson.D{{"$set",
			bsonData,
		}}
		_, err = collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			return nil, err
		}
		return orderCirculation, nil
	} else {
		return createOrderCirculationModel(orderCirculation)
	}
}
