package order_brc20_service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strings"
)

const (
	claimFetchLimit int64 = 1
	claimDayLimit   int64 = 1
)

func FetchClaimOrder(req *request.OrderBrc20ClaimFetchOneReq, publicKey, ip string) (*respond.Brc20ClaimItem, error) {
	var (
		entity                       *model.OrderBrc20Model
		netParams                    *chaincfg.Params = GetNetParams(req.Net)
		count                        int64            = 0
		todayStartTime, todayEndTime int64            = tool.GetToday0Time(), tool.GetToday24Time()
		err                          error
	)
	_ = count
	_ = todayStartTime
	_ = todayEndTime

	canCount, err := getWhitelistCount(req.Address, ip, model.WhitelistTypeClaim)
	if err != nil {
		return nil, err
	}
	if canCount <= 0 {
		return nil, errors.New("already had claimed")
	}

	entity, err = GetUnoccupiedClaimBrc20PsbtList(req.Net, req.Tick, claimFetchLimit)
	if err != nil {
		return nil, err
	}
	//defer ReleaseClaimOrderList([]*model.OrderBrc20Model{entity})
	if entity == nil {
		return nil, errors.New("Claim Order is empty. ")
	}

	netParams = GetNetParams(entity.Net)

	verified, err := CheckPublicKeyAddress(netParams, publicKey, req.Address)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
	}
	if !verified {
		return nil, errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
	}

	if entity.FreeState == model.FreeStateClaim {

		//return nil, errors.New("Claim Not Yet. ")
		//todo
		//if count >= claimDayLimit {
		//	return nil, errors.New(fmt.Sprintf("The number of purchases of the day has exceeded. "))
		//}
	}

	item := &respond.Brc20ClaimItem{
		Net:        entity.Net,
		OrderId:    entity.OrderId,
		Tick:       entity.Tick,
		Fee:        entity.Amount,
		CoinAmount: entity.CoinAmount,
		PsbtRaw:    entity.PsbtRawPreAsk,
	}
	return item, nil
}

func UpdateClaimOrder(req *request.OrderBrc20ClaimUpdateReq, publicKey, ip string) (string, error) {
	var (
		finalAskPsbtBuilder *PsbtBuilder
		netParams           *chaincfg.Params = GetNetParams(req.Net)
		entityOrder         *model.OrderBrc20Model
		err                 error
	)
	entityOrder, _ = mongo_service.FindOrderBrc20ModelByOrderId(req.OrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("Order is empty. ")
	}

	defer ReleaseClaimOrderList([]*model.OrderBrc20Model{entityOrder})

	finalAskPsbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
	if err != nil {
		return "", errors.New(fmt.Sprintf("PSBT: NewPsbtBuilder err:%s", err.Error()))
	}
	txRaw, err := finalAskPsbtBuilder.ExtractPsbtTransaction()
	if err != nil {
		return "", errors.New(fmt.Sprintf("PSBT: ExtractPsbtTransaction err:%s", err.Error()))
	}

	if len(finalAskPsbtBuilder.GetInputs()) < 4 {
		return "", errors.New(fmt.Sprintf("PSBT: No match inputs length err"))
	}
	buyerAddress := ""
	buyerInput := finalAskPsbtBuilder.GetInputs()[3]
	buyerInputTxId := buyerInput.PreviousOutPoint.Hash.String()
	buyerInputIndex := buyerInput.PreviousOutPoint.Index
	if strings.ToLower(req.Net) != "testnet" {
		buyerTx, err := oklink_service.GetTxDetail(buyerInputTxId)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Get Buyer preTx err:%s", err.Error()))
		}
		buyerAddress = buyerTx.OutputDetails[buyerInputIndex].OutputHash
	} else {
		buyerAddress = req.Address
	}

	verified, err := CheckPublicKeyAddress(netParams, publicKey, buyerAddress)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
	}
	if !verified {
		return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
	}

	entityOrder.BuyerAddress = buyerAddress
	entityOrder.BuyerIp = ip

	txRawByte, _ := hex.DecodeString(txRaw)
	txAsk := wire.NewMsgTx(2)
	err = txAsk.Deserialize(bytes.NewReader(txRawByte))
	if err != nil {
		return "", errors.New(fmt.Sprintf("txAsk Deserialize err: %v. ", err.Error()))
	}
	txId := txAsk.TxHash().String()

	entityOrder.PsbtAskTxId = txId

	txPsbtResp, err := unisat_service.BroadcastTx(entityOrder.Net, txRaw)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Broadcast Psbt %s, orderId-%s err:%s", entityOrder.Net, entityOrder.OrderId, err.Error()))
	}

	entityOrder.PsbtAskTxId = txPsbtResp.Result
	entityOrder.OrderState = model.OrderStateFinishClaim

	entityOrder.DealTime = tool.MakeTimestamp()
	_, err = mongo_service.SetOrderBrc20Model(entityOrder)
	if err != nil {
		return "", err
	}
	return req.OrderId, nil
}
