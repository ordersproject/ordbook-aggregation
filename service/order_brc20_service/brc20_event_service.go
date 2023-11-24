package order_brc20_service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"ordbook-aggregation/config"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/inscription_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
	"time"
)

func FetchEventOwnerReward(req *request.PoolBrc20RewardReq) (*respond.PoolBrc20RewardResp, error) {
	var (
		entityRewardSeller       *model.EventRewardCount
		entityRewardBuyer        *model.EventRewardCount
		entityRewardOrderCount   *model.PoolRewardOrderCount
		totalRewardAmount        uint64 = 0
		totalRewardExtraAmount   uint64 = 0
		hadClaimRewardAmount     uint64 = 0
		hasReleasePoolOrderCount int64  = 0
		tick                     string = "rdex"
		rewardTick               string = config.EventOneRewardTick
	)

	//if req.Tick != config.PlatformRewardTick {
	//	return nil, errors.New(fmt.Sprintf("tick wrong:%s", config.PlatformRewardTick))
	//}

	//hasReleasePoolOrderCount, _ = mongo_service.CountPoolBrc20ModelList(req.Net, req.Tick, "", req.Address, model.PoolTypeAll, model.PoolStateUsed)

	entityRewardSeller, _ = mongo_service.CountOwnEventOrderBrc20RewardBySeller(req.Net, tick, req.Address, config.EventOneStartTime)
	if entityRewardSeller != nil {
		totalRewardAmount = totalRewardAmount + uint64(entityRewardSeller.RewardAmountTotal)/2
	}
	entityRewardBuyer, _ = mongo_service.CountOwnEventOrderBrc20RewardByBuyer(req.Net, tick, "", req.Address, config.EventOneStartTime)
	if entityRewardBuyer != nil {
		totalRewardAmount = totalRewardAmount + uint64(entityRewardBuyer.RewardAmountTotal)/2
	}

	entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, rewardTick, "", req.Address, model.RewardTypeEventOneBid)
	if entityRewardOrderCount != nil {
		hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)
	}
	return &respond.PoolBrc20RewardResp{
		Net:                      req.Net,
		Tick:                     req.Tick,
		RewardTick:               rewardTick,
		TotalRewardAmount:        totalRewardAmount,
		TotalRewardExtraAmount:   totalRewardExtraAmount,
		HadClaimRewardAmount:     hadClaimRewardAmount,
		HasReleasePoolOrderCount: hasReleasePoolOrderCount,
	}, nil
}

func CalEventClaimFee(req *request.OrderBrc20CalFeeReq) (*respond.CalFeeResp, error) {
	var (
		feeAmountForRewardInscription         int64  = 4000
		feeAmountForRewardSend                int64  = 4000
		_, platformAddressRewardBrc20FeeUtxos string = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(req.Net)
	)
	if req.Version == 2 {
		_, feeAmountForRewardInscription, feeAmountForRewardSend = GenerateBidTakerFee(req.NetworkFeeRate)
	}
	return &respond.CalFeeResp{
		RewardInscriptionFee: feeAmountForRewardInscription,
		RewardSendFee:        feeAmountForRewardSend,
		FeeAddress:           platformAddressRewardBrc20FeeUtxos,
	}, nil
}

