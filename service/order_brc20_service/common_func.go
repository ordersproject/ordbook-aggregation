package order_brc20_service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/node"
	"ordbook-aggregation/service/common_service"
	"ordbook-aggregation/service/mempool_space_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"ordbook-aggregation/ws_service/ws"
	"strconv"
	"strings"
)

const (
	coinPriceDecimalNumDefault int32 = 8
)

// Get real price
func GetPrice(coinAmount, coinPrice int64, coinPriceDecimalNum int32) (int64, error) {
	var (
		price        int64
		coinPriceDe  decimal.Decimal = decimal.NewFromInt(coinPrice)
		coinAmountDe decimal.Decimal = decimal.NewFromInt(coinAmount)
		changeDe     decimal.Decimal = decimal.New(1, coinPriceDecimalNum)
	)
	price = coinAmountDe.Mul(coinPriceDe).Div(changeDe).IntPart()
	if price == 0 {
		return 0, errors.New("The quantity is too small and the total price is less than 1sats. Please increase the quantity. ")
	}
	return price, nil
}

func MakePrice(coinAmount, amount int64) (int64, int32) {
	var (
		coinPrice           int64           = 0
		coinPriceDecimalNum int32           = coinPriceDecimalNumDefault
		coinAmountDe        decimal.Decimal = decimal.NewFromInt(coinAmount)
		amountDe            decimal.Decimal = decimal.NewFromInt(amount)
		changeDe            decimal.Decimal = decimal.New(1, coinPriceDecimalNum)
	)
	coinPrice = amountDe.Div(coinAmountDe).Mul(changeDe).IntPart()
	return coinPrice, coinPriceDecimalNum
}

func UpdateMarketPrice(net, tick, pair string) *model.Brc20TickModel {
	var (
		askList     []*model.OrderBrc20Model
		bidList     []*model.OrderBrc20Model
		marketPrice uint64 = 0
		totalPrice  uint64 = 0
		total       uint64 = 0
		tickInfo    *model.Brc20TickModel
		sellPrice   uint64 = 0
		sellTotal   uint64 = 0
		buyPrice    uint64 = 0
		buyTotal    uint64 = 0
		version     int    = 0

		lastPrice uint64 = 0

		askLastFinish *model.OrderBrc20Model
		bidLastFinish *model.OrderBrc20Model
	)
	askList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "", model.OrderTypeSell, model.OrderStateCreate, 10, 0, 0,
		"coinRatePrice", 1, 0, 0)
	bidList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "", model.OrderTypeBuy, model.OrderStateCreate, 10, 0, 0,
		"coinRatePrice", -1, 0, 0)
	for _, v := range askList {
		if v.CoinRatePrice == 0 {
			continue
		}
		sellPrice = sellPrice + v.CoinRatePrice
		totalPrice = totalPrice + v.CoinRatePrice
		total++
		sellTotal++
	}
	if sellTotal != 0 {
		sellPrice = sellPrice / sellTotal
	}

	for _, v := range bidList {
		if v.CoinRatePrice == 0 {
			continue
		}
		buyPrice = buyPrice + v.CoinRatePrice
		totalPrice = totalPrice + v.CoinRatePrice
		total++
		buyTotal++
	}
	if buyTotal != 0 {
		buyPrice = buyPrice / buyTotal
	}
	if total != 0 {
		marketPrice = totalPrice / total
	}

	askLastFinish, _ = mongo_service.FindLastOrderBrc20ModelFinish(net, tick, model.OrderTypeSell, model.OrderStateFinish)
	bidLastFinish, _ = mongo_service.FindLastOrderBrc20ModelFinish(net, tick, model.OrderTypeBuy, model.OrderStateFinish)
	if askLastFinish != nil && bidLastFinish != nil {
		if askLastFinish.DealTime > bidLastFinish.DealTime {
			lastPrice = askLastFinish.CoinRatePrice
		} else {
			lastPrice = bidLastFinish.CoinRatePrice
		}
	} else if askLastFinish != nil && bidLastFinish == nil {
		lastPrice = askLastFinish.CoinRatePrice
	} else if askLastFinish == nil && bidLastFinish != nil {
		lastPrice = bidLastFinish.CoinRatePrice
	}

	tickInfo, _ = mongo_service.FindBrc20TickModelByPair(net, pair, version)
	if tickInfo == nil {
		tickInfo = &model.Brc20TickModel{
			Net:  net,
			Tick: tick,
			Pair: pair,
		}
	}
	tickInfo.Buy = buyPrice
	tickInfo.Sell = sellPrice
	tickInfo.AvgPrice = marketPrice
	tickInfo.Last = lastPrice
	tickInfo.Version = version

	_, err := mongo_service.SetBrc20TickModel(tickInfo)
	if err != nil {
		return nil
	}
	ws.SendTickInfo(ws.NewWsNotifyTick(tickInfo))
	return tickInfo
}

