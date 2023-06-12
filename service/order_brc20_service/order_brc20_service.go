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
	"github.com/shopspring/decimal"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/create_key"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
	"time"
)

const (
	inSize uint64 = 180
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
				inscriptionBrc20BalanceItem *oklink_service.BalanceItem
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
				inscriptionBrc20BalanceItem, err = CheckBrc20Ordinals(v, req.Tick, sellerAddress)
				if err != nil {
					continue
				}
				has = true
			}

			if req.Net == "mainnet"|| req.Net == "livenet" {
				if !has || inscriptionBrc20BalanceItem == nil {
					return "", errors.New("Wrong Psbt: Empty inscription. ")
				}
				coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
			}

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
		_, platformAddressSendBrc20 string = GetPlatformKeyAndAddressSendBrc20(req.Net)
	)
	//if strings.ToLower(req.Net) != "mainnet" && strings.ToLower(req.Net) != "livenet" {
	//	return nil, errors.New("Net not yet. ")
	//}
	if strings.ToLower(req.Net) == "testnet" {
		utxoFakerBrc20 := GetTestFakerInscription(req.Net)
		for _, v := range utxoFakerBrc20 {
			inscriptionId := fmt.Sprintf("%si%d", v.TxId, v.Index)
			if CheckBidInscriptionIdExist(inscriptionId) {
				continue
			}
			list = append(list, &respond.AvailableItem{
				InscriptionId:     inscriptionId,
				InscriptionNumber: fmt.Sprintf("test%d", v.Index),
				CoinAmount:        "120",
			})
		}

	}else {
		brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressSendBrc20, req.Tick, 1, 50)
		if err != nil  {
			return nil, err
		}
		for _, v := range brc20BalanceResult.TransferBalanceList {
			if CheckBidInscriptionIdExist(v.InscriptionId) {
				continue
			}
			list = append(list, &respond.AvailableItem{
				InscriptionId:     v.InscriptionId,
				InscriptionNumber: v.InscriptionNumber,
				CoinAmount:            v.Amount,
			})
		}
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
		netParams *chaincfg.Params = GetNetParams(req.Net)
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
		platformPrivateKeyReceiveBidValue, platformAddressReceiveBidValue string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		platformPrivateKeySendBrc20, platformAddressSendBrc20 string = GetPlatformKeyAndAddressSendBrc20(req.Net)
		inputSignsExchangePriHex string = platformPrivateKeySendBrc20
		inputSignsExchangePkScript string = ""
		inputSignsUtxoType UtxoType = Witness
		inputSignsTxHex string = ""
		inputSignsAmount uint64 = 0
	)
	_ = platformPrivateKeyReceiveBidValue

	if strings.ToLower(req.Net) == "testnet" {
		inscriptionId = req.InscriptionId
		inscriptionNumber = req.InscriptionNumber

		utxoFakerBrc20 := GetTestFakerInscription(req.Net)
		for _, v := range utxoFakerBrc20 {
			fakerBrc20 := fmt.Sprintf("%si%d", v.TxId, v.Index)
			if req.InscriptionId == fakerBrc20 {
				bidBalanceItem = &oklink_service.BalanceItem{
					InscriptionId:     fmt.Sprintf("%si%d", v.TxId, v.Index),
					InscriptionNumber: fmt.Sprintf("test%d", v.Index),
					Amount:        "120",
				}
				inscriptionTxId = v.TxId
				inscriptionTxIndex = v.Index
				inputSignsExchangePriHex = v.PrivateKeyHex
				inputSignsExchangePkScript = v.PkScript
				inputSignsUtxoType = Witness
				inputSignsAmount = v.Amount
				req.CoinAmount = bidBalanceItem.Amount
				break
			}
		}
	}else {
		brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressSendBrc20, req.Tick, 1, 50)
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

		//txHex, stateCode, err := mempool_space_service.GetTxHex(req.Net, inscriptionTxId)
		//if stateCode == 429 {
		//	return nil, errors.New("Exceed limits. ")
		//}
		//if err != nil {
		//	return nil, err
		//}
		//inputSignsTxHex = txHex
		preTx, err := oklink_service.GetTxDetail(inscriptionTxId)
		if err != nil {
			return nil, err
		}
		preTxOut := preTx.OutputDetails[inscriptionTxIndex]
		preTxOutAmountDe, err := decimal.NewFromString(preTxOut.Amount)
		if err != nil {
			return nil, errors.New("The value of platform brc input decimal parse err. ")
		}
		inputSignsAmount = uint64(preTxOutAmountDe.Mul(decimal.New(1, 8)).IntPart())

		inputSignsExchangePriHex = platformPrivateKeySendBrc20
		addr, err := btcutil.DecodeAddress(platformAddressSendBrc20, netParams)
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}

		inputSignsExchangePkScript = hex.EncodeToString(pkScript)
		inputSignsUtxoType = Witness
	}


	inputs := make([]Input, 0)
	inputs = append(inputs, Input{
		OutTxId:  inscriptionTxId,
		OutIndex: uint32(inscriptionTxIndex),
	})
	marketPrice := GetMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	if marketPrice <= 0 {
		return nil, errors.New("Empty Market. ")
	}
	coinAmountInt, _ := strconv.ParseUint(req.CoinAmount, 10, 64)
	fmt.Printf("marketPrice:%d\n", marketPrice)
	marketPrice = marketPrice * coinAmountInt
	fmt.Printf("finalMarketPrice:%d\n", marketPrice)
	fmt.Printf("coinAmountInt:%d\n", coinAmountInt)

	//todo check marketPrice >= req.amount

	outputs := make([]Output, 0)
	outputs = append(outputs, Output{
		Address: platformAddressReceiveBidValue,
		Amount:  marketPrice,
	})
	inputSigns := make([]*InputSign, 0)


	inputSigns = append(inputSigns, &InputSign{
		Index:       0,
		OutRaw:      inputSignsTxHex,
		PkScript:    inputSignsExchangePkScript,
		SighashType: txscript.SigHashSingle | txscript.SigHashAnyOneCanPay,
		PriHex:      inputSignsExchangePriHex,
		UtxoType:    inputSignsUtxoType,
		Amount:      inputSignsAmount,
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
		OrderId: entityOrder.OrderId,
		PsbtRaw: psbtRaw,
	}, nil
}

