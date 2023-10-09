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
	"ordbook-aggregation/service/inscription_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/service/unisat_service"
	"strconv"
	"strings"
	"time"
)

func jobForRewardOrder() {
	jobForCheckRewardOrderInscription()
	jobForCheckRewardOrderSend()
}

func jobForCheckRewardOrderInscription() {
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
		revealOutValue            int64 = 546
	)

	entityList, _ = mongo_service.FindPoolRewardOrderModelListByTimestamp(net, tick, pair, limit, timestamp, model.RewardStateCreate)
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

		utxoRewardInscriptionList, err = order_brc20_service.GetUnoccupiedUtxoList(net, utxoLimit, 0, model.UtxoTypeRewardInscription)
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
			utxoRewardInscriptionList[0].TxId, utxoRewardInscriptionList[0].Index, int64(utxoRewardInscriptionList[0].Amount),
			net, tick, v.RewardCoinAmount, revealOutValue)
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

func inscriptionReward(utxoTxId string, utxoTxIndex, utxoAmount int64, net, tick string, amount, revealOutValue int64) (string, string, error) {
	var (
		netParams                                *chaincfg.Params = order_brc20_service.GetNetParams(net)
		_, platformAddressRewardBrc20            string           = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20(net)
		platformPrivateKeyRewardBrc20FeeUtxos, _ string           = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		transferContent                          string           = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, tick, amount)
		commitTxHash                             string           = ""
		revealTxHashList, inscriptionIdList      []string         = make([]string, 0), make([]string, 0)
		err                                      error
		brc20BalanceResult                       *oklink_service.OklinkBrc20BalanceDetails
		availableBalance                         int64                               = 0
		fees                                     int64                               = 0
		feeRate                                  int64                               = 14
		inscribeUtxoList                         []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
		changeAddress                            string                              = ""
	)
	inscribeUtxoList = append(inscribeUtxoList, &inscription_service.InscribeUtxo{
		OutTx:     utxoTxId,
		OutIndex:  utxoTxIndex,
		OutAmount: utxoAmount,
	})

	fmt.Println(transferContent)
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
			transferContent, feeRate, changeAddress, 1, inscribeUtxoList, "segwit", false, revealOutValue)
	if err != nil {
		return "", "", err
	}
	_ = commitTxHash
	_ = revealTxHashList
	_ = fees
	return commitTxHash, inscriptionIdList[0], nil
}

func jobForCheckRewardOrderSend() {
	var (
		net                string = "livenet"
		tick               string = config.PlatformRewardTick
		pair               string = fmt.Sprintf("%s-BTC", strings.ToUpper(tick))
		entityList         []*model.PoolRewardOrderModel
		limit              int64 = 50
		timestamp          int64 = 0
		utxoRewardSendList []*model.OrderUtxoModel
		sendId             string = ""
		err                error
		utxoLimit          int64 = 1
	)

	entityList, _ = mongo_service.FindPoolRewardOrderModelListByTimestamp(net, tick, pair, limit, timestamp, model.RewardStateInscription)
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

		utxoRewardSendList, err = order_brc20_service.GetUnoccupiedUtxoList(net, utxoLimit, 0, model.UtxoTypeRewardSend)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]get utxo err:%s", v.OrderId, err.Error()))
			order_brc20_service.ReleaseUtxoList(utxoRewardSendList)
			continue
		}
		if int64(len(utxoRewardSendList)) < utxoLimit {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]get utxo err: not encough", v.OrderId))
			order_brc20_service.ReleaseUtxoList(utxoRewardSendList)
			continue
		}

		sendId, err = sendReward(utxoRewardSendList[0].TxId, utxoRewardSendList[0].Index, int64(utxoRewardSendList[0].Amount), net, v.InscriptionId, v.InscriptionOutValue, v.Address)
		if err != nil {
			major.Println(fmt.Sprintf("[REWARD-SEND]  [%s]send err:%s", v.OrderId, err.Error()))
			order_brc20_service.ReleaseUtxoList(utxoRewardSendList)
			continue
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
		major.Println(fmt.Sprintf("[REWARD-SEND] [%s]SEND success", v.OrderId))
		time.Sleep(1 * time.Second)
	}
}

func sendReward(utxoTxId string, utxoTxIndex, utxoAmount int64, net, inscriptionId string, inscriptionOutValue int64, sendAddress string) (string, error) {
	var (
		netParams                                                                 *chaincfg.Params = order_brc20_service.GetNetParams(net)
		platformPrivateKeyRewardBrc20, platformAddressRewardBrc20                 string           = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20(net)
		platformPrivateKeyRewardBrc20FeeUtxos, platformAddressRewardBrc20FeeUtxos string           = order_brc20_service.GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		_, platformAddressReceiveBidValue                                         string           = order_brc20_service.GetPlatformKeyAndAddressReceiveBidValue(net)
		changeAddress                                                             string           = platformAddressReceiveBidValue
		inscriptionIdStrs                                                         []string
		txRaw                                                                     string = ""
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
		return "", nil
	}
	pkScriptRewardBrc20, err := txscript.PayToAddrScript(addrRewardBrc20)
	if err != nil {
		return "", nil
	}

	addrRewardBrc20FeeUtxos, err := btcutil.DecodeAddress(platformAddressRewardBrc20FeeUtxos, netParams)
	if err != nil {
		return "", nil
	}
	pkScriptRewardBrc20FeeUtxos, err := txscript.PayToAddrScript(addrRewardBrc20FeeUtxos)
	if err != nil {
		return "", nil
	}

	fee := int64(10)
	inputs := make([]*order_brc20_service.TxInputUtxo, 0)
	inputs = append(inputs, &order_brc20_service.TxInputUtxo{
		TxId:     inscriptionTxId,
		TxIndex:  inscriptionTxIndex,
		PkScript: hex.EncodeToString(pkScriptRewardBrc20),
		Amount:   uint64(inscriptionOutValue),
		PriHex:   platformPrivateKeyRewardBrc20,
	})
	inputs = append(inputs, &order_brc20_service.TxInputUtxo{
		TxId:     utxoTxId,
		TxIndex:  utxoTxIndex,
		PkScript: hex.EncodeToString(pkScriptRewardBrc20FeeUtxos),
		Amount:   uint64(utxoAmount),
		PriHex:   platformPrivateKeyRewardBrc20FeeUtxos,
	})

	outputs := make([]*order_brc20_service.TxOutput, 0)
	outputs = append(outputs, &order_brc20_service.TxOutput{
		Address: sendAddress,
		Amount:  int64(inscriptionOutValue),
	})
	tx, err := order_brc20_service.BuildCommonTx(netParams, inputs, outputs, changeAddress, fee)
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
