package order_brc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/hiro_service"
	"ordbook-aggregation/service/mempool_space_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
)

func PushOrder(req *request.OrderBrc20PushReq) (string, error) {
	var (
		entity *model.OrderBrc20Model
		err error
		orderId string = ""
		psbtBuilder *PsbtBuilder
		sellerAddress string = ""
		buyerAddress string = ""
		coinAmount uint64 = 0
		coinDec int = 18
		outAmount uint64 = 0
		amountDec int = 8
		coinRatePrice uint64 = 0
		inscriptionId string = ""
	)

	if req.OrderState == model.OrderStateCreate {
		psbtBuilder, err = NewPsbtBuilder(&chaincfg.MainNetParams, req.PsbtRaw)
		if err !=  nil  {
			return "", err
		}
		switch req.OrderType {
		case model.OrderTypeSell:
			var (
				inscription *hiro_service.HiroInscription
				has = false
			)
			sellerAddress = req.Address
			coinAmount = req.CoinAmount

			preOutList := psbtBuilder.GetInputs()
			if preOutList == nil || len(preOutList) == 0 {
				return "", errors.New("Wrong Psbt: empty inputs. ")
			}
			for _, v := range preOutList {
				inscriptionId = fmt.Sprintf("%s:%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
				inscription, err = CheckOrdinals(v)
				if err != nil {
					continue
				}
				has = true

			}


			//if !has || inscription == nil {
			//	return "", errors.New("Wrong Psbt: Empty inscription. ")
			//}
			//if inscription.Address != req.Address {
			//	return "", errors.New("Wrong Psbt: Address dose not match. ")
			//}
			//sellerAddress = inscription.Address
			//
			//brc20, err := GetBrc20Data(inscription.Id)
			//if err != nil {
			//	return "", errors.New("Wrong Psbt: Empty brc20. ")
			//}
			//if brc20.Op != model.OP_TRANSFER {
			//	return "", errors.New("Wrong Psbt: Not a transfer op brc20. ")
			//}
			//if brc20.Tick != req.Tick {
			//	return "", errors.New("Wrong Psbt: Tick dose not match. ")
			//}
			//coinAmount, _ = strconv.ParseUint(brc20.Amt, 10, 64)
			_ = inscription
			_ = has

			outList := psbtBuilder.GetOutputs()
			if outList == nil || len(outList) == 0 {
				return "", errors.New("Wrong Psbt: empty outputs. ")
			}
			for _, v := range outList {
				outAmount = uint64(v.Value)
			}

			outAmountDe := decimal.NewFromInt(int64(outAmount))
			coinAmountDe := decimal.NewFromInt(int64(coinAmount))
			coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
			//coinRatePrice, _ = strconv.ParseFloat(coinRatePriceStr, 64)
			coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)

			orderId = fmt.Sprintf("%s_%s_%s_%s_%d_%d", req.Net, req.Tick, inscriptionId, sellerAddress, outAmount, coinAmount)
			orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
			break
		case model.OrderTypeBuy:
			return "", errors.New("Not yet. ")
			break
		default:
			return "", errors.New("Wrong OrderState. ")
		}
	}


	entity = &model.OrderBrc20Model{
		Net:            req.Net,
		OrderId:        orderId,
		Tick:           req.Tick,
		Amount:         outAmount,
		DecimalNum:     amountDec,
		CoinAmount:     coinAmount,
		CoinDecimalNum: coinDec,
		CoinRatePrice:  coinRatePrice,
		OrderState:     req.OrderState,
		OrderType:      req.OrderType,
		SellerAddress:  sellerAddress,
		BuyerAddress:   buyerAddress,
		PsbtRawPreAsk:     req.PsbtRaw,
		Timestamp:tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return "", err
	}
	UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	return "success", nil
}

