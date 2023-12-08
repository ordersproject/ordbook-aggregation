package task

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/common_service"
	"ordbook-aggregation/service/inscription_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
	"time"
)

func JobForRewardOrder() {
	//return
	currentNetworkFee := common_service.GetFeeSummary()

	//nomal
	jobForCheckRewardOrderInscription(model.RewardTypeNormal, currentNetworkFee)
	jobForCheckRewardOrderSend(model.RewardTypeNormal, currentNetworkFee)

	//event
	jobForCheckRewardOrderInscription(model.RewardTypeEventOneLp, currentNetworkFee)
	jobForCheckRewardOrderSend(model.RewardTypeEventOneLp, currentNetworkFee)

	jobForCheckRewardOrderSendEventBid(model.RewardTypeEventOneBid)

}

func getPlatformRewardPrivateKeyAndAddress(net string, rewardType model.RewardType) (string, string) {
	switch rewardType {
	case model.RewardTypeNormal, model.RewardTypeExtra:
		return order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20(net)
	case model.RewardTypeEventOneLp, model.RewardTypeEventOneBid:
		return config.EventPlatformPrivateKeyRewardBrc20, config.EventPlatformAddressRewardBrc20
	default:
		return "", ""
	}
}

func jobForCheckRewardOrderInscription(rewardType model.RewardType, currentNetworkFeeRate int64) {
	var (
		net                       string = "livenet"
		tick                      string = config.PlatformRewardTick
		pair                      string = fmt.Sprintf("%s-BTC", strings.ToUpper(tick))
		entityList                []*model.PoolRewardOrderModel
		limit                     int64 = 50
		timestamp                 int64 = 0
		utxoRewardInscriptionList []*model.OrderUtxoModel
		commitTxHash              string = ""
		inscriptionId             string = ""
		err                       error
		utxoLimit                 int64 = 1
		//utxoLimit      int64 = 4
		revealOutValue int64 = 546
	)

	entityList, _ = mongo_service.FindPoolRewardOrderModelListByTimestamp(net, tick, pair, limit, timestamp, model.RewardStateCreate, rewardType)
	if entityList == nil || len(entityList) == 0 {
		return
	}
	for _, v := range entityList {
		if v.Address == "" {
			continue
		}
		if v.RewardState != model.RewardStateCreate {
			continue
		}
		if v.InscriptionId != "" {
			continue
		}

		utxoRewardInscriptionList, err = order_brc20_service.GetUnoccupiedUtxoList(net, utxoLimit, 0, model.UtxoTypeRewardInscription, "", currentNetworkFeeRate)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-INSCRIPTION]  [%s]get utxo err:%s", v.OrderId, err.Error()))
			order_brc20_service.ReleaseUtxoList(utxoRewardInscriptionList)
			continue
		}
		if int64(len(utxoRewardInscriptionList)) < utxoLimit {
			major.Println(fmt.Sprintf("[REWARD-INSCRIPTION]  [%s]get utxo err: not encough", v.OrderId))
			order_brc20_service.ReleaseUtxoList(utxoRewardInscriptionList)
			continue
		}

		commitTxHash, inscriptionId, err = inscriptionReward(
			utxoRewardInscriptionList,
			net, tick, v.RewardCoinAmount, revealOutValue,
			rewardType, currentNetworkFeeRate)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]inscription err: %s", v.OrderId, err.Error()))
			order_brc20_service.ReleaseUtxoList(utxoRewardInscriptionList)
			continue
		}

		v.InscriptionId = inscriptionId
		v.InscriptionOutValue = revealOutValue
		v.RewardState = model.RewardStateInscription
		_, err = mongo_service.SetPoolRewardOrderModel(v)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]SetPoolRewardOrderModel err: %s", v.OrderId, err.Error()))
			order_brc20_service.ReleaseUtxoList(utxoRewardInscriptionList)
			continue
		}
		order_brc20_service.ReleaseUtxoList(utxoRewardInscriptionList)
		order_brc20_service.SetUsedRewardUtxo(utxoRewardInscriptionList, commitTxHash)
		major.Println(fmt.Sprintf("[REWARD-INSCRIPTION] [%s]inscription success", v.OrderId))
		time.Sleep(1 * time.Second)
	}
}

