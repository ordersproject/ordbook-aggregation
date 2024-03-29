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
	claimDayLimit   int64 = 2
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

	if req.Address == "bc1pxeyh7t7jsjy8cp82uyktluswrjks857g9p5jp9p3gznhh4l43vasxk73yh" ||
		req.Address == "bc1prqtv8aep7ucyxkvf7d6ysvjcaqt97w8rhgujdqfggsa75s038xssu0plp8" ||
		req.Address == "bc1pnjnls650g6jsfcz9khfe6whrgz554cr3qce6mtm3w9fm98yhad7q3gg548" ||
		req.Address == "bc1pcn5jrkj685js2drekqhfy3y7asty9l3gy6eqprk77ek5rh4vmftqwtqlsa" ||
		req.Address == "bc1qpdut0l6x4talcmrea0vy0dy3f8n6du9vkljnrt" ||
		req.Address == "bc1pwn878nk8fxkqtw5r3kwqftam3qdhu5m4mngyv8wax0jua9jhymwsyjkph2" ||
		req.Address == "bc1pt37lx4xls62l8fx79pk3tsk0xm4f94tzj3gccjm69h0u5ppct7sqzc0ccl" ||
		req.Address == "bc1ptf0n3jes6zv8zm6ttz020pnqvvx7pxq3grsx8tgt52wj9lnfruvqg7seaw" {

	} else {
		return nil, errors.New("The event has ended, thank you for participating. ")
	}

	claimCoinAmount, canCount, _ := getWhitelistCount(req.Net, req.Tick, req.Address, ip, model.WhitelistTypeClaim)
	if canCount <= 0 {
		claimCoinAmount, canCount, err = getWhitelistCount(req.Net, req.Tick, req.Address, ip, model.WhitelistTypeClaim1w)
		if err != nil {
			return nil, err
		}
		if canCount <= 0 {
			return nil, errors.New("already had claimed")
		}
	}

	entity, err = GetUnoccupiedClaimBrc20PsbtList(req.Net, req.Tick, claimFetchLimit, claimCoinAmount)
	if err != nil {
		return nil, err
	}
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
		Net:            entity.Net,
		OrderId:        entity.OrderId,
		Tick:           entity.Tick,
		Fee:            entity.Amount,
		CoinAmount:     entity.CoinAmount,
		PsbtRaw:        entity.PsbtRawPreAsk,
		AvailableCount: canCount,
	}
	return item, nil
}

func UpdateClaimOrder(req *request.OrderBrc20ClaimUpdateReq, buyerOwnAddress, publicKey, ip string) (string, error) {
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
		if buyerOwnAddress != "" {
			buyerAddress = buyerOwnAddress
		} else {
			buyerTx, err := oklink_service.GetTxDetail(buyerInputTxId)
			if err != nil {
				return "", errors.New(fmt.Sprintf("Get Buyer preTx err:%s", err.Error()))
			}
			buyerAddress = buyerTx.OutputDetails[buyerInputIndex].OutputHash
		}
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

	entityOrder.PsbtRawFinalAsk = req.PsbtRaw
	entityOrder.PsbtAskTxId = txPsbtResp.Result
	entityOrder.OrderState = model.OrderStateFinishClaim

	entityOrder.DealTime = tool.MakeTimestamp()
	_, err = mongo_service.SetOrderBrc20Model(entityOrder)
	if err != nil {
		return "", err
	}
	updateWhiteListUsed(entityOrder.BuyerAddress, ip, model.WhitelistTypeClaim)
	return req.OrderId, nil
}