func FetchOrders(req *request.OrderBrc20FetchReq) (*respond.OrderResponse, error) {
	var (
		entityList []*model.OrderBrc20Model
		list []*respond.Brc20Item
		total int64 = 0
		flag int64 = 0
	)
	total, _ = mongo_service.CountOrderBrc20ModelList(req.Net, req.Tick, req.SellerAddress, req.BuyerAddress, req.OrderType, req.OrderState)
	entityList, _ = mongo_service.FindOrderBrc20ModelList(req.Net, req.Tick, req.SellerAddress, req.BuyerAddress,
		req.OrderType, req.OrderState,
		req.Limit, req.Flag, req.SortKey, req.SortType)
	list = make([]*respond.Brc20Item, len(entityList))
	for k, v := range entityList {
		item := &respond.Brc20Item{
			Net:           v.Net,
			OrderId:           v.OrderId,
			Tick:           v.Tick,
			Amount:         v.Amount,
			DecimalNum:     v.DecimalNum,
			CoinAmount:     v.CoinAmount,
			CoinDecimalNum: v.CoinDecimalNum,
			CoinRatePrice:  v.CoinRatePrice,
			OrderState:     v.OrderState,
			OrderType:      v.OrderType,
			SellerAddress:  v.SellerAddress,
			BuyerAddress:   v.BuyerAddress,
			PsbtRaw:        v.PsbtRawPreAsk,
			Timestamp:      v.Timestamp,
		}
		flag = v.Timestamp
		//list = append(list, item)
		list[k] = item
	}
	return &respond.OrderResponse{
		Total:   total,
		Results: list,
		Flag:    flag,
	}, nil
}




//bid:
//1.Buyer request to Exchange for bid
//2.Exchange make PSBT(X) and signed with  "SIGHASH_SINGLE | ACP"
//3.Buyer Signed with "SIGHASH_ALL | ACP"
//4.Exchange make one input for price difference BTC and waiting for seller
//5.Seller make another PSBT(Y) and signed with "SIGHASH_SINGLE | ACP"
//6.Exchange Signed PSBT(Y) with "SIGHASH_DEFAULT" and broadcast
//7.Exchange add last input for PSBT(X) with "SIGHASH_DEFAULT" and broadcast
func FetchPreBid(req *request.OrderBrc20GetBidReq) (*respond.BidPre, error) {
	var (
		brc20BalanceResult *oklink_service.OklinkBrc20BalanceDetails
		err error
		list []*respond.AvailableItem = make([]*respond.AvailableItem, 0)
	)
	if strings.ToLower(req.Net) != "mainnet" {
		return nil, errors.New("Net not yet. ")
	}
	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(config.PlatformTaprootAddress, req.Tick, 1, 50)
	if err != nil  {
		return nil, err
	}
	for _, v := range brc20BalanceResult.TransferBalanceList {
		list = append(list, &respond.AvailableItem{
			InscriptionId:     v.InscriptionId,
			InscriptionNumber: v.InscriptionNumber,
			CoinAmount:            v.Amount,
		})
	}
	return &respond.BidPre{
		Net:           req.Net,
		Tick:          req.Tick,
		AvailableList: list,
	}, nil
}