func inscriptionReward(utxoList []*model.OrderUtxoModel, net, tick string, amount, revealOutValue int64, rewardType model.RewardType, currentNetworkFeeRate int64) (string, string, error) {
	var (
		netParams                                                                 *chaincfg.Params = order_brc20_service.GetNetParams(net)
		_, platformAddressRewardBrc20                                             string           = getPlatformRewardPrivateKeyAndAddress(net, rewardType)
		_, platformAddressReceiveBidValue                                         string           = order_brc20_service.GetPlatformKeyAndAddressReceiveBidValue(net)
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string           = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		transferContent                                                           string           = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, tick, amount)
		commitTxHash                                                              string           = ""
		revealTxHashList, inscriptionIdList                                       []string         = make([]string, 0), make([]string, 0)
		err                                                                       error
		brc20BalanceResult                                                        *oklink_service.OklinkBrc20BalanceDetails
		availableBalance                                                          int64 = 0
		fees                                                                      int64 = 0
		feeRate                                                                   int64 = 14
		//feeRate int64 = 10
		//feeRate          int64                               = 60
		inscribeUtxoList []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
		changeAddress    string                              = platformAddressReceiveBidValue
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
		feeRate = feeRate + int64(7)
	}
	if currentNetworkFeeRate != 0 {
		feeRate = currentNetworkFeeRate
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
			transferContent, feeRate, changeAddress, 1, inscribeUtxoList, false, "segwit", false, revealOutValue)
	if err != nil {
		return "", "", err
	}
	_ = commitTxHash
	_ = revealTxHashList
	_ = fees
	return commitTxHash, inscriptionIdList[0], nil
}