func ClaimEventReward(req *request.PoolBrc20ClaimRewardReq, publicKey, ip string) (string, error) {
	var (
		netParams                                                                 *chaincfg.Params = GetNetParams(req.Net)
		orderId                                                                   string           = ""
		entityOrder                                                               *model.PoolRewardOrderModel
		nowTime                                                                   int64 = tool.MakeTimestamp()
		entityRewardSeller                                                        *model.EventRewardCount
		entityRewardBuyer                                                         *model.EventRewardCount
		entityRewardOrderCount                                                    *model.PoolRewardOrderCount
		entityBlockReward                                                         *model.PoolRewardBlockUserCount
		totalRewardAmount                                                         uint64 = 0
		totalRewardExtraAmount                                                    uint64 = 0
		hadClaimRewardAmount                                                      uint64 = 0
		remainingRewardAmount                                                     int64  = 0
		tick                                                                      string = "rdex"
		rewardTick                                                                string = config.EventOneRewardTick
		revealOutValue                                                            int64  = 546
		_, platformAddressRewardBrc20                                             string = config.EventPlatformPrivateKeyRewardBrc20, config.EventPlatformAddressRewardBrc20
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(req.Net)
		feeAmountForRewardInscription                                             int64  = 4000
		feeAmountForRewardSend                                                    int64  = 4000
	)
	//if req.Tick != config.EventOneRewardTick {
	//	return "", errors.New(fmt.Sprintf("tick wrong:%s", config.EventOneRewardTick))
	//}

	verified, err := CheckPublicKeyAddress(netParams, publicKey, req.Address)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
	}
	if !verified {
		return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
	}

	_ = entityBlockReward
	//entityBlockReward, _ = mongo_service.CountPoolRewardBlockUser(req.Net, req.Address)
	//if entityBlockReward != nil {
	//	totalRewardAmount = uint64(entityBlockReward.RewardCoinAmountTotal)
	//	entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, "", "", req.Address)
	//	if entityRewardOrderCount != nil {
	//		hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)
	//		remainingRewardAmount = int64(totalRewardAmount) - int64(hadClaimRewardAmount)
	//	}
	//}

	entityRewardSeller, _ = mongo_service.CountOwnEventOrderBrc20RewardBySeller(req.Net, tick, req.Address, config.EventOneStartTime)
	if entityRewardSeller != nil {
		totalRewardAmount = totalRewardAmount + uint64(entityRewardSeller.RewardAmountTotal)/2
	}
	entityRewardBuyer, _ = mongo_service.CountOwnEventOrderBrc20RewardByBuyer(req.Net, tick, "", req.Address, config.EventOneStartTime)
	if entityRewardBuyer != nil {
		totalRewardAmount = totalRewardAmount + uint64(entityRewardBuyer.RewardAmountTotal)/2
	}

	entityRewardOrderCount, _ = mongo_service.CountOwnPoolRewardOrder(req.Net, rewardTick, "", req.Address, model.RewardTypeEventOneBid)
	if entityRewardOrderCount != nil {
		hadClaimRewardAmount = uint64(entityRewardOrderCount.RewardCoinAmountTotal)

		remainingRewardAmount = int64(totalRewardAmount+totalRewardExtraAmount) - int64(hadClaimRewardAmount)
	}

	if remainingRewardAmount < 0 {
		remainingRewardAmount = 0
	}
	if remainingRewardAmount < req.RewardAmount || req.RewardAmount <= 0 {
		return "", errors.New(fmt.Sprintf("You only have %d rdex to claim.", remainingRewardAmount))
	}

	brc20BalanceResult, err := oklink_service.GetAddressBrc20BalanceResult(platformAddressRewardBrc20, rewardTick, 1, 50)
	if err != nil {
		return "", err
	}
	availableBalance, _ := strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	fmt.Printf("availableBalance:%d, req.InscribeTransferAmount*req.Count: %d\n", availableBalance, req.RewardAmount*1)
	if availableBalance < req.RewardAmount*1 {
		return "", errors.New("AvailableBalance not enough. ")
	}

	orderId = fmt.Sprintf("%s_%s_%s_%d_%d_%d", req.Net, rewardTick, req.Address, req.RewardAmount, nowTime, model.RewardTypeEventOneBid)
	orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
	entityOrder, _ = mongo_service.FindPoolRewardOrderModelByOrderId(orderId)
	if entityOrder != nil {
		return "", errors.New("already exist")
	}

	entityOrder = &model.PoolRewardOrderModel{
		Net:              req.Net,
		Tick:             rewardTick,
		OrderId:          orderId,
		Pair:             fmt.Sprintf("%s-BTC", strings.ToUpper(rewardTick)),
		RewardCoinAmount: req.RewardAmount,
		Address:          req.Address,
		RewardState:      model.RewardStateCreate,
		Timestamp:        nowTime,
		RewardType:       model.RewardTypeEventOneBid,
		FeeRawTx:         req.FeeRawTx,
		FeeUtxoTxId:      req.FeeUtxoTxId,
		FeeInscription:   req.FeeInscription,
		FeeSend:          req.FeeSend,
		NetworkFeeRate:   req.NetworkFeeRate,
		Version:          req.Version,
	}

	txRawByte, _ := hex.DecodeString(req.FeeRawTx)
	tx := wire.NewMsgTx(2)
	err = tx.Deserialize(bytes.NewReader(txRawByte))
	if err != nil {
		fmt.Printf(fmt.Sprintf("[REWARD-INSCRIPTION]  feeRawTx Deserialize err:%s", err.Error()))
		return "", err
	}
	entityOrder.FeeUtxoTxId = tx.TxHash().String()

	_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
	if err != nil {
		return "", errors.New("create order err")
	}

	addr, err := btcutil.DecodeAddress(platformAddressRewardBrc20FeeUtxos, GetNetParams(entityOrder.Net))
	if err != nil {
		return "", err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}

	if len(tx.TxOut) < 2 {
		return "", errors.New("feeRawTx wrong")
	}
	if tx.TxOut[0].Value != entityOrder.FeeInscription || tx.TxOut[1].Value != entityOrder.FeeSend ||
		hex.EncodeToString(tx.TxOut[0].PkScript) != hex.EncodeToString(pkScriptByte) ||
		hex.EncodeToString(tx.TxOut[1].PkScript) != hex.EncodeToString(pkScriptByte) ||
		tx.TxHash().String() != entityOrder.FeeUtxoTxId {
		return "", errors.New("feeRawTx wrong")
	}

	if entityOrder.Version == 2 {
		_, feeAmountForRewardInscription, feeAmountForRewardSend = GenerateBidTakerFee(req.NetworkFeeRate)
		if entityOrder.FeeInscription != feeAmountForRewardInscription || entityOrder.FeeSend != feeAmountForRewardSend {
			return "", errors.New("feeAmount wrong")
		}

		txRaw, err := ToRaw(tx)
		if err != nil {
			return "", err
		}
		_, err = unisat_service.BroadcastTx(req.Net, txRaw)
		if err != nil {
			return "", err
		}

		utxoRewardInscriptionList := make([]*model.OrderUtxoModel, 0)
		utxoRewardSendList := make([]*model.OrderUtxoModel, 0)

		utxoRewardInscriptionList = append(utxoRewardInscriptionList, &model.OrderUtxoModel{
			UtxoId:        fmt.Sprintf("%s_%d", req.FeeUtxoTxId, 0),
			Net:           entityOrder.Net,
			UtxoType:      model.UtxoTypeRewardInscription,
			Amount:        uint64(entityOrder.FeeInscription),
			Address:       platformAddressRewardBrc20FeeUtxos,
			PrivateKeyHex: platformPrivateKeyRewardBrc20FeeUtxos,
			TxId:          entityOrder.FeeUtxoTxId,
			Index:         0,
			PkScript:      hex.EncodeToString(pkScriptByte),
		})
		utxoRewardSendList = append(utxoRewardSendList, &model.OrderUtxoModel{
			UtxoId:        fmt.Sprintf("%s_%d", req.FeeUtxoTxId, 1),
			Net:           entityOrder.Net,
			UtxoType:      model.UtxoTypeRewardSend,
			Amount:        uint64(entityOrder.FeeSend),
			Address:       platformAddressRewardBrc20FeeUtxos,
			PrivateKeyHex: platformPrivateKeyRewardBrc20FeeUtxos,
			TxId:          entityOrder.FeeUtxoTxId,
			Index:         1,
			PkScript:      hex.EncodeToString(pkScriptByte),
		})

		//inscribe
		time.Sleep(1500 * time.Millisecond)
		_, inscriptionId, err := inscriptionRewardForEvent(
			utxoRewardInscriptionList,
			entityOrder.Net, entityOrder.Tick, entityOrder.RewardCoinAmount, revealOutValue,
			entityOrder.NetworkFeeRate, req.Address)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]err: %s", entityOrder.OrderId, err.Error()))
			return "", err
		}
		entityOrder.InscriptionId = inscriptionId
		entityOrder.InscriptionOutValue = revealOutValue
		entityOrder.RewardState = model.RewardStateInscription
		_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]SetPoolRewardOrderModel err: %s", entityOrder.OrderId, err.Error()))
			return "", err
		}

		//send
		sendId, err := sendReward(utxoRewardSendList,
			entityOrder.Net, entityOrder.InscriptionId, entityOrder.InscriptionOutValue, entityOrder.Address,
			entityOrder.NetworkFeeRate)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]send err:%s", entityOrder.OrderId, err.Error()))
			return "", err
		}

		entityOrder.SendId = sendId
		entityOrder.RewardState = model.RewardStateSend
		entityOrder.FeeRawTx = ""
		_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]SetPoolRewardOrderModel err:%s", entityOrder.OrderId, err.Error()))
			return "", err
		}
	}
	return "success", nil
}