func GetMarketPrice(net, tick, pair string) uint64 {
	fmt.Printf("net:%s, tick:%s, pair:%s\n", net, tick, pair)

	tickInfo, _ := mongo_service.FindBrc20TickModelByPair(net, pair, 0)
	if tickInfo == nil {
		tickInfo = UpdateMarketPrice(net, tick, pair)
	}

	if tickInfo == nil {
		otherPriceInfo := getOtherMarketPrice(tick)
		if otherPriceInfo != nil && otherPriceInfo.UpdateTime != 0 {
			price, _ := strconv.ParseUint(otherPriceInfo.LastPrice, 10, 64)
			return price
		}
	} else {
		marketPrice := uint64(0)
		if tickInfo.Sell != 0 && tickInfo.Buy != 0 {
			marketPrice = (tickInfo.Sell + tickInfo.Buy) / 2
		} else if tickInfo.Sell != 0 && tickInfo.Buy == 0 {
			marketPrice = tickInfo.Sell
		} else if tickInfo.Sell == 0 && tickInfo.Buy != 0 {
			marketPrice = tickInfo.Buy
		} else {
			otherPriceInfo := getOtherMarketPrice(tick)
			if otherPriceInfo != nil && otherPriceInfo.UpdateTime != 0 {
				price, _ := strconv.ParseUint(otherPriceInfo.LastPrice, 10, 64)
				return price
			}
		}
		return marketPrice
	}

	//guideEntity, _ := mongo_service.FindOrderBrc20MarketPriceModelByPair(net, pair)
	//if tickInfo == nil {
	//	if guideEntity == nil {
	//		return 0
	//	}
	//	return uint64(guideEntity.Price)
	//} else {
	//	//if guideEntity != nil && guideEntity.Price > int64(tickInfo.AvgPrice) {
	//	//	return uint64(guideEntity.Price)
	//	//}
	//}

	return tickInfo.AvgPrice
}

func getOtherMarketPrice(tick string) *common_service.PriceInfo {
	var (
		marketInfo *common_service.PriceInfo
		nowTime    int64 = tool.MakeTimestamp()
		ok         bool
	)
	marketInfo, ok = common_service.Brc20TickMarketDataMap[tick]
	if ok && nowTime-marketInfo.UpdateTime <= 1000*60*10 {
		return marketInfo
	}
	marketInfo = &common_service.PriceInfo{}

	btcPriceInfo, _ := oklink_service.GetBrc20TickMarketData("")
	if btcPriceInfo == nil || len(btcPriceInfo) == 0 {
		return nil
	}

	inscriptionId, ok := common_service.Brc20TickInscriptionMap[tick]
	if !ok {
		return nil
	}
	tickPriceInfo, _ := oklink_service.GetBrc20TickMarketData(inscriptionId)
	if tickPriceInfo == nil || len(tickPriceInfo) == 0 {
		return nil
	}
	btcDe, _ := decimal.NewFromString(btcPriceInfo[0].LastPrice)
	tickDe, _ := decimal.NewFromString(tickPriceInfo[0].LastPrice)
	//priceDe := tickDe.Div(btcDe.Div(decimal.New(1, 8))).Mul(decimal.New(1, coinPriceDecimalNumDefault))
	priceDe := tickDe.Div(btcDe.Div(decimal.New(1, 8)))
	marketInfo.LastPrice = priceDe.StringFixed(0)
	if marketInfo.LastPrice == "0" {
		marketInfo.VisionPrice = priceDe.StringFixed(8)
	} else {
		marketInfo.VisionPrice = priceDe.StringFixed(0)
	}

	fmt.Printf("[other-price] %s, %s\n", tick, marketInfo.LastPrice)
	//high
	//low
	marketInfo.UpdateTime = nowTime
	common_service.Brc20TickMarketDataMap[tick] = marketInfo
	return marketInfo
}