func FetchBidPsbt(req *request.OrderBrc20GetBidReq) (*respond.BidPsbt, error) {
	var (
		brc20BalanceResult *oklink_service.OklinkBrc20BalanceDetails
		err error
		bidBalanceItem *oklink_service.BalanceItem
		netParams *chaincfg.Params = &chaincfg.MainNetParams
		inscriptions *oklink_service.OklinkInscriptionDetails
		inscriptionTxId string = ""
		inscriptionTxIndex int64 = 0
		builder *PsbtBuilder
		psbtRaw string = ""
		entityOrder *model.OrderBrc20Model
		orderId string = ""
		coinDec int = 18
		amountDec int = 8
		coinRatePrice uint64 = 0
		inscriptionId string = ""
		inscriptionNumber string = ""
	)
	switch strings.ToLower(req.Net) {
	case "mainnet":
		netParams = &chaincfg.MainNetParams
		break
	case "signet":
		netParams = &chaincfg.SigNetParams
		break
	case "testnet":
		netParams = &chaincfg.TestNet3Params
		break
	}

	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(config.PlatformTaprootAddress, req.Tick, 1, 50)
	if err != nil  {
		return nil, err
	}
	inscriptionId = req.InscriptionId
	inscriptionNumber = req.InscriptionNumber
	for _, v := range brc20BalanceResult.TransferBalanceList {
		if req.InscriptionId == v.InscriptionId &&
			req.InscriptionNumber == v.InscriptionNumber &&
			req.CoinAmount == v.Amount {
			bidBalanceItem = v
			break
		}
	}
	if bidBalanceItem == nil {
		return nil, errors.New("No Available bid. ")
	}
	inscriptions, err = oklink_service.GetInscriptions("", bidBalanceItem.InscriptionId, bidBalanceItem.InscriptionNumber, 1, 50)
	if err != nil  {
		return nil, err
	}
	for _, v := range inscriptions.InscriptionsList {
		if req.InscriptionId == v.InscriptionId &&
			req.InscriptionNumber == v.InscriptionNumber{
			location := v.Location
			locationStrs := strings.Split(location, ":")
			if len(locationStrs) < 3 {
				continue
			}
			inscriptionTxId = locationStrs[0]
			inscriptionTxIndex, _ = strconv.ParseInt(locationStrs[1], 10, 64)
			break
		}
	}

	inputs := make([]Input, 0)
	inputs = append(inputs, Input{
		OutTxId:  inscriptionTxId,
		OutIndex: uint32(inscriptionTxIndex),
	})
	marketPrice := GetMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	if marketPrice <= 0{
		return nil, errors.New("Empty market price. ")
	}

	outputs := make([]Output, 0)
	outputs = append(outputs, Output{
		Address: config.PlatformTaprootAddress2,
		Amount:  marketPrice,
	})
	inputSigns := make([]InputSign, 0)
	txHex, stateCode, err := mempool_space_service.GetTxHex(req.Net, inscriptionTxId)
	if stateCode == 429 {
		return nil, errors.New("Exceed limits. ")
	}
	if err != nil {
		return nil, err
	}

	inputSigns = append(inputSigns, InputSign{
		Index:       0,
		OutRaw:      txHex,
		SighashType: txscript.SigHashSingle | txscript.SigHashAnyOneCanPay,
		PriHex:      config.PlatformPrivateKey,
	})

	builder, err = CreatePsbtBuilder(netParams, inputs, outputs)
	if err != nil {
		return nil, err
	}
	err = builder.UpdateAndSignInput(inputSigns)
	if err != nil {
		return nil, err
	}
	psbtRaw, err = builder.ToString()
	if err != nil {
		return nil, err
	}
	coinAmountInt, _ := strconv.ParseUint(req.CoinAmount, 10, 64)

	//save
	outAmountDe := decimal.NewFromInt(int64(req.Amount))
	coinAmountDe := decimal.NewFromInt(int64(coinAmountInt))
	coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
	coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)
	orderId = fmt.Sprintf("%s_%s_%s_%s_%d_%d", req.Net, req.Tick, inscriptionId, req.Address, req.Amount, coinAmountInt)
	orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
	entityOrder = &model.OrderBrc20Model{
		Net:               req.Net,
		OrderId:           orderId,
		Tick:              req.Tick,
		Amount:            req.Amount,
		DecimalNum:        amountDec,
		CoinAmount:        coinAmountInt,
		CoinDecimalNum:    coinDec,
		CoinRatePrice:     coinRatePrice,
		OrderState:        model.OrderStatePreCreate,
		OrderType:         model.OrderTypeBuy,
		SellerAddress:     "",
		BuyerAddress:      req.Address,
		MarketAmount:      marketPrice,
		PlatformTx:        inscriptionTxId,
		InscriptionId:     inscriptionId,
		InscriptionNumber: inscriptionNumber,
		PsbtRawPreAsk:     "",
		PsbtRawPreBid:     psbtRaw,
		Timestamp:         tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetOrderBrc20Model(entityOrder)
	if err != nil {
		return nil, err
	}

	return &respond.BidPsbt{
		Net:     req.Net,
		Tick:    req.Tick,
		PsbtRaw: psbtRaw,
	}, nil
}