func jobForCheckRewardOrderSend(rewardType model.RewardType, currentNetworkFeeRate int64) {
	var (
		net                                                                       string = "livenet"
		tick                                                                      string = config.PlatformRewardTick
		pair                                                                      string = fmt.Sprintf("%s-BTC", strings.ToUpper(tick))
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		entityList                                                                []*model.PoolRewardOrderModel
		limit                                                                     int64                   = 50
		timestamp                                                                 int64                   = 0
		utxoRewardSendList                                                        []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		sendId                                                                    string                  = ""
		err                                                                       error
		utxoLimit                                                                 int64 = 1
		//utxoLimit             int64 = 5
	)

	entityList, _ = mongo_service.FindPoolRewardOrderModelListByTimestamp(net, tick, pair, limit, timestamp, model.RewardStateInscription, rewardType)
	if entityList == nil || len(entityList) == 0 {
		return
	}
	for _, v := range entityList {
		if v.Address == "" {
			continue
		}
		if v.RewardState != model.RewardStateInscription {
			continue
		}
		if v.InscriptionId == "" {
			continue
		}
		if v.RewardType == model.RewardTypeEventOneBid {
			if tool.MakeTimestamp()-v.Timestamp < 1000*60*60*1 {
				continue
			}
			if v.SendId != "" {
				continue
			}
		}

		if v.RewardType == model.RewardTypeEventOneBid {
			addr, err := btcutil.DecodeAddress(platformAddressRewardBrc20FeeUtxos, order_brc20_service.GetNetParams(v.Net))
			if err != nil {
				continue
			}
			pkScriptByte, err := txscript.PayToAddrScript(addr)
			if err != nil {
				continue
			}
			utxoRewardSendList = make([]*model.OrderUtxoModel, 0)
			utxoRewardSendList = append(utxoRewardSendList, &model.OrderUtxoModel{
				UtxoId:        fmt.Sprintf("%s_%d", v.FeeUtxoTxId, 1),
				Net:           v.Net,
				UtxoType:      model.UtxoTypeRewardSend,
				Amount:        uint64(v.FeeSend),
				Address:       platformAddressRewardBrc20FeeUtxos,
				PrivateKeyHex: platformPrivateKeyRewardBrc20FeeUtxos,
				TxId:          v.FeeUtxoTxId,
				Index:         1,
				PkScript:      hex.EncodeToString(pkScriptByte),
			})
		} else {
			utxoRewardSendList, err = order_brc20_service.GetUnoccupiedUtxoList(net, utxoLimit, 0, model.UtxoTypeRewardSend, "", currentNetworkFeeRate)
			if err != nil {
				major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]get utxo err:%s", v.OrderId, err.Error()))
				order_brc20_service.ReleaseUtxoList(utxoRewardSendList)
				continue
			}
		}

		if int64(len(utxoRewardSendList)) < utxoLimit {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]get utxo err: not encough", v.OrderId))
			order_brc20_service.ReleaseUtxoList(utxoRewardSendList)
			continue
		}

		sendId, err = sendReward(utxoRewardSendList, net, v.InscriptionId, v.InscriptionOutValue, v.Address, rewardType, currentNetworkFeeRate)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]send err:%s", v.OrderId, err.Error()))
			order_brc20_service.ReleaseUtxoList(utxoRewardSendList)
			continue
			//return
		}

		v.SendId = sendId
		v.RewardState = model.RewardStateSend
		_, err = mongo_service.SetPoolRewardOrderModel(v)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]SetPoolRewardOrderModel err:%s", v.OrderId, err.Error()))
			order_brc20_service.ReleaseUtxoList(utxoRewardSendList)
			continue
		}
		order_brc20_service.ReleaseUtxoList(utxoRewardSendList)
		order_brc20_service.SetUsedRewardUtxo(utxoRewardSendList, sendId)
		major.Println(fmt.Sprintf("[REWARD-SEND] [%s][%d]SEND success", v.OrderId, rewardType))
		time.Sleep(2 * time.Second)

	}
}