func UpdateBidPsbt(req *request.OrderBrc20UpdateBidReq) (string, error) {
	var (
		entityOrder *model.OrderBrc20Model
		err error
		psbtBuilder *PsbtBuilder
		netParams *chaincfg.Params = GetNetParams(req.Net)
		coinRatePrice uint64 = 0
		buyerAddress string = ""
		_, platformAddressReceiveBidValue string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
	)
	entityOrder, _ = mongo_service.FindOrderBrc20ModelByOrderId(req.OrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("order is empty. ")
	}
	if entityOrder.OrderType != model.OrderTypeBuy {
		return "", errors.New("Order not a bid. ")
	}

	psbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
	if err !=  nil  {
		return "", errors.New("Wrong Psbt: Parse err. ")
	}
	preOutList := psbtBuilder.GetInputs()
	if preOutList == nil || len(preOutList) == 0 {
		return "", errors.New("Wrong Psbt: empty inputs. ")
	}
	if len(preOutList) < 4 {
		return "", errors.New("Wrong Psbt: No match inputs length. ")
	}

	//check platform brc20 utxo
	exchangeInput := preOutList[2]
	if exchangeInput.PreviousOutPoint.Hash.String() != entityOrder.PlatformTx {
		return "", errors.New("Wrong Psbt: No inscription input. ")
	}
	//check buyer pay utxo //todo 2+ utxo
	buyerInput := preOutList[3]
	buyerInputTxId := buyerInput.PreviousOutPoint.Hash.String()
	buyerInputIndex := buyerInput.PreviousOutPoint.Index
	buyerInAmount := uint64(0)
	if strings.ToLower(req.Net) != "testnet" {
		buyerTx, err := oklink_service.GetTxDetail(buyerInputTxId)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Get Buyer preTx err:%s", err.Error()))
		}
		//buyerInAmount, _ = strconv.ParseUint(buyerTx.OutputDetails[buyerInputIndex].Amount, 10, 64)
		buyerAmountDe, err := decimal.NewFromString(buyerTx.OutputDetails[buyerInputIndex].Amount)
		if err != nil {
			return "", errors.New("Wrong Psbt: The value of buyer input decimal parse err. ")
		}
		buyerInAmount = uint64(buyerAmountDe.Mul(decimal.New(1, 8)).IntPart())
		fmt.Printf("buyerInputIndex:%d, buyerInAmount:%d, req.Amount:%d\n", buyerInputIndex, buyerInAmount, req.Amount)
		if  buyerInAmount <= req.Amount {
			return "", errors.New("Wrong Psbt: The value of buyer input dose not match. ")
		}
		buyerAddress = buyerTx.OutputDetails[buyerInputIndex].OutputHash
	}else {
		buyerInAmount = req.BuyerInValue
		buyerAddress = req.Address
	}


	//check out: len-6 for no buyer changeWallet
	outList := psbtBuilder.GetOutputs()
	if len(outList) != 6 && len(outList) != 7 {
		return "", errors.New("Wrong Psbt: No match outputs. ")
	}

	//check out for platform receive marketPrice amount and address
	exchangeOut := outList[2]
	if uint64(exchangeOut.Value) != entityOrder.MarketAmount {
		return "", errors.New("Wrong Psbt: wrong value of out for exchange. ")
	}
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(exchangeOut.PkScript, netParams)
	if err != nil {
		return "", errors.New("Wrong Psbt: Extract address from out for exchange. ")
	}
	if addrs[0].EncodeAddress() != platformAddressReceiveBidValue {
		return "", errors.New("Wrong Psbt: wrong address of out for exchange. ")
	}

	platformFeeOut := outList[3]
	if uint64(platformFeeOut.Value) == 0 {
		return "", errors.New("Wrong Psbt: wrong value of platform fee. ")
	}
	platformFee := uint64(platformFeeOut.Value)
	changeAmount := uint64(0)
	if len(outList) == 7 {
		changeOut := outList[6]
		_, changeAddrs, _, err := txscript.ExtractPkScriptAddrs(changeOut.PkScript, netParams)
		if err != nil {
			return "", errors.New("Wrong Psbt: Extract address from out for buyer changeWallet. ")
		}
		if changeAddrs[0].EncodeAddress() != buyerAddress {
			return "", errors.New("Wrong Psbt: wrong address of out for buyer changeWallet. ")
		}
		changeAmount = uint64(changeOut.Value)
	}

	//check amount
	buyAmount := int64(buyerInAmount) - int64(platformFee) - int64(changeAmount) - int64(req.Fee) - (600*2)
	fmt.Printf("buyerInAmount:%d, platformFee:%d, changeAmount:%d, req.Fee:%d\n", buyerInAmount, platformFee, changeAmount, req.Fee)
	fmt.Printf("buyAmount:%d, req.Amount:%d\n", buyAmount, req.Amount)
	if buyAmount <= 0 {
		return "", errors.New("Wrong Psbt: The purchase amount is less than 0. ")
	}
	if buyAmount != int64(req.Amount) {
		return "", errors.New("Wrong Psbt: The purchase amount dose not match. ")
	}
	supplementaryAmount := entityOrder.MarketAmount - uint64(buyAmount)
	if supplementaryAmount < 600 {
		return "", errors.New("Wrong Psbt: The purchase amount exceeds the market price. ")
	}


	for i := 0; i < 2; i++ {
		dummy := preOutList[i]
		SaveForUserBidDummy(entityOrder.Net, entityOrder.Tick, entityOrder.BuyerAddress, entityOrder.OrderId, dummy.PreviousOutPoint.Hash.String(), int64(dummy.PreviousOutPoint.Index), model.DummyStateLive)
	}


	outAmountDe := decimal.NewFromInt(int64(req.Amount))
	coinAmountDe := decimal.NewFromInt(int64(entityOrder.CoinAmount))
	coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
	coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)


	entityOrder.PlatformFee = platformFee
	entityOrder.ChangeAmount = changeAmount
	entityOrder.Fee = req.Fee
	entityOrder.FeeRate = req.Rate
	entityOrder.SupplementaryAmount = supplementaryAmount
	entityOrder.CoinRatePrice = coinRatePrice
	entityOrder.BuyerAddress = buyerAddress
	entityOrder.Amount = req.Amount
	entityOrder.OrderState = model.OrderStateCreate
	entityOrder.PsbtRawMidBid = req.PsbtRaw
	_, err = mongo_service.SetOrderBrc20Model(entityOrder)
	if err != nil {
		return "", err
	}
	UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

	return req.OrderId, nil
}