func inscriptionRewardForEvent(utxoList []*model.OrderUtxoModel, net, tick string, amount, revealOutValue int64, currentNetworkFeeRate int64, changeAddress string) (string, string, error) {
	var (
		netParams                                                                 *chaincfg.Params = GetNetParams(net)
		_, platformAddressRewardBrc20                                             string           = config.EventPlatformPrivateKeyRewardBrc20, config.EventPlatformAddressRewardBrc20
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string           = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		transferContent                                                           string           = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, tick, amount)
		commitTxHash                                                              string           = ""
		revealTxHashList, inscriptionIdList                                       []string         = make([]string, 0), make([]string, 0)
		err                                                                       error
		brc20BalanceResult                                                        *oklink_service.OklinkBrc20BalanceDetails
		availableBalance                                                          int64                               = 0
		fees                                                                      int64                               = 0
		feeRate                                                                   int64                               = currentNetworkFeeRate
		inscribeUtxoList                                                          []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
	)
	for _, utxo := range utxoList {
		if utxo.Address != platformAddressRewardBrc20FeeUtxos {
			continue
		}
		inscribeUtxoList = append(inscribeUtxoList, &inscription_service.InscribeUtxo{
			OutTx:     utxo.TxId,
			OutIndex:  utxo.Index,
			OutAmount: int64(utxo.Amount),
		})
	}

	fmt.Println(transferContent)
	fmt.Println(feeRate)
	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressRewardBrc20, tick, 1, 50)
	if err != nil {
		return "", "", err
	}
	availableBalance, _ = strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	fmt.Printf("availableBalance:%d, req.InscribeTransferAmount*req.Count: %d\n", availableBalance, amount*1)
	if availableBalance < amount*1 {
		return "", "", errors.New("AvailableBalance not enough. ")
	}
	commitTxHash, revealTxHashList, inscriptionIdList, fees, err =
		inscription_service.InscribeMultiDataFromUtxo(netParams, platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20,
			transferContent, feeRate, changeAddress, 1, inscribeUtxoList, true, "segwit", false, revealOutValue)
	if err != nil {
		return "", "", err
	}
	_ = commitTxHash
	_ = revealTxHashList
	_ = fees
	return commitTxHash, inscriptionIdList[0], nil
}