func UpdateMarketPriceV2(net, tick, pair string) *model.Brc20TickModel {
	var (
		lastList     []*model.OrderBrc20Model = make([]*model.OrderBrc20Model, 0)
		askList      []*model.OrderBrc20Model
		bidList      []*model.OrderBrc20Model
		marketPrice  uint64 = 0
		totalPrice   uint64 = 0
		total        uint64 = 0
		tickInfo     *model.Brc20TickModel
		sellPrice    uint64 = 0
		sellTotal    uint64 = 0
		buyPrice     uint64 = 0
		buyTotal     uint64 = 0
		lastTotal    uint64 = 0
		lastTopPrice uint64 = 0
		lastAllPrice uint64 = 0

		lastPrice uint64 = 0
		version   int    = 2

		askLastFinishList []*model.OrderBrc20Model
		bidLastFinishList []*model.OrderBrc20Model
	)
	askList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "", model.OrderTypeSell, model.OrderStateCreate, 10, 0, 0,
		"coinPrice", 1, 0, 0)
	bidList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "", model.OrderTypeBuy, model.OrderStateCreate, 10, 0, 0,
		"coinPrice", -1, 0, 0)
	for _, v := range askList {
		if v.CoinPrice == 0 {
			continue
		}
		sellPrice = sellPrice + uint64(v.CoinPrice)
		totalPrice = totalPrice + uint64(v.CoinPrice)
		total++
		sellTotal++
	}
	if sellTotal != 0 {
		sellPrice = sellPrice / sellTotal
	}

	for _, v := range bidList {
		if v.CoinPrice == 0 {
			continue
		}
		buyPrice = buyPrice + uint64(v.CoinPrice)
		totalPrice = totalPrice + uint64(v.CoinPrice)
		total++
		buyTotal++
	}
	if buyTotal != 0 {
		buyPrice = buyPrice / buyTotal
	}
	if total != 0 {
		marketPrice = totalPrice / total
	}

	askLastFinishList, _ = mongo_service.FindLastOrderBrc20ModelFinishList(net, tick, 10, model.OrderTypeSell, model.OrderStateFinish)
	bidLastFinishList, _ = mongo_service.FindLastOrderBrc20ModelFinishList(net, tick, 10, model.OrderTypeBuy, model.OrderStateFinish)
	if askLastFinishList != nil && len(askLastFinishList) != 0 {
		lastList = append(lastList, askLastFinishList...)
	}
	if bidLastFinishList != nil && len(bidLastFinishList) != 0 {
		lastList = append(lastList, bidLastFinishList...)
	}

	lastTime := int64(0)
	for _, v := range lastList {
		if lastTime == 0 || v.DealTime > lastTime {
			lastPrice = uint64(v.CoinPrice)
		}
		lastAllPrice = lastAllPrice + uint64(v.CoinPrice)
		lastTotal++
	}
	if lastTotal > 0 {
		lastTopPrice = lastAllPrice / lastTotal
	}

	tickInfo, _ = mongo_service.FindBrc20TickModelByPair(net, pair, version)
	if tickInfo == nil {
		tickInfo = &model.Brc20TickModel{
			Net:  net,
			Tick: tick,
			Pair: pair,
		}
	}
	tickInfo.Buy = buyPrice
	tickInfo.Sell = sellPrice
	tickInfo.AvgPrice = marketPrice
	tickInfo.Last = lastPrice
	tickInfo.LastTop = lastTopPrice
	tickInfo.LastTotal = lastTotal
	tickInfo.Version = version

	_, err := mongo_service.SetBrc20TickModel(tickInfo)
	if err != nil {
		return nil
	}

	//ws.SendTickInfo(ws.NewWsNotifyTick(tickInfo))
	return tickInfo
}