func DoBid(req *request.OrderBrc20DoBidReq) (*respond.DoBidResp, error) {
	var (
		entity *model.OrderBrc20Model
		err error
		psbtBuilder *PsbtBuilder
		netParams *chaincfg.Params = GetNetParams(req.Net)
		utxoDummyList []*model.OrderUtxoModel
		utxoBidYList []*model.OrderUtxoModel

		//startIndexDummy int64 = -1
		//startIndexBidY int64 = -1
		newPsbtBuilder *PsbtBuilder
		marketPrice uint64 = 0
		inscriptionId string = ""
		inscriptionBrc20BalanceItem *oklink_service.BalanceItem
		coinAmount uint64 = 0
		brc20ReceiveValue uint64 = 0
		platformPrivateKeyReceiveBidValue, platformAddressReceiveBidValue string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		_, platformAddressReceiveBrc20 string = GetPlatformKeyAndAddressReceiveBrc20(req.Net)
		_, platformAddressReceiveDummyValue string = GetPlatformKeyAndAddressReceiveDummyValue(req.Net)
		_, platformAddressSendBrc20 string = GetPlatformKeyAndAddressSendBrc20(req.Net)

		sellerSendAddress string = req.Address
		inValue uint64 = req.Value

		platformPayPerAmount uint64 = 10000
	)
	entity, _ = mongo_service.FindOrderBrc20ModelByOrderId(req.OrderId)
	if entity == nil {
		return nil, errors.New("Bid is empty. ")
	}
	newDummyOutPriKeyHex, newDummyOutSegwitAddress, err := create_key.CreateSegwitKey(netParams)
	if err != nil {
		return nil, err
	}

	entity.PsbtRawPreAsk = req.PsbtRaw

	psbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
	if err !=  nil  {
		return nil, err
	}
	preOutList := psbtBuilder.GetInputs()
	if preOutList == nil || len(preOutList) == 0 {
		return nil, errors.New("Wrong Psbt: empty inputs. ")
	}
	sellOuts := psbtBuilder.GetOutputs()
	if sellOuts == nil || len(sellOuts) == 0 {
		return nil, errors.New("Wrong Psbt: empty outputs. ")
	}
	if len(preOutList) != 1 || len(sellOuts) != 1 {
		return nil, errors.New("Wrong Psbt: wrong length of inputs or length of outputs. ")
	}
	if strings.ToLower(req.Net) != "testnet" {
		preSellBrc20Tx, err := oklink_service.GetTxDetail(preOutList[0].PreviousOutPoint.Hash.String())
		if err != nil {
			return nil, errors.New("Wrong Psbt: brc20 input is empty preTx. ")
		}
		//inValue, _ = strconv.ParseUint(preSellBrc20Tx.OutputDetails[preOutList[0].PreviousOutPoint.Index].Amount, 10, 64)
		inValueDe, err := decimal.NewFromString(preSellBrc20Tx.OutputDetails[preOutList[0].PreviousOutPoint.Index].Amount)
		if err != nil {
			return nil, errors.New("Wrong Psbt: The value of brc20 input decimal parse err. ")
		}
		inValue = uint64(inValueDe.Mul(decimal.New(1, 8)).IntPart())
		if inValue == 0 {
			return nil, errors.New("Wrong Psbt: brc20 out of preTx is empty amount. ")
		}
		sellerSendAddress = preSellBrc20Tx.OutputDetails[preOutList[0].PreviousOutPoint.Index].OutputHash
	}

	sellerReceiveValue := uint64(sellOuts[0].Value)
	_, addrs, _, err := txscript.ExtractPkScriptAddrs(sellOuts[0].PkScript, netParams)
	if err != nil {
		return nil, errors.New("Wrong Psbt: Extract address from out for Seller. ")
	}
	sellerReceiveAddress := addrs[0].EncodeAddress()
	if sellerReceiveValue != entity.Amount {
		return nil, errors.New("Wrong Psbt: Seller receive value dose not match. ")
	}

	has := false
	for _, v := range preOutList {
		inscriptionId = fmt.Sprintf("%s:%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
		inscriptionBrc20BalanceItem, err = CheckBrc20Ordinals(v, entity.Tick, sellerSendAddress)
		if err != nil {
			continue
		}
		has = true
	}
	_ = inscriptionId

	if req.Net == "mainnet"|| req.Net == "livenet" {
		if !has || inscriptionBrc20BalanceItem == nil {
			return nil, errors.New("Wrong Psbt: Empty inscription. ")
		}
		coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
	}else {
		coinAmount, _ = strconv.ParseUint(req.CoinAmount, 10, 64)
	}
	if coinAmount != entity.CoinAmount {
		return nil, errors.New("Wrong Psbt: brc20 coin amount dose not match. ")
	}

	brc20ReceiveValue = inValue



	//utxoDummyList, _ = mongo_service.FindUtxoList(req.Net, startIndexDummy, 2, model.UtxoTypeDummy)
	//if len(utxoDummyList) == 0 {
	//	return nil, errors.New("Service Upgrade for dummy. ")
	//}
	utxoDummyList, err = GetUnoccupiedUtxoList(req.Net, 2, model.UtxoTypeDummy)
	defer ReleaseUtxoList(utxoDummyList)
	if err != nil {
		return nil, err
	}



	//get bidY pay utxo
	limit := (entity.SupplementaryAmount + sellerReceiveValue + entity.Fee)/platformPayPerAmount + 1

	//utxoBidYList, _ = mongo_service.FindUtxoList(req.Net, startIndexBidY, int64(limit), model.UtxoTypeBidY)
	//if len(utxoBidYList) == 0 {
	//	return nil, errors.New("Service Upgrade for bid. ")
	//}
	//if uint64(len(utxoBidYList)) != limit {
	//	return nil, errors.New("Service Upgrade for bid. ")
	//}

	utxoBidYList, err = GetUnoccupiedUtxoList(req.Net, int64(limit), model.UtxoTypeBidY)
	defer ReleaseUtxoList(utxoDummyList)
	if err != nil {
		return nil, err
	}




	inputs := make([]Input, 0)
	outputs := make([]Output, 0)
	dummyOutValue := uint64(0)
	//add dummy ins - index: 0,1
	for _, dummy := range utxoDummyList {
		inputs = append(inputs, Input{
			OutTxId:  dummy.TxId,
			OutIndex: uint32(dummy.Index),
		})
		dummyOutValue = dummyOutValue + dummy.Amount
	}
	//add seller brc20 ins - index: 2
	inputs = append(inputs, Input{
		OutTxId:  preOutList[0].PreviousOutPoint.Hash.String(),
		OutIndex: preOutList[0].PreviousOutPoint.Index,
	})
	//add Exchange pay value ins - index: 3,3+
	for _, payBid := range utxoBidYList {
		inputs = append(inputs, Input{
			OutTxId:  payBid.TxId,
			OutIndex: uint32(payBid.Index),
		})
	}


	//add dummy outs - idnex: 0
	outputs = append(outputs, Output{
		Address: platformAddressReceiveDummyValue,
		Amount:  dummyOutValue,
	})
	//add receive brc20 outs - idnex: 1
	outputs = append(outputs, Output{
		Address: platformAddressReceiveBrc20,
		Amount:  brc20ReceiveValue,
	})
	//add receive seller outs - idnex: 2
	outputs = append(outputs, Output{
		Address: sellerReceiveAddress,
		Amount:  sellerReceiveValue,
	})
	_ = marketPrice
	//add receive exchange psbtX outs - idnex: 3
	psbtXValue := entity.MarketAmount - sellerReceiveValue
	exchangePsbtXOut :=  Output{
		Address: platformAddressReceiveBidValue,
		Amount:  psbtXValue,
	}
	outputs = append(outputs, exchangePsbtXOut)
	//add new dummy outs - idnex: 4,5
	newDummyOut := Output{
		Address: newDummyOutSegwitAddress,
		Amount:  600,
	}
	outputs = append(outputs, newDummyOut)
	outputs = append(outputs, newDummyOut)

	//finish PSBT(Y)
	newPsbtBuilder, err = CreatePsbtBuilder(netParams, inputs, outputs)
	if err !=  nil  {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): newPsbtBuilder err:%s",err.Error()))
	}

	finalScriptWitness := psbtBuilder.PsbtUpdater.Upsbt.Inputs[0].FinalScriptWitness
	witnessUtxo := psbtBuilder.PsbtUpdater.Upsbt.Inputs[0].WitnessUtxo
	sighashType := psbtBuilder.PsbtUpdater.Upsbt.Inputs[0].SighashType
	err = newPsbtBuilder.AddSigIn(witnessUtxo, sighashType, finalScriptWitness, 2)
	if err !=  nil  {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): AddPartialSigIn err:%s",err.Error()))
	}

	inSigns := make([]*InputSign, 0)
	//add dummy ins sign - index: 0,1
	for k, dummy := range utxoDummyList {
		inSigns = append(inSigns, &InputSign{
			Index:       k,
			PkScript:    dummy.PkScript,
			Amount:      dummy.Amount,
			SighashType: txscript.SigHashAll,
			PriHex:      dummy.PrivateKeyHex,
			UtxoType:    Witness,
		})
	}
	//add Exchange pay value ins - index: 3,3+
	for k, payBid := range utxoBidYList {
		inSigns = append(inSigns, &InputSign{
			Index:       k+3,
			PkScript:    payBid.PkScript,
			Amount:      payBid.Amount,
			SighashType: txscript.SigHashAll,
			PriHex:      payBid.PrivateKeyHex,
			UtxoType:    Witness,
		})
	}
	err = newPsbtBuilder.UpdateAndSignInput(inSigns)
	if err !=  nil  {
		return nil,  errors.New(fmt.Sprintf("PSBT(Y): UpdateAndSignInput err:%s",err.Error()))
	}
	psbtRawFinalAsk, err := newPsbtBuilder.ToString()
	if err !=  nil  {
		return nil,  errors.New(fmt.Sprintf("PSBT(Y): ToString err:%s",err.Error()))
	}
	entity.PsbtRawFinalAsk = psbtRawFinalAsk

	txRawPsbtY, err := newPsbtBuilder.ExtractPsbtTransaction()
	if err !=  nil  {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): ExtractPsbtTransaction err:%s",err.Error()))
	}
	txRawPsbtYByte, _ := hex.DecodeString(txRawPsbtY)

	txPsbtY := wire.NewMsgTx(2)
	err = txPsbtY.Deserialize(bytes.NewReader(txRawPsbtYByte))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): txRawPsbtY Deserialize err:%s",err.Error()))
	}

	psbtYTxId := txPsbtY.TxHash().String()




	//finish PSBT(X)
	bidPsbtBuilder, err := NewPsbtBuilder(netParams, entity.PsbtRawMidBid)
	if err !=  nil  {
		return nil, errors.New(fmt.Sprintf("PSBT(X): NewPsbtBuilder err:%s",err.Error()))
	}
	//Check PSBT(X) utxo valid
	insPsbtX := bidPsbtBuilder.GetInputs()
	addressUtxoMap := make(map[string][]*wire.TxIn)
	addressUtxoMap[entity.BuyerAddress] = make([]*wire.TxIn, 0)
	addressUtxoMap[platformAddressSendBrc20] = make([]*wire.TxIn, 0)
	for k, v := range insPsbtX {
		if k == 2 {
			addressUtxoMap[platformAddressSendBrc20] = append(addressUtxoMap[platformAddressSendBrc20], v)
		}else {
			addressUtxoMap[entity.BuyerAddress] = append(addressUtxoMap[entity.BuyerAddress], v)
		}
	}
	for address, list := range addressUtxoMap{
		utxoResp, err := oklink_service.GetAddressUtxo(address, 1, 50)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("PSBT(Y): Recheck address utxo list err:%s", err.Error()))
		}
		for _, bidIn := range list {
			bidInId := fmt.Sprintf("%s_%d", bidIn.PreviousOutPoint.Hash.String(), bidIn.PreviousOutPoint.Index)
			has := false
			for _, u := range utxoResp.UtxoList {
				uId := fmt.Sprintf("%s_%s", u.TxId, u.Index)
				if bidInId == uId {
					has = true
					break
				}
			}
			if !has {
				entity.OrderState = model.OrderStateErr
				_, err := mongo_service.SetOrderBrc20Model(entity)
				if err != nil {
					return nil, err
				}
				return nil, errors.New(fmt.Sprintf("PSBT(Y): Recheck address utxo list, utxo had been spent: %s", bidInId))
			}
		}
		time.Sleep(500 * time.Millisecond)
	}



	bidIn := Input{
		OutTxId:  psbtYTxId,
		OutIndex: 3,
	}
	bidInSign := &InputSign{
		UtxoType:    Witness,
		Index:       len(bidPsbtBuilder.GetInputs()),
		OutRaw:      txRawPsbtY,
		PkScript:    hex.EncodeToString(txPsbtY.TxOut[bidIn.OutIndex].PkScript),
		Amount:      psbtXValue,
		SighashType: txscript.SigHashAll|txscript.SigHashAnyOneCanPay,
		PriHex:      platformPrivateKeyReceiveBidValue,
	}
	err = bidPsbtBuilder.AddInput(bidIn, bidInSign)
	if err !=  nil  {
		return nil, errors.New(fmt.Sprintf("PSBT(X): AddInput err:%s",err.Error()))
	}
	psbtRawFinalBid, err := bidPsbtBuilder.ToString()
	if err !=  nil  {
		return nil, errors.New(fmt.Sprintf("PSBT(X): ToString err:%s",err.Error()))
	}
	entity.PsbtRawFinalBid = psbtRawFinalBid

	txRawPsbtX, err := bidPsbtBuilder.ExtractPsbtTransaction()
	if err !=  nil  {
		return nil, errors.New(fmt.Sprintf("PSBT(X): ExtractPsbtTransaction err:%s",err.Error()))
	}
	txRawPsbtXByte, _ := hex.DecodeString(txRawPsbtX)

	txPsbtX := wire.NewMsgTx(2)
	err = txPsbtX.Deserialize(bytes.NewReader(txRawPsbtXByte))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): txRawPsbtY Deserialize err:%s",err.Error()))
	}

	psbtXTxId := txPsbtX.TxHash().String()

	entity.PsbtAskTxId = psbtYTxId
	entity.PsbtBidTxId = psbtXTxId
	entity.SellerAddress = sellerSendAddress
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return nil, err
	}

	fmt.Printf("PsbtY:%s\n", txRawPsbtY)
	fmt.Printf("psbtYTxId: %s\n", psbtYTxId)
	fmt.Printf("PsbtX:%s\n", txRawPsbtX)
	fmt.Printf("psbtXTxId: %s\n", psbtXTxId)

	txPsbtXRespTxId := ""
	txPsbtYRespTxId := ""
	if req.Net == "mainnet" || req.Net == "livenet"  {
		txPsbtYResp, err := oklink_service.BroadcastTx(txRawPsbtY)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s", req.Net, err.Error()))
		}
		setUsedDummyUtxo(utxoDummyList, txPsbtYResp.TxId)
		setUsedBidYUtxo(utxoBidYList, txPsbtYResp.TxId)


		time.Sleep(2 * time.Second)
		txPsbtXResp, err := oklink_service.BroadcastTx(txRawPsbtX)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(X) %s err:%s", req.Net, err.Error()))
		}

		txPsbtXRespTxId = txPsbtXResp.TxId
		txPsbtYRespTxId = txPsbtYResp.TxId
	}else {

		setUsedDummyUtxo(utxoDummyList, psbtYTxId)
		setUsedBidYUtxo(utxoBidYList,psbtYTxId)
		fmt.Println("Check first")
		return &respond.DoBidResp{
			TxIdX: psbtXTxId,
			TxIdY: psbtYTxId,
		}, nil


		//txPsbtYResp, err := mempool_space_service.BroadcastTx(req.Net, txRawPsbtY)
		//if err != nil {
		//	return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s", req.Net, err.Error()))
		//}
		//setUsedDummyUtxo(utxoDummyList, txPsbtYResp)
		//setUsedBidYUtxo(utxoBidYList, txPsbtYResp)
		//txPsbtXResp, err := mempool_space_service.BroadcastTx(req.Net, txRawPsbtX)
		//if err != nil {
		//	return nil, errors.New(fmt.Sprintf("Broadcast Psbt(X) %s err:%s", req.Net, err.Error()))
		//}
		//serUsedFakerInscriptionUtxo(strings.ReplaceAll(entity.InscriptionId, "i", "_"), txPsbtXResp, model.UsedYes)

		//txPsbtXRespTxId = txPsbtXResp
		//txPsbtYRespTxId = txPsbtYResp
	}

	UpdateForOrderBidDummy(entity.OrderId, model.DummyStateFinish)
	saveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, 4, psbtXTxId)
	saveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, 5, psbtXTxId)

	entity.OrderState = model.OrderStateFinish
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return nil, err
	}

	return &respond.DoBidResp{
		TxIdX: txPsbtXRespTxId,
		TxIdY: txPsbtYRespTxId,
	}, nil
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
			state := model.DummyStateFinish
			if req.OrderState == model.OrderStateCancel {
				state = model.DummyStateCancel
			}
			UpdateForOrderBidDummy(entityOrder.OrderId, state)
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

