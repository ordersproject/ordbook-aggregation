package mongo_service

import (
	"context"
	"errors"
	"github.com/godaddy-x/jorm/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"ordbook-aggregation/model"
)

func FindOrderBrc20ModelByOrderId(orderId string) (*model.OrderBrc20Model, error) {
	collection, err :=  model.OrderBrc20Model{}.GetReadDB()
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


func createOrderBrc20Model(orderBrc20 *model.OrderBrc20Model) (*model.OrderBrc20Model, error)  {
	collection, err := model.OrderBrc20Model{}.GetWriteDB()
	if err != nil {
		return nil, err
	}

	CreateUniqueIndex(collection, "orderId")
	CreateIndex(collection, "net")
	CreateIndex(collection, "tick")
	CreateIndex(collection, "coinRatePrice")
	CreateIndex(collection, "metaBlockHeight")
	CreateIndex(collection, "orderState")
	CreateIndex(collection, "orderType")
	CreateIndex(collection, "sellerAddress")
	CreateIndex(collection, "buyerAddress")
	CreateIndex(collection, "buyerIp")
	CreateIndex(collection, "timestamp")
	CreateIndex(collection, "dealTime")

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
		OrderState:          orderBrc20.OrderState,
		OrderType:           orderBrc20.OrderType,
		SellerAddress:       orderBrc20.SellerAddress,
		BuyerAddress:        orderBrc20.BuyerAddress,
		BuyerIp:        orderBrc20.BuyerIp,
		MarketAmount:        orderBrc20.MarketAmount,
		PlatformFee:         orderBrc20.PlatformFee,
		ChangeAmount:        orderBrc20.ChangeAmount,
		Fee:                 orderBrc20.Fee,
		FeeRate:             orderBrc20.FeeRate,
		SupplementaryAmount: orderBrc20.SupplementaryAmount,
		PlatformTx:          orderBrc20.PlatformTx,
		InscriptionId:       orderBrc20.InscriptionId,
		InscriptionNumber:   orderBrc20.InscriptionNumber,
		PsbtRawPreAsk:       orderBrc20.PsbtRawPreAsk,
		PsbtRawFinalAsk:     orderBrc20.PsbtRawFinalAsk,
		PsbtAskTxId:         orderBrc20.PsbtAskTxId,
		PsbtRawPreBid:       orderBrc20.PsbtRawPreBid,
		PsbtRawMidBid:       orderBrc20.PsbtRawMidBid,
		PsbtRawFinalBid:     orderBrc20.PsbtRawFinalBid,
		PsbtBidTxId:         orderBrc20.PsbtBidTxId,
		Integral:            orderBrc20.Integral,
		FreeState:           orderBrc20.FreeState,
		DealTime:            orderBrc20.DealTime,
		Timestamp:           orderBrc20.Timestamp,
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

func SetOrderBrc20Model(orderBrc20 *model.OrderBrc20Model) (*model.OrderBrc20Model, error)  {
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
		bsonData = append(bsonData, bson.E{Key: "orderState", Value: orderBrc20.OrderState})
		bsonData = append(bsonData, bson.E{Key: "orderType", Value: orderBrc20.OrderType})
		bsonData = append(bsonData, bson.E{Key: "sellerAddress", Value: orderBrc20.SellerAddress})
		bsonData = append(bsonData, bson.E{Key: "buyerAddress", Value: orderBrc20.BuyerAddress})
		bsonData = append(bsonData, bson.E{Key: "buyerIp", Value: orderBrc20.BuyerIp})
		bsonData = append(bsonData, bson.E{Key: "marketAmount", Value: orderBrc20.MarketAmount})
		bsonData = append(bsonData, bson.E{Key: "platformFee", Value: orderBrc20.PlatformFee})
		bsonData = append(bsonData, bson.E{Key: "changeAmount", Value: orderBrc20.ChangeAmount})
		bsonData = append(bsonData, bson.E{Key: "fee", Value: orderBrc20.Fee})
		bsonData = append(bsonData, bson.E{Key: "feeRate", Value: orderBrc20.FeeRate})
		bsonData = append(bsonData, bson.E{Key: "supplementaryAmount", Value: orderBrc20.SupplementaryAmount})
		bsonData = append(bsonData, bson.E{Key: "platformTx", Value: orderBrc20.PlatformTx})
		bsonData = append(bsonData, bson.E{Key: "inscriptionId", Value: orderBrc20.InscriptionId})
		bsonData = append(bsonData, bson.E{Key: "inscriptionNumber", Value: orderBrc20.InscriptionNumber})
		bsonData = append(bsonData, bson.E{Key: "psbtRawPreAsk", Value: orderBrc20.PsbtRawPreAsk})
		bsonData = append(bsonData, bson.E{Key: "psbtRawFinalAsk", Value: orderBrc20.PsbtRawFinalAsk})
		bsonData = append(bsonData, bson.E{Key: "psbtAskTxId", Value: orderBrc20.PsbtAskTxId})
		bsonData = append(bsonData, bson.E{Key: "psbtRawPreBid", Value: orderBrc20.PsbtRawPreBid})
		bsonData = append(bsonData, bson.E{Key: "psbtRawMidBid", Value: orderBrc20.PsbtRawMidBid})
		bsonData = append(bsonData, bson.E{Key: "psbtRawFinalBid", Value: orderBrc20.PsbtRawFinalBid})
		bsonData = append(bsonData, bson.E{Key: "psbtBidTxId", Value: orderBrc20.PsbtBidTxId})
		bsonData = append(bsonData, bson.E{Key: "integral", Value: orderBrc20.Integral})
		bsonData = append(bsonData, bson.E{Key: "freeState", Value: orderBrc20.FreeState})
		bsonData = append(bsonData, bson.E{Key: "dealTime", Value: orderBrc20.DealTime})
		bsonData = append(bsonData, bson.E{Key: "timestamp", Value: orderBrc20.Timestamp})
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
	collection, err :=  model.OrderBrc20Model{}.GetReadDB()
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



func CountOrderBrc20ModelList(net, tick, sellerAddress, buyerAddress string, orderType model.OrderType, orderState model.OrderState) (int64, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state":    model.STATE_EXIST,
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

	total, err := collection.CountDocuments(context.TODO(), find)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func FindOrderBrc20ModelList(net, tick, sellerAddress, buyerAddress string,
	orderType model.OrderType, orderState model.OrderState,
	limit int64, flag, page int64, sortKey string, sortType int64) ([]*model.OrderBrc20Model, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return nil, errors.New("db connect error")
	}
	if collection == nil {
		return nil, errors.New("db connect error")
	}

	find := bson.M{
		"state":    model.STATE_EXIST,
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
		if orderState == model.OrderStateAll {
			find["orderState"] = bson.M{IN_:[]model.OrderState{
				model.OrderStateCreate,
				model.OrderStateFinish,
				model.OrderStateCancel,
				model.OrderStateErr,
			}}
		}else {
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
	}else {
		sortType = -1
		flagKey = LT_
	}


	skip := int64(0)
	if page != 0 {
		skip = (page-1)*limit
	}else if flag != 0 {
		find[sortKey] = bson.M{flagKey:flag}
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
		OR_:[]bson.M{buyer, seller},
		"state":    model.STATE_EXIST,
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
		OR_:[]bson.M{buyer, seller},
		"state":    model.STATE_EXIST,
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
			find["orderState"] = bson.M{IN_:[]model.OrderState{
				model.OrderStateCreate,
				model.OrderStateFinish,
				model.OrderStateCancel,
				model.OrderStateErr,
			}}
		}else {
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
	}else {
		sortType = -1
		flagKey = LT_
	}

	skip := int64(0)
	if page != 0 {
		skip = (page-1)*limit
	}else if flag != 0 {
		find[sortKey] = bson.M{flagKey:flag}
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



//
func FindOrderBrc20BidDummyModelByDummyId(dummyId string) (*model.OrderBrc20BidDummyModel, error) {
	collection, err :=  model.OrderBrc20BidDummyModel{}.GetReadDB()
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


func createOrderBrc20BidDummyModel(orderBrc20BidDummy *model.OrderBrc20BidDummyModel) (*model.OrderBrc20BidDummyModel, error)  {
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
		Id:             util.GetUUIDInt64(),
		Net:        orderBrc20BidDummy.Net,
		DummyId:        orderBrc20BidDummy.DummyId,
		OrderId:        orderBrc20BidDummy.OrderId,
		Tick:           orderBrc20BidDummy.Tick,
		Address:         orderBrc20BidDummy.Address,
		DummyState:     orderBrc20BidDummy.DummyState,
		Timestamp:      orderBrc20BidDummy.Timestamp,
		CreateTime:     util.Time(),
		State:          model.STATE_EXIST,
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

func SetOrderBrc20BidDummyModel(orderBrc20BidDummy *model.OrderBrc20BidDummyModel) (*model.OrderBrc20BidDummyModel, error)  {
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
		"state":    model.STATE_EXIST,
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
		"state":    model.STATE_EXIST,
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
	collection, err :=  model.OrderBrc20MarketPriceModel{}.GetReadDB()
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


func createOrderBrc20MarketPriceModel(orderBrc20MarketPrice *model.OrderBrc20MarketPriceModel) (*model.OrderBrc20MarketPriceModel, error)  {
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

func SetOrderBrc20MarketPriceModel(orderBrc20MarketPrice *model.OrderBrc20MarketPriceModel) (*model.OrderBrc20MarketPriceModel, error)  {
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


func CountBuyerOrderBrc20ModelList(net, tick, buyerAddress, buyerIp string, orderType model.OrderType, orderState model.OrderState, startTime, endTime int64) (int64, error) {
	collection, err := model.OrderBrc20Model{}.GetReadDB()
	if err != nil {
		return 0, err
	}
	find := bson.M{
		"state":    model.STATE_EXIST,
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

