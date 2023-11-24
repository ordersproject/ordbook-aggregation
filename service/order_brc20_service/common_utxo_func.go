package order_brc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
	"ordbook-aggregation/model"
	"ordbook-aggregation/redis"
	"ordbook-aggregation/service/cache_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
	"sync"
)

const (
	maxLimit int64 = 500

	DoBidUtxoPerAmount1w   int64 = 10000
	DoBidUtxoPerAmount5w   int64 = 50000
	DoBidUtxoPerAmount10w  int64 = 100000
	DoBidUtxoPerAmount50w  int64 = 500000
	DoBidUtxoPerAmount100w int64 = 1000000
)

var (
	unoccupiedUtxoLock *sync.RWMutex = new(sync.RWMutex)
	saveUtxoLock       *sync.RWMutex = new(sync.RWMutex)
)

func GetUnoccupiedUtxoList(net string, limit, totalNeedAmount int64, utxoType model.UtxoType, fromOrderId string, networkFeeRate int64) ([]*model.OrderUtxoModel, error) {
	var (
		cacheType          string                  = cache_service.CacheLockUtxoTypeDummy
		redisKeyPrefix     string                  = ""
		sortIndexList      []int                   = make([]int, 0)
		utxoIdKeyList      []string                = make([]string, 0)
		startIndex         int64                   = -1
		utxoList           []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		unoccupiedUtxoList []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		perAmount          int64                   = 0
	)
	switch utxoType {
	case model.UtxoTypeDummy:
		cacheType = cache_service.CacheLockUtxoTypeDummy
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeDummy_)
		break
	case model.UtxoTypeDummy1200:
		cacheType = cache_service.CacheLockUtxoTypeDummy1200
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeDummy1200_)
		break
	case model.UtxoTypeDummyAsk:
		cacheType = cache_service.CacheLockUtxoTypeDummyAsk
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeDummyAsk_)
		break
	case model.UtxoTypeDummy1200Ask:
		cacheType = cache_service.CacheLockUtxoTypeDummy1200Ask
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeDummy1200Ask_)
		break
	case model.UtxoTypeDummyBidX:
		cacheType = cache_service.CacheLockUtxoTypeDummyBidX
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeDummyBidX_)
		break
	case model.UtxoTypeDummy1200BidX:
		cacheType = cache_service.CacheLockUtxoTypeDummy1200BidX
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeDummy1200BidX_)
		break
	case model.UtxoTypeBidY:
		cacheType = cache_service.CacheLockUtxoTypeBidpay
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeBidY_)
		perAmount = DoBidUtxoPerAmount1w
		//if totalNeedAmount > 0 && totalNeedAmount < doBidUtxoPerAmount5w {
		//	perAmount = doBidUtxoPerAmount1w
		//} else {
		//	perAmount = doBidUtxoPerAmount5w
		//}

		//else if totalNeedAmount >= doBidUtxoPerAmount5w && totalNeedAmount < doBidUtxoPerAmount10w {
		//	perAmount = doBidUtxoPerAmount5w
		//} else if totalNeedAmount >= doBidUtxoPerAmount10w && totalNeedAmount < doBidUtxoPerAmount50w {
		//	perAmount = doBidUtxoPerAmount10w
		//} else {
		//	perAmount = doBidUtxoPerAmount50w
		//}
		limit = totalNeedAmount/perAmount + 1
		break
	case model.UtxoTypeMultiInscription:
		//perAmount = 5000
		cacheType = cache_service.CacheLockUtxoTypeMultiSigInscription
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeMultiSigInscription_)
		break
	case model.UtxoTypeMultiInscriptionFromRelease:
		perAmount = 5000
		cacheType = cache_service.CacheLockUtxoTypeMultiSigInscriptionFromRelease
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeMultiSigInscriptionFromRelease_)
		break
	case model.UtxoTypeRewardInscription:
		cacheType = cache_service.CacheLockUtxoTypeRewardInscription
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeRewardInscription_)
		break
	case model.UtxoTypeRewardSend:
		cacheType = cache_service.CacheLockUtxoTypeRewardSend
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeRewardSend_)
		break
	case model.UtxoTypeLoop:
		cacheType = cache_service.CacheLockUtxoTypeLoop
		redisKeyPrefix = fmt.Sprintf("%s%s", redis.CacheGetUtxo_, redis.UtxoTypeLoop_)
		break
	default:
		return nil, errors.New("Unoccupied-Utxo: wrong type")
	}
	_ = cacheType
	unoccupiedUtxoLock.RLock()
	defer unoccupiedUtxoLock.RUnlock()

	utxoIdKeyList, sortIndexList, _ = redis.GetUtxoInfoKeyValueList(redisKeyPrefix)
	for _, v := range sortIndexList {
		if startIndex == -1 {
			startIndex = int64(v)
		} else if startIndex > int64(v) {
			startIndex = int64(v)
		}
	}
	fmt.Printf("Get utxoIdKeyList: %+v\n", utxoIdKeyList)
	fmt.Printf("Get sortIndexList: %+v\n", sortIndexList)

	confirmStatus := model.ConfirmStatus(-1)
	if utxoType == model.UtxoTypeDummyBidX || utxoType == model.UtxoTypeDummy1200BidX ||
		utxoType == model.UtxoTypeDummy || utxoType == model.UtxoTypeDummy1200 ||
		utxoType == model.UtxoTypeDummyAsk || utxoType == model.UtxoTypeDummy1200Ask ||
		utxoType == model.UtxoTypeRewardInscription || utxoType == model.UtxoTypeRewardSend ||
		utxoType == model.UtxoTypeBidY || utxoType == model.UtxoTypeMultiInscription {
		confirmStatus = model.Confirmed
	}
	if fromOrderId != "" && utxoType == model.UtxoTypeMultiInscription {
		confirmStatus = model.ConfirmStatus(-1)
	}

	utxoList, _ = mongo_service.FindUtxoList(net, startIndex, maxLimit, perAmount, utxoType, confirmStatus, fromOrderId, networkFeeRate)
	if len(utxoList) == 0 {
		return nil, errors.New("Unoccupied-Utxo: Empty utxo list. Please contact customer service and wait for the platform to add UTXO. ")
	}
	for _, v := range utxoList {
		has := false
		for _, utxoId := range utxoIdKeyList {
			if utxoId == v.UtxoId {
				has = true
				break
			}
		}
		if has {
			continue
		}
		unoccupiedUtxoList = append(unoccupiedUtxoList, v)
	}
	if int64(len(unoccupiedUtxoList)) < limit {
		fmt.Printf("Unoccupied-Utxo[%d]: Not enough - have[%d], need[%d]", utxoType, len(unoccupiedUtxoList), limit)
		return nil, errors.New(fmt.Sprintf("Unoccupied-Utxo[%d]: Not enough - have[%d], need[%d]. Please contact customer service and wait for the platform to add UTXO. ", utxoType, len(unoccupiedUtxoList), limit))
	}
	unoccupiedUtxoList = unoccupiedUtxoList[:limit]
	for _, v := range unoccupiedUtxoList {
		addr, err := btcutil.DecodeAddress(v.Address, GetNetParams(v.Net))
		if err != nil {
			return nil, err
		}
		pkScriptByte, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		v.PkScript = hex.EncodeToString(pkScriptByte)
	}

	cacheUtxoList(unoccupiedUtxoList)
	return unoccupiedUtxoList, nil
}