func CheckBrc20(req *request.CheckBrc20InscriptionReq) (*respond.CheckBrc20InscriptionReq, error) {
	var (
		inscriptionResp *oklink_service.OklinkInscriptionDetails
		inscription *oklink_service.InscriptionItem
		err error
		balanceDetail *oklink_service.OklinkBrc20BalanceDetails
		availableTransferState string = "fail"
		amount string = "0"
	)
	inscriptionResp, err = oklink_service.GetInscriptions("", req.InscriptionId, req.InscriptionNumber, 1, 5)
	if err != nil {
		return nil, err
	}
	for _, v := range inscriptionResp.InscriptionsList {
		if req.InscriptionId == v.InscriptionId || req.InscriptionNumber == v.InscriptionNumber {
			inscription = v
			break
		}
	}
	if inscription == nil {
		return nil, errors.New("inscription is empty")
	}
	balanceDetail, _ = oklink_service.GetAddressBrc20BalanceResult(inscription.OwnerAddress, inscription.Token, 1, 60)
	if balanceDetail != nil {
		for _,v := range balanceDetail.TransferBalanceList {
			if inscription.InscriptionId == v.InscriptionId {
				amount = v.Amount
				availableTransferState = "success"
				break
			}
		}
	}
	return &respond.CheckBrc20InscriptionReq{
		InscriptionId:          inscription.InscriptionId,
		InscriptionNumber:      inscription.InscriptionNumber,
		Location:               inscription.Location,
		InscriptionState:       inscription.State,
		Token:                  inscription.Token,
		TokenType:              inscription.TokenType,
		ActionType:             inscription.ActionType,
		OwnerAddress:           inscription.OwnerAddress,
		BlockHeight:            inscription.BlockHeight,
		TxId:                   inscription.TxId,
		AvailableTransferState: availableTransferState,
		Amount:amount,
	}, nil
}

func CheckUtxoValid()  {
	
}