func GetMarketPriceV2(net, tick, pair string) uint64 {
	fmt.Printf("[V2]net:%s, tick:%s, pair:%s\n", net, tick, pair)
	tickInfo, _ := mongo_service.FindBrc20TickModelByPair(net, pair, 2)
	if tickInfo == nil {
		fmt.Printf("[V2]net:%s, tick:%s, pair:%s update\n", net, tick, pair)

		tickInfo = UpdateMarketPriceV2(net, tick, pair)
	}
	guideEntity, _ := mongo_service.FindOrderBrc20MarketPriceModelByPair(net, pair)
	if tickInfo == nil {
		if guideEntity == nil {
			return 0
		}
		return uint64(guideEntity.Price)
	} else {
		if guideEntity != nil && guideEntity.Price > int64(tickInfo.AvgPrice) {
			return uint64(guideEntity.Price)
		}
	}

	marketPrice := tickInfo.LastTop
	if tickInfo.LastTotal < 5 {
		priceInfo := getOtherMarketPrice(tick)
		visionPriceDe, _ := decimal.NewFromString(priceInfo.VisionPrice)
		marketPrice = uint64(visionPriceDe.Mul(decimal.New(1, 8)).IntPart())
	}

	return marketPrice
}

func GetNetParams(net string) *chaincfg.Params {
	var (
		netParams *chaincfg.Params = &chaincfg.MainNetParams
	)
	switch strings.ToLower(net) {
	case "mainnet", "livenet":
		netParams = &chaincfg.MainNetParams
		break
	case "signet":
		netParams = &chaincfg.SigNetParams
		break
	case "testnet":
		netParams = &chaincfg.TestNet3Params
		break
	}
	return netParams
}

func GetTxHash(rawTxByte []byte) string {
	txHash := tool.DoubleSHA256(rawTxByte)
	for i := 0; i < len(txHash)/2; i++ {
		h := txHash[len(txHash)-1-i]
		txHash[len(txHash)-1-i] = txHash[i]
		txHash[i] = h
	}
	return hex.EncodeToString(txHash)
}

func GetTestFakerInscription(net string) []*model.OrderUtxoModel {
	utxoMockInscriptionList, _ := mongo_service.FindUtxoList(net, -1, 1000, 0, model.UtxoTypeFakerInscription, -1)
	return utxoMockInscriptionList
}

func SaveForUserBidDummy(net, tick, address, orderId, dummyPreTxId string, dummyPreIndex int64, state model.DummyState) {
	dummyEntity := &model.OrderBrc20BidDummyModel{
		Net:        net,
		DummyId:    fmt.Sprintf("%s:%d", dummyPreTxId, dummyPreIndex),
		OrderId:    orderId,
		Tick:       tick,
		Address:    address,
		DummyState: state,
		Timestamp:  tool.MakeTimestamp(),
	}
	mongo_service.SetOrderBrc20BidDummyModel(dummyEntity)
}

func SaveForUserBidUtxo(net, tick, address, orderId, utxoPreTxId string, utxoPreIndex int64, state model.DummyState) {
	dummyEntity := &model.OrderBrc20BidDummyModel{
		Net:        net,
		DummyId:    fmt.Sprintf("%s:%d", utxoPreTxId, utxoPreIndex),
		OrderId:    orderId,
		Tick:       tick,
		Address:    address,
		DummyState: state,
		Timestamp:  tool.MakeTimestamp(),
	}
	mongo_service.SetOrderBrc20BidDummyModel(dummyEntity)
}

func UpdateForOrderBidDummy(orderId string, state model.DummyState) {
	dummyList, _ := mongo_service.FindOrderBrc20BidDummyModelList(orderId, "", model.DummyStateLive, 0, 10)
	for _, v := range dummyList {
		v.DummyState = state
		mongo_service.SetOrderBrc20BidDummyModel(v)
	}
}

func GetPlatformKeyAndAddressSendBrc20(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeySendBrc20, config.PlatformTestnetAddressSendBrc20
	}
	return config.PlatformMainnetPrivateKeySendBrc20, config.PlatformMainnetAddressSendBrc20
}

func GetPlatformKeyAndAddressSendBrc20ForAsk(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeySendBrc20ForAsk, config.PlatformTestnetAddressSendBrc20ForAsk
	}
	return config.PlatformMainnetPrivateKeySendBrc20ForAsk, config.PlatformMainnetAddressSendBrc20ForAsk
}

func GetPlatformKeyAndAddressReceiveValueForAsk(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyReceiveValueForAsk, config.PlatformTestnetAddressReceiveValueForAsk
	}
	return config.PlatformMainnetPrivateKeyReceiveValueForAsk, config.PlatformMainnetAddressReceiveValueForAsk
}

func GetPlatformKeyAndAddressReceiveBrc20(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyReceiveBrc20, config.PlatformTestnetAddressReceiveBrc20
	}
	return config.PlatformMainnetPrivateKeyReceiveBrc20, config.PlatformMainnetAddressReceiveBrc20
}

