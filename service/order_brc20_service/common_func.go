package order_brc20_service

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"ordbook-aggregation/config"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
	"ordbook-aggregation/ws_service/ws"
	"strings"
)

func UpdateMarketPrice(net, tick, pair string) *model.Brc20TickModel{
	var(
		askList []*model.OrderBrc20Model
		bidList []*model.OrderBrc20Model
		marketPrice uint64 = 0
		totalPrice uint64 = 0
		total uint64 = 0
		tickInfo *model.Brc20TickModel
		sellPrice uint64 = 0
		sellTotal uint64 = 0
		buyPrice uint64 = 0
		buyTotal uint64 = 0
	)
	askList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "", model.OrderTypeSell, model.OrderStateCreate, 10, 0, 0,
		"coinRatePrice", 1)
	bidList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "", model.OrderTypeBuy, model.OrderStateCreate, 10, 0, 0,
		"coinRatePrice", -1)
	for _, v := range askList{
		if v.CoinRatePrice == 0 {
			continue
		}
		sellPrice = sellPrice + v.CoinRatePrice
		totalPrice = totalPrice + v.CoinRatePrice
		total++
		sellTotal++
	}
	if sellTotal != 0 {
		sellPrice = sellPrice/sellTotal
	}

	for _, v := range bidList{
		if v.CoinRatePrice == 0 {
			continue
		}
		buyPrice = buyPrice + v.CoinRatePrice
		totalPrice = totalPrice + v.CoinRatePrice
		total++
		buyTotal++
	}
	if buyTotal != 0 {
		buyPrice = buyPrice/buyTotal
	}
	if total != 0 {
		marketPrice = totalPrice/total
	}


	tickInfo, _ = mongo_service.FindBrc20TickModelByPair(net, pair)
	if tickInfo == nil {
		tickInfo = &model.Brc20TickModel{
			Net:                net,
			Tick:               tick,
			Pair:               pair,
			Buy:                buyPrice,
			Sell:               sellPrice,
			AvgPrice:           marketPrice,
		}
	}
	tickInfo.Buy = buyPrice
	tickInfo.Sell = sellPrice
	tickInfo.AvgPrice = marketPrice
	_, err := mongo_service.SetBrc20TickModel(tickInfo)
	if err != nil {
		return nil
	}
	ws.SendTickInfo(ws.NewWsNotifyTick(tickInfo))
	return tickInfo
}

func GetMarketPrice(net, tick, pair string) uint64 {
	tickInfo, _ := mongo_service.FindBrc20TickModelByPair(net, pair)
	if tickInfo == nil {
		tickInfo = UpdateMarketPrice(net, tick, pair)
	}
	guideEntity, _ := mongo_service.FindOrderBrc20MarketPriceModelByPair(net, pair)
	if tickInfo == nil {
		if guideEntity == nil {
			return 0
		}
		return uint64(guideEntity.Price)
	}else {
		if guideEntity != nil && guideEntity.Price > int64(tickInfo.AvgPrice) {
			return uint64(guideEntity.Price)
		}
	}
	return tickInfo.AvgPrice
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
	//翻转
	for i := 0; i < len(txHash)/2; i++ {
		h := txHash[len(txHash)-1-i]
		txHash[len(txHash)-1-i] = txHash[i]
		txHash[i] = h
	}
	return hex.EncodeToString(txHash)
}

func GetTestFakerInscription(net string) []*model.OrderUtxoModel {
	utxoMockInscriptionList, _ := mongo_service.FindUtxoList(net, -1, 1000, model.UtxoTypeFakerInscription)
	return utxoMockInscriptionList
}

func SaveForUserBidDummy(net, tick, address, orderId, dummyPreTxId string, dummyPreIndex int64, state model.DummyState)  {
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

func UpdateForOrderBidDummy(orderId string, state model.DummyState)  {
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

func CheckBidInscriptionIdExist(inscriptionId string) bool {
	entity, _ := mongo_service.FindOrderBrc20ModelByInscriptionId(inscriptionId, model.OrderStateCreate)
	if entity == nil || entity.Id == 0 {
		return false
	}
	return true
}

func setUsedDummyUtxo(utxoDummyList []*model.OrderUtxoModel, useTx string)  {
	for _, v := range utxoDummyList {
		v.UseTx = useTx
		v.UsedState = model.UsedYes
		err := mongo_service.UpdateOrderUtxoModelForUsed(v.UtxoId, useTx, v.UsedState)
		if err != nil {
			continue
		}
	}
}

func setUsedBidYUtxo(utxoBidYList []*model.OrderUtxoModel, useTx string)  {
	for _, v := range utxoBidYList {
		v.UseTx = useTx
		v.UsedState = model.UsedYes
		err := mongo_service.UpdateOrderUtxoModelForUsed(v.UtxoId, useTx, v.UsedState)
		if err != nil {
			continue
		}
	}
}

func serUsedFakerInscriptionUtxo(utxoId, useTx string, useState model.UsedState)  {
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