func ReleaseUtxoList(utxoList []*model.OrderUtxoModel) {
	for _, v := range utxoList {
		cacheUtxoType := redis.UtxoTypeDummy_
		switch v.UtxoType {
		case model.UtxoTypeDummy:
			cacheUtxoType = redis.UtxoTypeDummy_
			break
		case model.UtxoTypeDummy1200:
			cacheUtxoType = redis.UtxoTypeDummy1200_
			break
		case model.UtxoTypeDummyAsk:
			cacheUtxoType = redis.UtxoTypeDummyAsk_
			break
		case model.UtxoTypeDummy1200Ask:
			cacheUtxoType = redis.UtxoTypeDummy1200Ask_
			break
		case model.UtxoTypeDummyBidX:
			cacheUtxoType = redis.UtxoTypeDummyBidX_
			break
		case model.UtxoTypeDummy1200BidX:
			cacheUtxoType = redis.UtxoTypeDummy1200BidX_
			break
		case model.UtxoTypeBidY:
			cacheUtxoType = redis.UtxoTypeBidY_
			break
		case model.UtxoTypeMultiInscription:
			cacheUtxoType = redis.UtxoTypeMultiSigInscription_
			break
		case model.UtxoTypeMultiInscriptionFromRelease:
			cacheUtxoType = redis.UtxoTypeMultiSigInscriptionFromRelease_
			break
		case model.UtxoTypeRewardInscription:
			cacheUtxoType = redis.UtxoTypeRewardInscription_
			break
		case model.UtxoTypeRewardSend:
			cacheUtxoType = redis.UtxoTypeRewardSend_
			break
		case model.UtxoTypeLoop:
			cacheUtxoType = redis.UtxoTypeLoop_
			break
		default:
			continue
		}
		err := redis.UnSetUtxoInfo(cacheUtxoType, v.UtxoId)
		if err != nil {
			fmt.Printf("UnSetUtxoInfo err:%s\n", err.Error())
		}
	}
}