func GetPlatformKeyAndAddressReceiveBidValue(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyReceiveBidValue, config.PlatformTestnetAddressReceiveBidValue
	}
	return config.PlatformMainnetPrivateKeyReceiveBidValue, config.PlatformMainnetAddressReceiveBidValue
}

func GetPlatformKeyAndAddressReceiveBidValueToX(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyReceiveBidValueToX, config.PlatformTestnetAddressReceiveBidValueToX
	}
	return config.PlatformMainnetPrivateKeyReceiveBidValueToX, config.PlatformMainnetAddressReceiveBidValueToX
}

func GetPlatformKeyAndAddressReceiveBidValueToReturn(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyReceiveBidValueToReturn, config.PlatformTestnetAddressReceiveBidValueToReturn
	}
	return config.PlatformMainnetPrivateKeyReceiveBidValueToReturn, config.PlatformMainnetAddressReceiveBidValueToReturn
}

func GetPlatformKeyAndAddressReceiveDummyValue(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyReceiveDummyValue, config.PlatformTestnetAddressReceiveDummyValue
	}
	return config.PlatformMainnetPrivateKeyReceiveDummyValue, config.PlatformMainnetAddressReceiveDummyValue
}

func GetPlatformKeyAndAddressReceiveFee(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyReceiveFee, config.PlatformTestnetAddressReceiveFee
	}
	return config.PlatformMainnetPrivateKeyReceiveFee, config.PlatformMainnetAddressReceiveFee
}

func GetPlatformKeyAndAddressReceiveValueForPoolBtc(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyReceiveValueForPoolBtc, config.PlatformTestnetAddressReceiveValueForPoolBtc
	}
	return config.PlatformMainnetPrivateKeyReceiveValueForPoolBtc, config.PlatformMainnetAddressReceiveValueForPoolBtc
}

func GetPlatformKeyMultiSig(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyMultiSig, config.PlatformTestnetPublicKeyMultiSig
	}
	return config.PlatformMainnetPrivateKeyMultiSig, config.PlatformMainnetPublicKeyMultiSig
}

func GetPlatformKeyMultiSigForBtc(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyMultiSigBtc, config.PlatformTestnetPublicKeyMultiSigBtc
	}
	return config.PlatformMainnetPrivateKeyMultiSigBtc, config.PlatformMainnetPublicKeyMultiSigBtc
}

func GetPlatformKeyAndAddressForMultiSigInscription(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyInscriptionMultiSig, config.PlatformTestnetAddressInscriptionMultiSig
	}
	return config.PlatformMainnetPrivateKeyInscriptionMultiSig, config.PlatformMainnetAddressInscriptionMultiSig
}

func GetPlatformKeyAndAddressForMultiSigInscriptionAndReceiveValue(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyInscriptionMultiSigForReceiveValue, config.PlatformTestnetAddressInscriptionMultiSigForReceiveValue
	}
	return config.PlatformMainnetPrivateKeyInscriptionMultiSigForReceiveValue, config.PlatformMainnetAddressInscriptionMultiSigForReceiveValue
}

func GetPlatformKeyAndAddressForRewardBrc20(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyRewardBrc20, config.PlatformTestnetAddressRewardBrc20
	}
	return config.PlatformMainnetPrivateKeyRewardBrc20, config.PlatformMainnetAddressRewardBrc20
}

func GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyRewardBrc20FeeUtxos, config.PlatformTestnetAddressRewardBrc20FeeUtxos
	}
	return config.PlatformMainnetPrivateKeyRewardBrc20FeeUtxos, config.PlatformMainnetAddressRewardBrc20FeeUtxos
}

func GetPlatformKeyAndAddressForRepurchaseReceiveBrc20(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyRepurchaseReceiveBrc20, config.PlatformTestnetAddressRepurchaseReceiveBrc20
	}
	return config.PlatformMainnetPrivateKeyRepurchaseReceiveBrc20, config.PlatformMainnetAddressRepurchaseReceiveBrc20
}

func GetPlatformKeyAndAddressForDummy(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyDummy, config.PlatformTestnetAddressDummy
	}
	return config.PlatformMainnetPrivateKeyDummy, config.PlatformMainnetAddressDummy
}