func sendReward(utxoList []*model.OrderUtxoModel, net, inscriptionId string, inscriptionOutValue int64, sendAddress string, currentNetworkFeeRate int64) (string, error) {
	var (
		netParams                                                                 *chaincfg.Params = GetNetParams(net)
		platformPrivateKeyRewardBrc20, platformAddressRewardBrc20                 string           = config.EventPlatformPrivateKeyRewardBrc20, config.EventPlatformAddressRewardBrc20
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string           = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		changeAddress                                                             string           = sendAddress
		inscriptionIdStrs                                                         []string
		txRaw                                                                     string = ""
		feeRate                                                                          = currentNetworkFeeRate
	)
	if inscriptionId == "" {
		return "", errors.New("inscriptionId is empty")
	}
	inscriptionIdStrs = strings.Split(inscriptionId, "i")
	if len(inscriptionIdStrs) < 2 {
		return "", errors.New("inscriptionId format invalid")
	}
	inscriptionTxId := inscriptionIdStrs[0]
	inscriptionTxIndex, _ := strconv.ParseInt(inscriptionIdStrs[1], 10, 64)

	addrRewardBrc20, err := btcutil.DecodeAddress(platformAddressRewardBrc20, netParams)
	if err != nil {
		return "", err
	}
	pkScriptRewardBrc20, err := txscript.PayToAddrScript(addrRewardBrc20)
	if err != nil {
		return "", err
	}

	addrRewardBrc20FeeUtxos, err := btcutil.DecodeAddress(platformAddressRewardBrc20FeeUtxos, netParams)
	if err != nil {
		return "", err
	}
	pkScriptRewardBrc20FeeUtxos, err := txscript.PayToAddrScript(addrRewardBrc20FeeUtxos)
	if err != nil {
		return "", err
	}

	inputs := make([]*TxInputUtxo, 0)
	inputs = append(inputs, &TxInputUtxo{
		TxId:     inscriptionTxId,
		TxIndex:  inscriptionTxIndex,
		PkScript: hex.EncodeToString(pkScriptRewardBrc20),
		Amount:   uint64(inscriptionOutValue),
		PriHex:   platformPrivateKeyRewardBrc20,
	})

	for _, utxo := range utxoList {
		if utxo.Address != platformAddressRewardBrc20FeeUtxos {
			continue
		}
		inputs = append(inputs, &TxInputUtxo{
			TxId:     utxo.TxId,
			TxIndex:  utxo.Index,
			PkScript: hex.EncodeToString(pkScriptRewardBrc20FeeUtxos),
			Amount:   uint64(utxo.Amount),
			PriHex:   platformPrivateKeyRewardBrc20FeeUtxos,
		})
	}

	outputs := make([]*TxOutput, 0)
	outputs = append(outputs, &TxOutput{
		Address: sendAddress,
		Amount:  int64(inscriptionOutValue),
	})
	tx, err := BuildCommonTx(netParams, inputs, outputs, changeAddress, feeRate)
	if err != nil {
		fmt.Printf("[REWARD-SEND]BuildCommonTx err:%s\n", err.Error())
		return "", err
	}

	txRaw, err = ToRaw(tx)
	if err != nil {
		return "", err
	}
	txResp, err := unisat_service.BroadcastTx(net, txRaw)
	if err != nil {
		return "", err
	}
	return txResp.Result, nil
}