func UpdateBidPsbt(req *request.OrderBrc20UpdateBidReq) (string, error) {
	var (
		entityOrder *model.OrderBrc20Model
		err error
		psbtBuilder *PsbtBuilder
		netParams *chaincfg.Params = &chaincfg.MainNetParams
	)
	switch strings.ToLower(req.Net) {
	case "mainnet":
		netParams = &chaincfg.MainNetParams
		break
	case "signet":
		netParams = &chaincfg.SigNetParams
		break
	case "testnet":
		netParams = &chaincfg.TestNet3Params
		break
	}

	entityOrder, _ = mongo_service.FindOrderBrc20ModelByOrderId(req.OrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("")
	}
	if entityOrder.OrderType != model.OrderTypeBuy {
		return "", errors.New("Order not a bid. ")
	}

	psbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
	if err !=  nil  {
		return "", err
	}
	preOutList := psbtBuilder.GetInputs()
	if preOutList == nil || len(preOutList) == 0 {
		return "", errors.New("Wrong Psbt: empty inputs. ")
	}
	if len(preOutList) != 4 {
		return "", errors.New("Wrong Psbt: No match inputs. ")
	}
	exchangeInput := preOutList[2]
	if exchangeInput.PreviousOutPoint.Hash.String() != entityOrder.PlatformTx {
		return "", errors.New("Wrong Psbt: No inscription input. ")
	}
	buyerInput := preOutList[3]
	buyerInputTxId := buyerInput.PreviousOutPoint.Hash.String()
	buyerInputIndex := buyerInput.PreviousOutPoint.Index
	buyerTx, err := oklink_service.GetTxDetail(buyerInputTxId)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Get Buyer preTx err:%s", err.Error()))
	}
	buyerAmount, _ := strconv.ParseUint(buyerTx.OutputDetails[buyerInputIndex].Amount, 10, 64)
	if  buyerAmount <= req.Amount {
		return "", errors.New("Wrong Psbt: The value of buyer input dose not match. ")
	}

	outList := psbtBuilder.GetOutputs()
	if len(preOutList) != 6 && len(preOutList) != 7 {
		return "", errors.New("Wrong Psbt: No match outputs. ")
	}
	exchangeOut := outList[2]
	if uint64(exchangeOut.Value) != entityOrder.MarketAmount {
		return "", errors.New("Wrong Psbt: wrong value of out for exchange. ")
	}
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(exchangeOut.PkScript, netParams)
	if err != nil {
		return "", errors.New("Wrong Psbt: Extract address from out for exchange. ")
	}
	if addrs[0].EncodeAddress() != config.PlatformTaprootAddress2 {
		return "", errors.New("Wrong Psbt: wrong address of out for exchange. ")
	}

	entityOrder.OrderState = model.OrderStateCreate
	entityOrder.PsbtRawMidBid = req.PsbtRaw
	_, err = mongo_service.SetOrderBrc20Model(entityOrder)
	if err != nil {
		return "", err
	}
	UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

	return req.OrderId, nil
}

func DoBid(req *request.OrderBrc20DoBidReq) (string, error) {
	var (
		entity *model.OrderBrc20Model
		err error
		psbtBuilder *PsbtBuilder
		netParams *chaincfg.Params = &chaincfg.MainNetParams
		utxoDummyList []*model.OrderUtxoModel
		utxoBidYList []*model.OrderUtxoModel

		startIndexDummy int64 = 0
		startIndexBidY int64 = 0
	)
	entity, _ = mongo_service.FindOrderBrc20ModelByOrderId(req.OrderId)
	if entity == nil {
		return "", errors.New("Bid is empty. ")
	}

	entity.PsbtRawPreAsk = req.PsbtRaw

	psbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
	if err !=  nil  {
		return "", err
	}
	preOutList := psbtBuilder.GetInputs()
	if preOutList == nil || len(preOutList) == 0 {
		return "", errors.New("Wrong Psbt: empty inputs. ")
	}

	utxoDummyList, _ = mongo_service.FindUtxoList(req.Net, startIndexDummy, 2, model.UtxoTypeDummy)
	if len(utxoDummyList) == 0 {
		return "", errors.New("Service Upgrade for dummy. ")
	}
	utxoBidYList, _ = mongo_service.FindUtxoList(req.Net, startIndexBidY, 5, model.UtxoTypeBidY)
	if len(utxoBidYList) == 0 {
		return "", errors.New("Service Upgrade for bid. ")
	}


	return "", nil
}

func UpdateOrder(req *request.OrderBrc20UpdateReq) (string, error) {
	var (
		entityOrder *model.OrderBrc20Model
		err error
	)
	entityOrder, _ = mongo_service.FindOrderBrc20ModelByOrderId(req.OrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("Order is empty. ")
	}

	if req.OrderState == model.OrderStateFinish || req.OrderState == model.OrderStateCancel {
		entityOrder.OrderState = req.OrderState
		switch entityOrder.OrderType {
		case model.OrderTypeSell:
			entityOrder.PsbtRawFinalAsk = req.PsbtRaw
			break
		case model.OrderTypeBuy:
			entityOrder.PsbtRawFinalBid = req.PsbtRaw
			break
		}
		_, err = mongo_service.SetOrderBrc20Model(entityOrder)
		if err != nil {
			return "", err
		}
	}else {
		return "", errors.New("Wrong state. ")
	}


	return req.OrderId, nil
}