func GetPlatformKeyAndAddressForLp(net string) (string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.PlatformTestnetPrivateKeyLp, config.PlatformTestnetAddressLp
	}
	return config.PlatformMainnetPrivateKeyLp, config.PlatformMainnetAddressLp
}

func CheckBidInscriptionIdExist(inscriptionId string) bool {
	entity, _ := mongo_service.FindOrderBrc20ModelByInscriptionId(inscriptionId, model.OrderStateCreate)
	if entity == nil || entity.Id == 0 {
		return false
	}
	return true
}

func SetUsedDummyUtxo(utxoDummyList []*model.OrderUtxoModel, useTx string) {
	if utxoDummyList == nil || len(utxoDummyList) == 0 {
		return
	}
	for _, v := range utxoDummyList {
		v.UseTx = useTx
		v.UsedState = model.UsedYes
		err := mongo_service.UpdateOrderUtxoModelForUsed(v.UtxoId, useTx, v.UsedState)
		if err != nil {
			continue
		}
	}
}

func SetOccupiedDummyUtxo(utxoDummyList []*model.OrderUtxoModel, orderId string) {
	if utxoDummyList == nil || len(utxoDummyList) == 0 {
		return
	}
	for _, v := range utxoDummyList {
		v.OrderId = orderId
		v.UsedState = model.UsedOccupied
		err := mongo_service.UpdateOrderUtxoModelForOccupied(v.UtxoId, orderId, v.UsedState)
		if err != nil {
			continue
		}
	}
}

// release occupied dummy utxo
func ReleaseOccupiedDummyUtxo(utxoDummyList []*model.OrderUtxoModel) {
	if utxoDummyList == nil || len(utxoDummyList) == 0 {
		return
	}
	for _, v := range utxoDummyList {
		v.OrderId = ""
		v.UsedState = model.UsedNo
		err := mongo_service.UpdateOrderUtxoModelForOccupied(v.UtxoId, "", v.UsedState)
		if err != nil {
			continue
		}
	}
}

func setUsedBidYUtxo(utxoBidYList []*model.OrderUtxoModel, useTx string) {
	if utxoBidYList == nil || len(utxoBidYList) == 0 {
		return
	}
	for _, v := range utxoBidYList {
		v.UseTx = useTx
		v.UsedState = model.UsedYes
		err := mongo_service.UpdateOrderUtxoModelForUsed(v.UtxoId, useTx, v.UsedState)
		if err != nil {
			continue
		}
	}
}

func setUsedMultiSigInscriptionUtxo(utxoMultiSigInscriptionList []*model.OrderUtxoModel, useTx string) {
	for _, v := range utxoMultiSigInscriptionList {
		v.UseTx = useTx
		v.UsedState = model.UsedYes
		err := mongo_service.UpdateOrderUtxoModelForUsed(v.UtxoId, useTx, v.UsedState)
		if err != nil {
			continue
		}
	}
}

func SetUsedRewardUtxo(utxoRewardInscriptionList []*model.OrderUtxoModel, useTx string) {
	for _, v := range utxoRewardInscriptionList {
		v.UseTx = useTx
		v.UsedState = model.UsedYes
		err := mongo_service.UpdateOrderUtxoModelForUsed(v.UtxoId, useTx, v.UsedState)
		if err != nil {
			continue
		}
	}
}

func setUsedFakerInscriptionUtxo(utxoId, useTx string, useState model.UsedState) {
	err := mongo_service.UpdateOrderUtxoModelForUsed(utxoId, useTx, useState)
	if err != nil {
		return
	}
}

func CheckPublicKeyAddress(netParams *chaincfg.Params, publicKeyStr, checkAddress string) (bool, error) {
	if publicKeyStr == "" {
		return true, nil
	}
	publicKeyByte, err := hex.DecodeString(publicKeyStr)
	if err != nil {
		return false, err
	}

	publicKey, err := btcec.ParsePubKey(publicKeyByte)
	if err != nil {
		return false, err
	}

	legacyAddress, err := btcutil.NewAddressPubKey(publicKeyByte, netParams)
	if err != nil {
		return false, err
	}
	if legacyAddress.EncodeAddress() == checkAddress {
		return true, nil
	}

	nativeSegwitAddress, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(publicKey.SerializeCompressed()), netParams)
	if err != nil {
		return false, err
	}
	if nativeSegwitAddress.EncodeAddress() == checkAddress {
		return true, nil
	}

	taprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(publicKey)), netParams)
	if err != nil {
		return false, err
	}
	if taprootAddress.EncodeAddress() == checkAddress {
		return true, nil
	}
	return false, nil
}