func ClaimJob(orderId string) (string, error) {
	var (
		entityOrder    *model.PoolRewardOrderModel
		revealOutValue int64 = 546
		//_, platformAddressRewardBrc20                                             string = config.EventPlatformPrivateKeyRewardBrc20, config.EventPlatformAddressRewardBrc20

		feeAmountForRewardInscription int64 = 4000
		feeAmountForRewardSend        int64 = 4000
	)
	entityOrder, _ = mongo_service.FindPoolRewardOrderModelByOrderId(orderId)
	if entityOrder == nil {
		return "", errors.New("empty")
	}
	platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos := GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(entityOrder.Net)

	addr, err := btcutil.DecodeAddress(platformAddressRewardBrc20FeeUtxos, GetNetParams(entityOrder.Net))
	if err != nil {
		return "", err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}

	_, feeAmountForRewardInscription, feeAmountForRewardSend = GenerateBidTakerFee(entityOrder.NetworkFeeRate)
	if entityOrder.FeeInscription != feeAmountForRewardInscription || entityOrder.FeeSend != feeAmountForRewardSend {
		return "", errors.New("feeAmount wrong")
	}

	utxoRewardInscriptionList := make([]*model.OrderUtxoModel, 0)
	utxoRewardSendList := make([]*model.OrderUtxoModel, 0)

	utxoRewardInscriptionList = append(utxoRewardInscriptionList, &model.OrderUtxoModel{
		UtxoId:        fmt.Sprintf("%s_%d", entityOrder.FeeUtxoTxId, 0),
		Net:           entityOrder.Net,
		UtxoType:      model.UtxoTypeRewardInscription,
		Amount:        uint64(entityOrder.FeeInscription),
		Address:       platformAddressRewardBrc20FeeUtxos,
		PrivateKeyHex: platformPrivateKeyRewardBrc20FeeUtxos,
		TxId:          entityOrder.FeeUtxoTxId,
		Index:         0,
		PkScript:      hex.EncodeToString(pkScriptByte),
	})
	utxoRewardSendList = append(utxoRewardSendList, &model.OrderUtxoModel{
		UtxoId:        fmt.Sprintf("%s_%d", entityOrder.FeeUtxoTxId, 1),
		Net:           entityOrder.Net,
		UtxoType:      model.UtxoTypeRewardSend,
		Amount:        uint64(entityOrder.FeeSend),
		Address:       platformAddressRewardBrc20FeeUtxos,
		PrivateKeyHex: platformPrivateKeyRewardBrc20FeeUtxos,
		TxId:          entityOrder.FeeUtxoTxId,
		Index:         1,
		PkScript:      hex.EncodeToString(pkScriptByte),
	})

	//inscribe
	_, inscriptionId, err := inscriptionRewardForEvent(
		utxoRewardInscriptionList,
		entityOrder.Net, entityOrder.Tick, entityOrder.RewardCoinAmount, revealOutValue,
		entityOrder.NetworkFeeRate, entityOrder.Address)
	if err != nil {
		major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]err: %s", entityOrder.OrderId, err.Error()))
		return "", err
	}
	entityOrder.InscriptionId = inscriptionId
	entityOrder.InscriptionOutValue = revealOutValue
	entityOrder.RewardState = model.RewardStateInscription
	_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
	if err != nil {
		major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]SetPoolRewardOrderModel err: %s", entityOrder.OrderId, err.Error()))
		return "", err
	}

	//send
	sendId, err := sendReward(utxoRewardSendList,
		entityOrder.Net, entityOrder.InscriptionId, entityOrder.InscriptionOutValue, entityOrder.Address,
		entityOrder.NetworkFeeRate)
	if err != nil {
		major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]send err:%s", entityOrder.OrderId, err.Error()))
		return "", err
	}

	entityOrder.SendId = sendId
	entityOrder.RewardState = model.RewardStateSend
	entityOrder.FeeRawTx = ""
	_, err = mongo_service.SetPoolRewardOrderModel(entityOrder)
	if err != nil {
		major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]SetPoolRewardOrderModel err:%s", entityOrder.OrderId, err.Error()))
		return "", err
	}
	return "success", nil
}