func cacheUtxoList(utxoList []*model.OrderUtxoModel) {
	for _, v := range utxoList {
		cacheUtxoType := redis.UtxoTypeDummy_
		switch v.UtxoType {
		case model.UtxoTypeDummy:
			cacheUtxoType = redis.UtxoTypeDummy_
			break
		case model.UtxoTypeDummy1200:
			cacheUtxoType = redis.UtxoTypeDummy1200_
			break
		case model.UtxoTypeDummyAsk:
			cacheUtxoType = redis.UtxoTypeDummyAsk_
			break
		case model.UtxoTypeDummy1200Ask:
			cacheUtxoType = redis.UtxoTypeDummy1200Ask_
			break
		case model.UtxoTypeDummyBidX:
			cacheUtxoType = redis.UtxoTypeDummyBidX_
			break
		case model.UtxoTypeDummy1200BidX:
			cacheUtxoType = redis.UtxoTypeDummy1200BidX_
			break
		case model.UtxoTypeBidY:
			cacheUtxoType = redis.UtxoTypeBidY_
			break
		case model.UtxoTypeMultiInscription:
			cacheUtxoType = redis.UtxoTypeMultiSigInscription_
			break
		case model.UtxoTypeMultiInscriptionFromRelease:
			cacheUtxoType = redis.UtxoTypeMultiSigInscriptionFromRelease_
			break
		case model.UtxoTypeRewardInscription:
			cacheUtxoType = redis.UtxoTypeRewardInscription_
			break
		case model.UtxoTypeRewardSend:
			cacheUtxoType = redis.UtxoTypeRewardSend_
			break
		case model.UtxoTypeLoop:
			cacheUtxoType = redis.UtxoTypeLoop_
			break
		default:
			continue
		}
		_, err := redis.SetRedisUtxoInfo(cacheUtxoType, v.UtxoId, int(v.SortIndex))
		if err != nil {
			fmt.Printf("SetRedisUtxoInfo err:%s\n", err.Error())
		}
	}
}

func GetSaveStartIndex(net string, utxoType model.UtxoType, perAmount int64) int64 {
	saveUtxoLock.RLock()
	t1 := tool.MakeTimestamp()
	fmt.Println("[LOCK]-Save-utxo")
	defer func() {
		saveUtxoLock.RUnlock()
		fmt.Printf("[UNLOCK]-Save-utxo-timeConsuming:%d\n", tool.MakeTimestamp()-t1)
	}()
	startIndex := int64(0)
	latestUtxo, _ := mongo_service.GetLatestStartIndexUtxo(net, utxoType, perAmount)
	if latestUtxo != nil {
		startIndex = latestUtxo.SortIndex
	}
	return startIndex
}