func UpdateTickRecentlyInfo(net, tick string) {
	var (
		limit                   int64 = 5000
		entityList              []*model.OrderBrc20Model
		lastFinish              *model.OrderBrc20Model
		lastFinish24Ago         *model.OrderBrc20Model
		highest, lowest, volume uint64 = 0, 0, 0
		err                     error
		startTime, endTime      int64  = 0, 0
		orderLastTime           int64  = 0
		percentage              string = ""
		nowPrice, lastPrice     int64  = 0, 0
		dis                     int64  = 1000 * 60 * 60 * 24
		entity                  *model.Brc20TickRecentlyInfoModel
		tickId                  string = fmt.Sprintf("%s_%s_%s", net, tick, model.RecentlyType24h)
	)
	_ = lastFinish24Ago
	_ = percentage
	_ = nowPrice
	_ = lastPrice
	endTime = tool.MakeTimestamp()
	startTime = endTime - dis

	entityList, _ = mongo_service.FindOrderBrc20ModelListByDealTimestamp(net, tick, 0, model.OrderStateFinish,
		limit, startTime, endTime)
	if entityList == nil || len(entityList) == 0 {
		lastFinish, _ = mongo_service.FindLastOrderBrc20ModelFinish(net, tick, 0, model.OrderStateFinish)
		if lastFinish == nil {
			return
		}
		startTime, endTime = lastFinish.DealTime-dis, lastFinish.DealTime
		entityList, _ = mongo_service.FindOrderBrc20ModelListByDealTimestamp(net, tick, 0, model.OrderStateFinish,
			limit, startTime, endTime)
		if entityList == nil || len(entityList) == 0 {
			return
		}
	}
	volume = uint64(len(entityList))
	for _, v := range entityList {
		if orderLastTime == 0 || orderLastTime < v.DealTime {
			orderLastTime = v.DealTime
			nowPrice = int64(v.CoinRatePrice)
		}
		if highest == 0 || v.CoinRatePrice > highest {
			highest = v.CoinRatePrice
		}
		if lowest == 0 || lowest > v.CoinRatePrice {
			lowest = v.CoinRatePrice
		}
	}

	entity, _ = mongo_service.FindBrc20TickRecentlyInfoModelByTickId(tickId)
	if entity == nil {
		entity = &model.Brc20TickRecentlyInfoModel{
			TickId:        tickId,
			Net:           net,
			Tick:          tick,
			RecentlyType:  model.RecentlyType24h,
			OrderLastTime: orderLastTime,
			Timestamp:     tool.MakeTimestamp(),
		}
	}
	if entity.Highest != fmt.Sprintf("%d", highest) {
		entity.Highest = fmt.Sprintf("%d", highest)
	}
	if entity.Lowest != fmt.Sprintf("%d", lowest) {
		entity.Lowest = fmt.Sprintf("%d", lowest)
	}
	if entity.Volume != int64(volume) {
		entity.Volume = int64(volume)
	}

	_, err = mongo_service.SetBrc20TickRecentlyInfoModel(entity)
	if err != nil {
		major.Println(fmt.Sprintf("SetBrc20TickRecentlyInfoModel err:%s", err))
	}
}

// address to pkScript
func AddressToPkScript(net, address string) (string, error) {
	netParams := GetNetParams(net)
	addr, err := btcutil.DecodeAddress(address, netParams)
	if err != nil {
		return "", err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}
	pkScript := hex.EncodeToString(pkScriptByte)
	return pkScript, nil
}