func sendReward(utxoList []*model.OrderUtxoModel, net, inscriptionId string, inscriptionOutValue int64, sendAddress string, rewardType model.RewardType, currentNetworkFeeRate int64) (string, error) {
	var (
		netParams *chaincfg.Params = order_brc20_service.GetNetParams(net)
		//platformPrivateKeyRewardBrc20, platformAddressRewardBrc20 string           = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20(net)
		platformPrivateKeyRewardBrc20, platformAddressRewardBrc20                 string = getPlatformRewardPrivateKeyAndAddress(net, rewardType)
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		_, platformAddressReceiveBidValue                                         string = order_brc20_service.GetPlatformKeyAndAddressReceiveBidValue(net)
		changeAddress                                                             string = platformAddressReceiveBidValue
		//changeAddress     string = ""
		inscriptionIdStrs []string
		txRaw             string = ""
		feeRate                  = int64(10)
		//feeRate = int64(60)
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

	inputs := make([]*order_brc20_service.TxInputUtxo, 0)
	brcTxInput := &order_brc20_service.TxInputUtxo{
		TxId:     inscriptionTxId,
		TxIndex:  inscriptionTxIndex,
		PkScript: hex.EncodeToString(pkScriptRewardBrc20),
		Amount:   uint64(inscriptionOutValue),
		PriHex:   platformPrivateKeyRewardBrc20,
	}
	inputs = append(inputs, brcTxInput)

	//utxoFeeRate := currentNetworkFeeRate
	for _, utxo := range utxoList {
		if utxo.Address != platformAddressRewardBrc20FeeUtxos {
			continue
		}
		txInput := &order_brc20_service.TxInputUtxo{
			TxId:     utxo.TxId,
			TxIndex:  utxo.Index,
			PkScript: hex.EncodeToString(pkScriptRewardBrc20FeeUtxos),
			Amount:   uint64(utxo.Amount),
			PriHex:   platformPrivateKeyRewardBrc20FeeUtxos,
		}
		inputs = append(inputs, txInput)
		feeRate = feeRate + int64(2)

		if utxo.NetworkFeeRate != 0 {
			currentNetworkFeeRate = utxo.NetworkFeeRate
		}
	}
	if currentNetworkFeeRate != 0 {
		feeRate = currentNetworkFeeRate - 5
	}

	outputs := make([]*order_brc20_service.TxOutput, 0)
	outputs = append(outputs, &order_brc20_service.TxOutput{
		Address: sendAddress,
		Amount:  int64(inscriptionOutValue),
	})
	tx, err := order_brc20_service.BuildCommonTxV2(netParams, inputs, outputs, changeAddress, feeRate)
	if err != nil {
		fmt.Printf("[REWARD-SEND]BuildCommonTx err:%s\n", err.Error())
		return "", err
	}

	txRaw, err = order_brc20_service.ToRaw(tx)
	if err != nil {
		return "", err
	}
	txResp, err := unisat_service.BroadcastTx(net, txRaw)
	if err != nil {
		return "", err
	}
	return txResp.Result, nil
}

func jobForCheckRewardOrderSendEventBid(rewardType model.RewardType) {
	var (
		net                                                                       string = "livenet"
		tick                                                                      string = config.PlatformRewardTick
		pair                                                                      string = fmt.Sprintf("%s-BTC", strings.ToUpper(tick))
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		entityList                                                                []*model.PoolRewardOrderModel
		limit                                                                     int64  = 50
		timestamp                                                                 int64  = 0
		sendId                                                                    string = ""
	)

	entityList, _ = mongo_service.FindPoolRewardOrderModelListByTimestamp(net, tick, pair, limit, timestamp, model.RewardStateInscription, rewardType)
	if entityList == nil || len(entityList) == 0 {
		return
	}
	for _, v := range entityList {
		if v.Address == "" {
			continue
		}
		if v.RewardState != model.RewardStateInscription {
			continue
		}
		if v.InscriptionId == "" {
			continue
		}
		if v.RewardType == model.RewardTypeEventOneBid {
			if tool.MakeTimestamp()-v.Timestamp < 1000*60*60*1 {
				continue
			}
			if v.SendId != "" {
				continue
			}
		}

		addr, err := btcutil.DecodeAddress(platformAddressRewardBrc20FeeUtxos, order_brc20_service.GetNetParams(v.Net))
		if err != nil {
			continue
		}
		pkScriptByte, err := txscript.PayToAddrScript(addr)
		if err != nil {
			continue
		}
		utxoRewardSendList := make([]*model.OrderUtxoModel, 0)
		utxoRewardSendList = append(utxoRewardSendList, &model.OrderUtxoModel{
			UtxoId:        fmt.Sprintf("%s_%d", v.FeeUtxoTxId, 1),
			Net:           v.Net,
			UtxoType:      model.UtxoTypeRewardSend,
			Amount:        uint64(v.FeeSend),
			Address:       platformAddressRewardBrc20FeeUtxos,
			PrivateKeyHex: platformPrivateKeyRewardBrc20FeeUtxos,
			TxId:          v.FeeUtxoTxId,
			Index:         1,
			PkScript:      hex.EncodeToString(pkScriptByte),
		})

		//send
		sendId, err = order_brc20_service.SendReward(v.RewardType, utxoRewardSendList,
			v.Net, v.InscriptionId, v.InscriptionOutValue, v.Address,
			v.NetworkFeeRate)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]send err:%s", v.OrderId, err.Error()))
			continue
		}

		v.SendId = sendId
		v.RewardState = model.RewardStateSend
		v.FeeRawTx = ""
		_, err = mongo_service.SetPoolRewardOrderModel(v)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]SetPoolRewardOrderModel err:%s", v.OrderId, err.Error()))
			continue
		}
		time.Sleep(2 * time.Second)
	}
}