func GetTxDetail(net, txId string) (*oklink_service.TxDetail, error) {
	var (
		rawTx         string = ""
		err           error
		tx            *wire.MsgTx
		txDetail      *oklink_service.TxDetail
		outputDetails []*oklink_service.OutputItem = make([]*oklink_service.OutputItem, 0)
	)
	rawTx, _, err = mempool_space_service.GetTxHex(net, txId)
	if err != nil {
		return nil, err
	}
	txRawByte, _ := hex.DecodeString(rawTx)
	tx = wire.NewMsgTx(2)
	err = tx.Deserialize(bytes.NewReader(txRawByte))
	if err != nil {
		return nil, errors.New("ParseTx err")
	}
	for _, out := range tx.TxOut {
		valueDe := decimal.NewFromInt(out.Value)
		valueDe = valueDe.Div(decimal.NewFromInt(100000000))

		_, addrs, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, &chaincfg.MainNetParams)
		if err != nil {
			return nil, errors.New("Extract address from out for parse. ")
		}
		outputHash := addrs[0].EncodeAddress()
		outputDetails = append(outputDetails, &oklink_service.OutputItem{
			Amount:     valueDe.StringFixed(8),
			OutputHash: outputHash,
		})
	}

	txDetail = &oklink_service.TxDetail{
		TxId:          txId,
		Height:        "",
		OutputDetails: outputDetails,
	}
	return txDetail, nil
}

// check the inscription of order if exist
func CheckInscriptionExist(address, inscriptionId string) bool {
	liveUtxoList := make([]*oklink_service.UtxoItem, 0)
	//utxoResp, err := oklink_service.GetAddressUtxo(address, 1, 100)
	//if err != nil {
	//	return true
	//}
	//if utxoResp.UtxoList != nil && len(utxoResp.UtxoList) != 0 {
	//	liveUtxoList = append(liveUtxoList, utxoResp.UtxoList...)
	//}

	utxoList, err := unisat_service.GetAddressUtxo(address)
	if err != nil {
		fmt.Printf("[Check]GetAddressUtxo err:%s\n", err)
		return true
	}
	if utxoList != nil && len(utxoList) != 0 {
		for _, uu := range utxoList {
			liveUtxoList = append(liveUtxoList, &oklink_service.UtxoItem{
				TxId:          uu.TxId,
				Index:         strconv.FormatInt(uu.OutputIndex, 10),
				Height:        "",
				BlockTime:     "",
				Address:       uu.ScriptPk,
				UnspentAmount: strconv.FormatInt(uu.Satoshis, 10),
			})
		}
	}

	utxoInscription, err := unisat_service.GetAddressInscriptions(address)
	if err != nil {
		fmt.Printf("[Check]GetAddressInscriptions err:%s\n", err)
		return true
	}
	if utxoInscription != nil && len(utxoInscription) != 0 {
		for _, ui := range utxoInscription {
			output := ui.Output
			outputStrs := strings.Split(output, ":")
			if len(outputStrs) <= 2 {
				continue
			}
			liveUtxoList = append(liveUtxoList, &oklink_service.UtxoItem{
				TxId:          outputStrs[0],
				Index:         outputStrs[1],
				Height:        "",
				BlockTime:     "",
				Address:       ui.Address,
				UnspentAmount: strconv.FormatInt(ui.OutputValue, 10),
			})
		}
	}
	has := false
	for _, u := range liveUtxoList {
		uId := fmt.Sprintf("%si%s", u.TxId, u.Index)
		//uId := fmt.Sprintf("%s_%d", u.TxId, u.OutputIndex)
		fmt.Printf("liveUtxo:[%s]\n", uId)
		if inscriptionId == uId {
			has = true
			break
		}
	}
	return has
}

func GetTxBlock(txId string) int64 {
	var (
		blockHeight int64 = 0
	)
	//nodeTx, _ := node.GetTx("livenet", txId)
	//if nodeTx != nil {
	//	fmt.Printf("[RPC][%s]-%d\n", nodeTx.TxID, nodeTx.BlockHeight)
	//	blockHeight = int64(nodeTx.BlockHeight)
	//	return blockHeight
	//}

	tx, err := oklink_service.GetTxDetail(txId)
	if err != nil {
		return 0
	}
	blockHeight, _ = strconv.ParseInt(tx.Height, 10, 64)
	return blockHeight
}

func GetTxConfirm(txId string) int64 {
	var (
		blockHeight int64 = 0
	)
	nodeTx, _ := node.GetTx("livenet", txId)
	if nodeTx != nil {
		fmt.Printf("[RPC][%s]-%d\n", nodeTx.TxID, nodeTx.Confirmations)
		blockHeight = int64(nodeTx.Confirmations)
		return blockHeight
	}

	tx, err := oklink_service.GetTxDetail(txId)
	if err != nil {
		return 0
	}
	blockHeight, _ = strconv.ParseInt(tx.Height, 10, 64)
	return blockHeight
}
