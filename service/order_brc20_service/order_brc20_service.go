package order_brc20_service

import (
	"bytes"
	"crypto/sha256"
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
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
	"time"
)

const (
	inSize uint64 = 180
)

func PushOrder(req *request.OrderBrc20PushReq, publicKey string) (string, error) {
	var (
		netParams     *chaincfg.Params = GetNetParams(req.Net)
		entity        *model.OrderBrc20Model
		err           error
		orderId       string = ""
		psbtBuilder   *PsbtBuilder
		sellerAddress string = ""
		buyerAddress  string = ""
		coinAmount    uint64 = 0
		coinDec       int    = 18
		outAmount     uint64 = 0
		amountDec     int    = 8
		coinRatePrice uint64 = 0
		inscriptionId string = ""
	)

	if req.OrderState == model.OrderStateCreate {
		psbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
		if err != nil {
			return "", err
		}
		switch req.OrderType {
		case model.OrderTypeSell:
			var (
				inscriptionBrc20BalanceItem *oklink_service.BalanceItem
				has                         = false
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

			if req.Net == "mainnet" || req.Net == "livenet" {
				if !has || inscriptionBrc20BalanceItem == nil {
					return "", errors.New("Wrong Psbt: Empty inscription. ")
				}
				coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
			}
			verified, err := CheckPublicKeyAddress(netParams, publicKey, sellerAddress)
			if err != nil {
				return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
			}
			if !verified {
				return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
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
		PsbtRawPreAsk:  req.PsbtRaw,
		InscriptionId:  inscriptionId,
		Timestamp:      tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return "", err
	}
	UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	//setWhitelist(entity.SellerAddress, model.WhitelistTypeClaim, 1, 0)
	return "success", nil
}

// bid:
// 1.Buyer request to Exchange for bid
// 2.Exchange make PSBT(X) and signed with  "SIGHASH_SINGLE | ACP"
// 3.Buyer Signed with "SIGHASH_ALL | ACP"
// 4.Exchange make one input for price difference BTC and waiting for seller
// 5.Seller make another PSBT(Y) and signed with "SIGHASH_SINGLE | ACP"
// 6.Exchange Signed PSBT(Y) with "SIGHASH_DEFAULT" and broadcast
// 7.Exchange add last input for PSBT(X) with "SIGHASH_DEFAULT" and broadcast
func FetchPreBid(req *request.OrderBrc20GetBidReq) (*respond.BidPre, error) {
	var (
		brc20BalanceResult          *oklink_service.OklinkBrc20BalanceDetails
		err                         error
		list                        []*respond.AvailableItem = make([]*respond.AvailableItem, 0)
		_, platformAddressSendBrc20 string                   = GetPlatformKeyAndAddressSendBrc20(req.Net)
		poolOrderList               []*model.PoolBrc20Model
		poolOrderTotal              int64 = 0
	)
	if req.IsPool {
		poolOrderList, poolOrderTotal, err = getPoolBrc20PsbtOrder(req.Net, req.Tick, req.Limit, req.Page, 0)
		if err != nil {
			return nil, err
		}
		for _, v := range poolOrderList {
			finishCount, _ := mongo_service.FindUsedInscriptionPoolFinish(v.InscriptionId)
			if finishCount != 0 {
				fmt.Printf("finishCount InscriptionPool: [%s]\n", v.InscriptionId)
				continue
			}

			list = append(list, &respond.AvailableItem{
				InscriptionId:     v.InscriptionId,
				InscriptionNumber: v.InscriptionNumber,
				CoinAmount:        strconv.FormatUint(v.CoinAmount, 10),
				PoolOrderId:       v.OrderId,
				CoinRatePrice:     v.CoinRatePrice,
				PoolType:          v.PoolType,
				BtcPoolMode:       v.BtcPoolMode,
			})
		}
	} else {
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
		} else {
			brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressSendBrc20, req.Tick, 1, 50)
			if err != nil {
				return nil, err
			}
			for _, v := range brc20BalanceResult.TransferBalanceList {
				if CheckBidInscriptionIdExist(v.InscriptionId) {
					continue
				}
				list = append(list, &respond.AvailableItem{
					InscriptionId:     v.InscriptionId,
					InscriptionNumber: v.InscriptionNumber,
					CoinAmount:        v.Amount,
				})
			}
		}
	}

	return &respond.BidPre{
		Net:           req.Net,
		Tick:          req.Tick,
		AvailableList: list,
		Total:         poolOrderTotal,
	}, nil
}

func FetchBidPsbt(req *request.OrderBrc20GetBidReq) (*respond.BidPsbt, error) {
	var (
		brc20BalanceResult                                                *oklink_service.OklinkBrc20BalanceDetails
		err                                                               error
		bidBalanceItem                                                    *oklink_service.BalanceItem
		netParams                                                         *chaincfg.Params = GetNetParams(req.Net)
		inscriptions                                                      *oklink_service.OklinkInscriptionDetails
		inscriptionTxId                                                   string = ""
		inscriptionTxIndex                                                int64  = 0
		builder                                                           *PsbtBuilder
		psbtRaw                                                           string = ""
		poolPsbtRaw                                                       string = ""
		entityOrder                                                       *model.OrderBrc20Model
		poolOrder                                                         *model.PoolBrc20Model
		orderId                                                           string   = ""
		coinDec                                                           int      = 18
		amountDec                                                         int      = 8
		coinRatePrice                                                     uint64   = 0
		inscriptionId                                                     string   = ""
		inscriptionNumber                                                 string   = ""
		platformPrivateKeyReceiveBidValue, platformAddressReceiveBidValue string   = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		platformPrivateKeySendBrc20, platformAddressSendBrc20             string   = GetPlatformKeyAndAddressSendBrc20(req.Net)
		inputSignsExchangePriHex                                          string   = platformPrivateKeySendBrc20
		inputSignsExchangePkScript                                        string   = ""
		inputSignsUtxoType                                                UtxoType = Witness
		inputSignsTxHex                                                   string   = ""
		inputSignsAmount                                                  uint64   = 0
		poolOrderId                                                       string   = ""
		marketPrice                                                       uint64   = 0
	)
	_ = platformPrivateKeyReceiveBidValue

	req.InscriptionId = strings.ReplaceAll(req.InscriptionId, ":", "i")
	marketPrice = GetMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

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
					Amount:            "120",
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
	} else {
		if req.IsPool {
			poolOrderId = req.PoolOrderId
			poolOrder, err = getOnePoolBrc20OrderByOrderId(poolOrderId)
			if err != nil {
				return nil, err
			}

			marketPrice = poolOrder.CoinRatePrice
			bidBalanceItem = &oklink_service.BalanceItem{
				InscriptionId:     poolOrder.InscriptionId,
				InscriptionNumber: poolOrder.InscriptionNumber,
				Amount:            strconv.FormatUint(poolOrder.CoinAmount, 10),
			}
			poolPsbtRaw = poolOrder.CoinPsbtRaw
			req.CoinAmount = bidBalanceItem.Amount
		} else {
			brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressSendBrc20, req.Tick, 1, 50)
			if err != nil {
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
		}
		if marketPrice <= 0 {
			return nil, errors.New("Empty Market. ")
		}
		if bidBalanceItem == nil {
			return nil, errors.New("No Available bid. ")
		}
		inscriptions, err = oklink_service.GetInscriptions("", bidBalanceItem.InscriptionId, bidBalanceItem.InscriptionNumber, 1, 50)
		if err != nil {
			return nil, err
		}
		for _, v := range inscriptions.InscriptionsList {
			if req.InscriptionId == v.InscriptionId &&
				req.InscriptionNumber == v.InscriptionNumber {
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

	coinAmountInt, _ := strconv.ParseUint(req.CoinAmount, 10, 64)
	fmt.Printf("marketPrice:%d\n", marketPrice)
	marketPrice = marketPrice * coinAmountInt
	fmt.Printf("coinAmountInt:%d， finalSellPrice:%d\n", coinAmountInt, marketPrice)
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

	if req.IsPool {
		psbtRaw = poolPsbtRaw
	}

	//save
	outAmountDe := decimal.NewFromInt(int64(req.Amount))
	coinAmountDe := decimal.NewFromInt(int64(coinAmountInt))
	coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
	coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)
	//orderId = fmt.Sprintf("%s_%s_%s_%s_%d_%d", req.Net, req.Tick, inscriptionId, req.Address, req.Amount, coinAmountInt)
	orderId = fmt.Sprintf("%s_%s_%s_%s_%d", req.Net, req.Tick, inscriptionId, req.Address, coinAmountInt)
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
		PoolOrderId:       poolOrderId,
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
		entityOrder                       *model.OrderBrc20Model
		err                               error
		psbtBuilder                       *PsbtBuilder
		netParams                         *chaincfg.Params = GetNetParams(req.Net)
		coinRatePrice                     uint64           = 0
		buyerAddress                      string           = ""
		_, platformAddressReceiveBidValue string           = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
	)
	entityOrder, _ = mongo_service.FindOrderBrc20ModelByOrderId(req.OrderId)
	if entityOrder == nil || entityOrder.Id == 0 {
		return "", errors.New("order is empty. ")
	}
	if entityOrder.OrderType != model.OrderTypeBuy {
		return "", errors.New("Order not a bid. ")
	}

	psbtBuilder, err = NewPsbtBuilder(netParams, req.PsbtRaw)
	if err != nil {
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
		if buyerInAmount <= req.Amount {
			return "", errors.New("Wrong Psbt: The value of buyer input dose not match. ")
		}
		buyerAddress = buyerTx.OutputDetails[buyerInputIndex].OutputHash
	} else {
		buyerInAmount = req.BuyerInValue
		buyerAddress = req.Address
	}

	//check out: len-6 for no buyer changeWallet
	outList := psbtBuilder.GetOutputs()
	if len(outList) != 6 && len(outList) != 7 {
		return "", errors.New("Wrong Psbt: No match outputs. ")
	}

	//check out for platform receive marketPrice amount and address
	exchangeOrPoolOut := outList[2]
	if uint64(exchangeOrPoolOut.Value) != entityOrder.MarketAmount {
		return "", errors.New(fmt.Sprintf("Wrong Psbt: wrong value of out for receive [%d]-[%d]. ", exchangeOrPoolOut.Value, entityOrder.MarketAmount))
	}
	if entityOrder.PoolOrderId != "" {
		poolOrder, _ := mongo_service.FindPoolBrc20ModelByOrderId(entityOrder.PoolOrderId)
		if poolOrder == nil {
			entityOrder.OrderState = model.OrderStateErr
			_, err := mongo_service.SetOrderBrc20Model(entityOrder)
			if err != nil {
				return "", err
			}
			return "", errors.New(fmt.Sprintf("PSBT(X): Recheck pool order is empty"))
		} else if poolOrder.PoolState != model.PoolStateAdd {
			entityOrder.OrderState = model.OrderStateErr
			_, err := mongo_service.SetOrderBrc20Model(entityOrder)
			if err != nil {
				return "", err
			}
			return "", errors.New(fmt.Sprintf("PSBT(X): Recheck pool order status not AddState"))
		} else {
			MultiSigScriptBtye, err := hex.DecodeString(poolOrder.MultiSigScript)
			if err != nil {
				return "", err
			}
			h := sha256.New()
			h.Write(MultiSigScriptBtye)
			nativeSegwitAddress, err := btcutil.NewAddressWitnessScriptHash(h.Sum(nil), netParams)
			if err != nil {
				fmt.Println("Failed to create native SegWit address:", err)
				return "", err
			}
			_, addrs, _, err := txscript.ExtractPkScriptAddrs(exchangeOrPoolOut.PkScript, netParams)
			if err != nil {
				return "", errors.New("Wrong Psbt: Extract address from out for Seller. ")
			}
			multiSigAddress := addrs[0].EncodeAddress()
			if multiSigAddress != nativeSegwitAddress.EncodeAddress() {
				return "", errors.New("Wrong Psbt: wrong MultiSigScript of out for pool. ")
			}

			//pkScriptHex := hex.EncodeToString(exchangeOut.PkScript)
			//if pkScriptHex != poolOrder.MultiSigScript {
			//	return "", errors.New("Wrong Psbt: wrong MultiSigScript of out for pool. ")
			//}
		}
	} else {
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(exchangeOrPoolOut.PkScript, netParams)
		if err != nil {
			return "", errors.New("Wrong Psbt: Extract address from out for exchange. ")
		}
		if addrs[0].EncodeAddress() != platformAddressReceiveBidValue {
			return "", errors.New("Wrong Psbt: wrong address of out for exchange. ")
		}
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
	buyAmount := int64(buyerInAmount) - int64(platformFee) - int64(changeAmount) - int64(req.Fee) - (600 * 2)
	fmt.Printf("buyerInAmount:%d, platformFee:%d, changeAmount:%d, req.Fee:%d\n", buyerInAmount, platformFee, changeAmount, req.Fee)
	fmt.Printf("buyAmount:%d, req.Amount:%d\n", buyAmount, req.Amount)
	if buyAmount <= 0 {
		return "", errors.New("Wrong Psbt: The purchase amount is less than 0. ")
	}
	if buyAmount != int64(req.Amount) {
		return "", errors.New("Wrong Psbt: The purchase amount dose not match. ")
	}
	supplementaryAmount := int64(entityOrder.MarketAmount) - int64(buyAmount)
	fmt.Printf("supplementaryAmount:%d, entityOrder.MarketAmount:%d\n", supplementaryAmount, entityOrder.MarketAmount)
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
	if coinRatePrice == 0 {
		coinRatePrice = 1
	}

	entityOrder.PlatformFee = platformFee
	entityOrder.ChangeAmount = changeAmount
	entityOrder.Fee = req.Fee
	entityOrder.FeeRate = req.Rate
	entityOrder.SupplementaryAmount = uint64(supplementaryAmount)
	entityOrder.CoinRatePrice = coinRatePrice
	entityOrder.BuyerAddress = buyerAddress
	entityOrder.Amount = req.Amount
	entityOrder.OrderState = model.OrderStateCreate
	entityOrder.PsbtRawMidBid = req.PsbtRaw
	_, err = mongo_service.SetOrderBrc20Model(entityOrder)
	if err != nil {
		//fmt.Printf("SetOrderBrc20Model: %+v\n", entityOrder)
		return "", err
	}
	UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

	return req.OrderId, nil
}

func DoBid(req *request.OrderBrc20DoBidReq) (*respond.DoBidResp, error) {
	var (
		entity         *model.OrderBrc20Model
		err            error
		psbtBuilder    *PsbtBuilder
		btcPsbtBuilder *PsbtBuilder
		netParams      *chaincfg.Params = GetNetParams(req.Net)
		utxoDummyList  []*model.OrderUtxoModel
		utxoBidYList   []*model.OrderUtxoModel

		//startIndexDummy int64 = -1
		//startIndexBidY int64 = -1
		newPsbtBuilder                                                          *PsbtBuilder
		marketPrice                                                             uint64 = 0
		inscriptionId                                                           string = ""
		inscriptionBrc20BalanceItem                                             *oklink_service.BalanceItem
		coinAmount                                                              uint64 = 0
		brc20ReceiveValue                                                       uint64 = 0
		platformPrivateKeyReceiveBidValueToX, platformAddressReceiveBidValueToX string = GetPlatformKeyAndAddressReceiveBidValueToX(req.Net)
		_, platformAddressReceiveBidValue                                       string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		//platformPrivateKeyReceiveBidValue, platformAddressReceiveBidValue string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		_, platformAddressReceiveBrc20                                                        string = GetPlatformKeyAndAddressReceiveBrc20(req.Net)
		_, platformAddressReceiveDummyValue                                                   string = GetPlatformKeyAndAddressReceiveDummyValue(req.Net)
		_, platformAddressSendBrc20                                                           string = GetPlatformKeyAndAddressSendBrc20(req.Net)
		platformPrivateKeyReceiveBidValueForPoolBtc, platformAddressReceiveBidValueForPoolBtc string = GetPlatformKeyAndAddressReceiveValueForPoolBtc(req.Net)

		sellerSendAddress string = req.Address
		inValue           uint64 = req.Value

		platformPayPerAmount uint64 = 10000
		addressSendBrc20     string = platformAddressSendBrc20
		multiSigScript       string = ""

		inscriptionOutputValue uint64                  = 0
		poolBtcUtxoList        []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		poolBtcAmount          uint64                  = 0
		poolBtcPsbtInput       Input
		poolBtcPsbtSigIn       SigIn
		poolBtcOutput          Output

		dealTxIndex, dealTxOutValue         int64 = 0, 0 // receive btc
		dealCoinTxIndex, dealCoinTxOutValue int64 = 0, 0 // receive brc20

		bidYOffsetIndex int = 3
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
	if err != nil {
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
		inValueDe, err := decimal.NewFromString(preSellBrc20Tx.OutputDetails[preOutList[0].PreviousOutPoint.Index].Amount)
		if err != nil {
			return nil, errors.New("Wrong Psbt: The value of brc20 input decimal parse err. ")
		}
		inValue = uint64(inValueDe.Mul(decimal.New(1, 8)).IntPart())
		if inValue == 0 {
			return nil, errors.New("Wrong Psbt: brc20 out of preTx is empty amount. ")
		}
		sellerSendAddress = preSellBrc20Tx.OutputDetails[preOutList[0].PreviousOutPoint.Index].OutputHash
		time.Sleep(1000 * time.Millisecond)
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

	if req.Net == "mainnet" || req.Net == "livenet" {
		if !has || inscriptionBrc20BalanceItem == nil {
			return nil, errors.New("Wrong Psbt: Empty inscription. ")
		}
		coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
	} else {
		coinAmount, _ = strconv.ParseUint(req.CoinAmount, 10, 64)
	}
	if coinAmount != entity.CoinAmount {
		return nil, errors.New("Wrong Psbt: brc20 coin amount dose not match. ")
	}

	brc20ReceiveValue = inValue

	if entity.PoolOrderId != "" {
		poolOrder, _ := mongo_service.FindPoolBrc20ModelByOrderId(entity.PoolOrderId)
		if poolOrder == nil {
			entity.OrderState = model.OrderStateErr
			_, err := mongo_service.SetOrderBrc20Model(entity)
			if err != nil {
				return nil, err
			}
			return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck pool order is empty"))
		} else if poolOrder.PoolState != model.PoolStateAdd {
			entity.OrderState = model.OrderStateErr
			_, err := mongo_service.SetOrderBrc20Model(entity)
			if err != nil {
				return nil, err
			}
			return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck pool order status not AddState"))
		} else {
			addressSendBrc20 = poolOrder.CoinAddress
			multiSigScript = poolOrder.MultiSigScript
			inscriptionOutputValue = 5000 - brc20ReceiveValue
			if poolOrder.PoolType == model.PoolTypeBoth {
				switch poolOrder.BtcPoolMode {
				case model.PoolModeCustody, model.PoolModeNone:
					if poolOrder.UtxoId == "" {
						return nil, errors.New("pool order has empty UtxoId")
					}
					utxoIdStrs := strings.Split(poolOrder.UtxoId, "_")
					if len(utxoIdStrs) == 2 {
						addr, err := btcutil.DecodeAddress(platformAddressReceiveBidValueForPoolBtc, netParams)
						if err != nil {
							return nil, err
						}
						pkScriptBtc, err := txscript.PayToAddrScript(addr)
						if err != nil {
							return nil, err
						}
						poolBtcAmount = poolOrder.Amount
						btcUtxoTxId := utxoIdStrs[0]
						btcUtxoTxIndex, _ := strconv.ParseInt(utxoIdStrs[1], 10, 64)
						poolBtcUtxoList = append(poolBtcUtxoList, &model.OrderUtxoModel{
							TxId:          btcUtxoTxId,
							Index:         btcUtxoTxIndex,
							Amount:        poolOrder.Amount,
							PrivateKeyHex: platformPrivateKeyReceiveBidValueForPoolBtc,
							PkScript:      hex.EncodeToString(pkScriptBtc),
						})
					}
					break
				case model.PoolModePsbt:
					btcPsbtBuilder, err = NewPsbtBuilder(netParams, poolOrder.PsbtRaw)
					if err != nil {
						return nil, err
					}
					btcPreOutList := btcPsbtBuilder.GetInputs()
					if btcPreOutList == nil || len(btcPreOutList) == 0 {
						return nil, errors.New("Wrong Psbt: empty inputs in btc psbt. ")
					}
					btcOutputList := btcPsbtBuilder.GetOutputs()
					if btcOutputList == nil || len(btcOutputList) == 0 {
						return nil, errors.New("Wrong Psbt: empty outputs in btc psbt. ")
					}
					poolBtcPsbtInput = Input{
						OutTxId:  btcPreOutList[0].PreviousOutPoint.Hash.String(),
						OutIndex: btcPreOutList[0].PreviousOutPoint.Index,
					}
					finalScriptWitness := btcPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].FinalScriptWitness
					witnessUtxo := btcPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].WitnessUtxo
					sighashType := btcPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].SighashType
					poolBtcPsbtSigIn = SigIn{
						WitnessUtxo:        witnessUtxo,
						SighashType:        sighashType,
						FinalScriptWitness: finalScriptWitness,
						Index:              bidYOffsetIndex,
					}
					bidYOffsetIndex = bidYOffsetIndex + 1
					poolBtcAmount = poolOrder.Amount

					poolBtcOutput = Output{
						Amount: uint64(btcOutputList[0].Value),
						Script: hex.EncodeToString(btcOutputList[0].PkScript),
					}
					break
				default:
					return nil, errors.New("pool order has wrong poolMode")
				}
			}
		}
	}

	utxoDummyList, err = GetUnoccupiedUtxoList(req.Net, 2, 0, model.UtxoTypeDummy)
	defer ReleaseUtxoList(utxoDummyList)
	if err != nil {
		return nil, err
	}

	//get bidY pay utxo
	totalNeedAmount := entity.SupplementaryAmount + sellerReceiveValue + entity.Fee + inscriptionOutputValue
	limit := totalNeedAmount/platformPayPerAmount + 1
	changeAmount := platformPayPerAmount*limit - totalNeedAmount
	if poolBtcAmount != 0 {
		totalNeedAmount = totalNeedAmount - poolBtcAmount
		limit = totalNeedAmount/platformPayPerAmount + 1
		changeAmount = platformPayPerAmount*limit - totalNeedAmount
	}
	fmt.Printf("[DO]totalNeedAmount: %d， poolBtcAmount：%d, changeAmount: %d\n", totalNeedAmount, poolBtcAmount, changeAmount)

	utxoBidYList, err = GetUnoccupiedUtxoList(req.Net, int64(limit), int64(totalNeedAmount), model.UtxoTypeBidY)
	defer ReleaseUtxoList(utxoBidYList)
	if err != nil {
		return nil, err
	}
	utxoBidYList = append(utxoBidYList, poolBtcUtxoList...)

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
	//add btc pool psbt ins - index: 3
	if poolBtcPsbtInput.OutTxId != "" {
		inputs = append(inputs, poolBtcPsbtInput)
	}

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
	receiveBrc20 := Output{
		Address: platformAddressReceiveBrc20,
		Amount:  brc20ReceiveValue + inscriptionOutputValue,
	}
	if entity.PoolOrderId != "" {
		multiSigScriptByte, err := hex.DecodeString(multiSigScript)
		if err != nil {
			return nil, err
		}
		h := sha256.New()
		h.Write(multiSigScriptByte)
		segwitAddress, err := btcutil.NewAddressWitnessScriptHash(h.Sum(nil), netParams)
		if err != nil {
			fmt.Println("Failed to create native SegWit address:", err)
			return nil, errors.New(fmt.Sprintf("Failed to create native SegWit address:%s", err.Error()))
		}
		segwitAddr, err := btcutil.DecodeAddress(segwitAddress.EncodeAddress(), netParams)
		if err != nil {
			return nil, err
		}
		receiveBrc20 = Output{
			Address: segwitAddr.EncodeAddress(),
			//Script: multiSigScript,
			Amount: brc20ReceiveValue + inscriptionOutputValue,
		}
		dealCoinTxIndex, dealCoinTxOutValue = 1, int64(brc20ReceiveValue+inscriptionOutputValue)
	}
	outputs = append(outputs, receiveBrc20)
	//add receive seller outs - idnex: 2
	outputs = append(outputs, Output{
		Address: sellerReceiveAddress,
		Amount:  sellerReceiveValue,
	})

	if poolBtcOutput.Script != "" || poolBtcOutput.Address != "" {
		//add receive pool btc outs - idnex: 3
		outputs = append(outputs, poolBtcOutput)
	}

	_ = marketPrice
	//add receive exchange psbtX outs - idnex: 3/4
	psbtXValue := entity.MarketAmount - sellerReceiveValue
	exchangePsbtXOut := Output{
		Address: platformAddressReceiveBidValueToX,
		Amount:  psbtXValue,
	}
	outputs = append(outputs, exchangePsbtXOut)
	//add new dummy outs - idnex: 4,5 / 5,6
	newDummyOut := Output{
		Address: newDummyOutSegwitAddress,
		Amount:  600,
	}
	outputs = append(outputs, newDummyOut)
	outputs = append(outputs, newDummyOut)

	if changeAmount >= 546 {
		outputs = append(outputs, Output{
			Address: platformAddressReceiveBidValue,
			Amount:  changeAmount,
		})
	}

	//finish PSBT(Y)
	newPsbtBuilder, err = CreatePsbtBuilder(netParams, inputs, outputs)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): newPsbtBuilder err:%s", err.Error()))
	}

	finalScriptWitness := psbtBuilder.PsbtUpdater.Upsbt.Inputs[0].FinalScriptWitness
	witnessUtxo := psbtBuilder.PsbtUpdater.Upsbt.Inputs[0].WitnessUtxo
	sighashType := psbtBuilder.PsbtUpdater.Upsbt.Inputs[0].SighashType
	err = newPsbtBuilder.AddSigIn(witnessUtxo, sighashType, finalScriptWitness, 2)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): AddPartialSigIn err:%s", err.Error()))
	}

	if poolBtcPsbtSigIn.WitnessUtxo != nil {
		err = newPsbtBuilder.AddSinInStruct(&poolBtcPsbtSigIn)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("PSBT(Y): AddPartialSigIn err:%s", err.Error()))
		}
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
	fmt.Printf("bidYOffsetIndex: %d\n", bidYOffsetIndex)
	//add Exchange pay value ins - index: 3,3+
	for k, payBid := range utxoBidYList {
		inSigns = append(inSigns, &InputSign{
			Index:       k + bidYOffsetIndex,
			PkScript:    payBid.PkScript,
			Amount:      payBid.Amount,
			SighashType: txscript.SigHashAll,
			PriHex:      payBid.PrivateKeyHex,
			UtxoType:    Witness,
		})
	}
	err = newPsbtBuilder.UpdateAndSignInput(inSigns)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): UpdateAndSignInput err:%s", err.Error()))
	}
	psbtRawFinalAsk, err := newPsbtBuilder.ToString()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): ToString err:%s", err.Error()))
	}
	entity.PsbtRawFinalAsk = psbtRawFinalAsk

	txRawPsbtY, err := newPsbtBuilder.ExtractPsbtTransaction()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): ExtractPsbtTransaction err:%s", err.Error()))
	}
	txRawPsbtYByte, _ := hex.DecodeString(txRawPsbtY)

	txPsbtY := wire.NewMsgTx(2)
	err = txPsbtY.Deserialize(bytes.NewReader(txRawPsbtYByte))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): txRawPsbtY Deserialize err:%s", err.Error()))
	}

	psbtYTxId := txPsbtY.TxHash().String()
	//
	fmt.Printf("Tx:\n")
	fmt.Printf("%+v\n", txPsbtY)
	for _, in := range txPsbtY.TxIn {
		fmt.Printf("%+v\n", *in)
	}
	for _, out := range txPsbtY.TxOut {
		fmt.Printf("%+v\n", *out)
	}
	fmt.Printf("\n")

	//finish PSBT(X)
	bidPsbtBuilder, err := NewPsbtBuilder(netParams, entity.PsbtRawMidBid)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(X): NewPsbtBuilder err:%s", err.Error()))
	}
	//Check PSBT(X) utxo valid
	insPsbtX := bidPsbtBuilder.GetInputs()
	addressUtxoMap := make(map[string][]*wire.TxIn)
	addressUtxoMap[entity.BuyerAddress] = make([]*wire.TxIn, 0)
	addressUtxoMap[addressSendBrc20] = make([]*wire.TxIn, 0)
	for k, v := range insPsbtX {
		if k == 2 {
			addressUtxoMap[addressSendBrc20] = append(addressUtxoMap[addressSendBrc20], v)
		} else {
			addressUtxoMap[entity.BuyerAddress] = append(addressUtxoMap[entity.BuyerAddress], v)
		}
	}
	liveUtxoList := make([]*oklink_service.UtxoItem, 0)
	//liveUtxoList := make([]*unisat_service.UtxoDetailItem, 0)
	if entity.Net != "testnet" {
		for address, _ := range addressUtxoMap {
			utxoResp, err := oklink_service.GetAddressUtxo(address, 1, 50)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck address utxo list err:%s", err.Error()))
			}
			if utxoResp.UtxoList != nil && len(utxoResp.UtxoList) != 0 {
				liveUtxoList = append(liveUtxoList, utxoResp.UtxoList...)
			}

			utxoList, err := unisat_service.GetAddressUtxo(address)
			//if err != nil {
			//	return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck address utxo list err:%s", err.Error()))
			//}
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
			time.Sleep(1200 * time.Millisecond)
		}
	}

	for _, v := range insPsbtX {
		bidInId := fmt.Sprintf("%s_%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
		has := false
		for _, u := range liveUtxoList {
			uId := fmt.Sprintf("%s_%s", u.TxId, u.Index)
			//uId := fmt.Sprintf("%s_%d", u.TxId, u.OutputIndex)
			fmt.Printf("liveUtxo:[%s]\n", uId)
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
			return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck address utxo list, utxo had been spent: %s", bidInId))
		}
	}

	bidIn := Input{
		OutTxId:  psbtYTxId,
		OutIndex: uint32(bidYOffsetIndex),
	}
	bidInSign := &InputSign{
		UtxoType:    Witness,
		Index:       len(bidPsbtBuilder.GetInputs()),
		OutRaw:      txRawPsbtY,
		PkScript:    hex.EncodeToString(txPsbtY.TxOut[bidIn.OutIndex].PkScript),
		Amount:      psbtXValue,
		SighashType: txscript.SigHashAll | txscript.SigHashAnyOneCanPay,
		PriHex:      platformPrivateKeyReceiveBidValueToX,
	}
	err = bidPsbtBuilder.AddInput(bidIn, bidInSign)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(X): AddInput err:%s", err.Error()))
	}
	psbtRawFinalBid, err := bidPsbtBuilder.ToString()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(X): ToString err:%s", err.Error()))
	}
	entity.PsbtRawFinalBid = psbtRawFinalBid

	txRawPsbtX, err := bidPsbtBuilder.ExtractPsbtTransaction()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(X): ExtractPsbtTransaction err:%s", err.Error()))
	}
	txRawPsbtXByte, _ := hex.DecodeString(txRawPsbtX)

	txPsbtX := wire.NewMsgTx(2)
	err = txPsbtX.Deserialize(bytes.NewReader(txRawPsbtXByte))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(Y): txRawPsbtY Deserialize err:%s", err.Error()))
	}

	psbtXTxId := txPsbtX.TxHash().String()
	dealTxIndex = 2
	dealTxOutValue = txPsbtX.TxOut[dealTxIndex].Value

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
	if req.Net == "mainnet" || req.Net == "livenet" {
		txPsbtYResp, err := unisat_service.BroadcastTx(req.Net, txRawPsbtY)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s", req.Net, err.Error()))
		}
		SetUsedDummyUtxo(utxoDummyList, txPsbtYResp.Result)
		setUsedBidYUtxo(utxoBidYList, txPsbtYResp.Result)

		time.Sleep(2 * time.Second)
		txPsbtXResp, err := unisat_service.BroadcastTx(req.Net, txRawPsbtX)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(X) %s err:%s", req.Net, err.Error()))
		}

		txPsbtXRespTxId = txPsbtXResp.Result
		txPsbtYRespTxId = txPsbtYResp.Result

		//txPsbtYResp, err := node.BroadcastTx(req.Net, txRawPsbtY)
		//if err != nil {
		//	return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s", req.Net, err.Error()))
		//}
		//setUsedDummyUtxo(utxoDummyList, txPsbtYResp)
		//setUsedBidYUtxo(utxoBidYList, txPsbtYResp)
		//
		//time.Sleep(2 * time.Second)
		//txPsbtXResp, err := node.BroadcastTx(req.Net, txRawPsbtX)
		//if err != nil {
		//	return nil, errors.New(fmt.Sprintf("Broadcast Psbt(X) %s err:%s", req.Net, err.Error()))
		//}
		//
		//txPsbtXRespTxId = txPsbtXResp
		//txPsbtYRespTxId = txPsbtYResp

	} else {

		txPsbtYResp, err := unisat_service.BroadcastTx(req.Net, txRawPsbtY)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s", req.Net, err.Error()))
		}
		SetUsedDummyUtxo(utxoDummyList, txPsbtYResp.Result)
		setUsedBidYUtxo(utxoBidYList, txPsbtYResp.Result)

		time.Sleep(2 * time.Second)
		txPsbtXResp, err := unisat_service.BroadcastTx(req.Net, txRawPsbtX)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(X) %s err:%s", req.Net, err.Error()))
		}

		txPsbtXRespTxId = txPsbtXResp.Result
		txPsbtYRespTxId = txPsbtYResp.Result

		//txPsbtYResp, err := node.BroadcastTx(req.Net, txRawPsbtY)
		//if err != nil {
		//	return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s", req.Net, err.Error()))
		//}
		//setUsedDummyUtxo(utxoDummyList, txPsbtYResp)
		//setUsedBidYUtxo(utxoBidYList, txPsbtYResp)
		//txPsbtXResp, err := node.BroadcastTx(req.Net, txRawPsbtX)
		//if err != nil {
		//	return nil, errors.New(fmt.Sprintf("Broadcast Psbt(X) %s err:%s", req.Net, err.Error()))
		//}
		//serUsedFakerInscriptionUtxo(strings.ReplaceAll(entity.InscriptionId, "i", "_"), txPsbtXResp, model.UsedYes)
		//
		//txPsbtXRespTxId = txPsbtXResp
		//txPsbtYRespTxId = txPsbtYResp
	}

	UpdateForOrderBidDummy(entity.OrderId, model.DummyStateFinish)
	SaveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, 4, psbtYTxId)
	SaveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, 5, psbtYTxId)

	entity.DealTime = tool.MakeTimestamp()
	entity.OrderState = model.OrderStateFinish
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return nil, err
	}

	//todo use DB-Transaction
	if entity.PoolOrderId != "" {
		setStatusPoolBrc20Order(entity, model.PoolStateUsed, dealTxIndex, dealTxOutValue, entity.DealTime)
		setCoinStatusPoolBrc20Order(entity, model.PoolStateUsed, dealCoinTxIndex, dealCoinTxOutValue, entity.DealTime)
	}

	UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

	return &respond.DoBidResp{
		TxIdX: txPsbtXRespTxId,
		TxIdY: txPsbtYRespTxId,
	}, nil
}

func UpdateOrder(req *request.OrderBrc20UpdateReq, publicKey, ip string) (string, error) {
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
	if entityOrder.FreeState > 1 {
		return "", errors.New("Order err. ")
	}

	if req.OrderState == model.OrderStateFinish || req.OrderState == model.OrderStateCancel {
		entityOrder.OrderState = req.OrderState
		switch entityOrder.OrderType {
		case model.OrderTypeSell:
			if req.OrderState == model.OrderStateCancel {
				verified, err := CheckPublicKeyAddress(netParams, publicKey, entityOrder.SellerAddress)
				if err != nil {
					return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
				}
				if !verified {
					return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
				}
			} else {
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

				if req.BroadcastIndex == 1 {
					txPsbtResp, err := unisat_service.BroadcastTx(entityOrder.Net, txRaw)
					if err != nil {
						return "", errors.New(fmt.Sprintf("Broadcast Psbt %s, orderId-%s err:%s", entityOrder.Net, entityOrder.OrderId, err.Error()))
					}

					entityOrder.PsbtAskTxId = txPsbtResp.Result
					entityOrder.OrderState = model.OrderStateFinish
					//setWhitelist(entityOrder.BuyerAddress, model.WhitelistTypeClaim, 1, 0)
				}

			}

			entityOrder.PsbtRawFinalAsk = req.PsbtRaw
			break
		case model.OrderTypeBuy:
			if req.OrderState == model.OrderStateCancel {
				verified, err := CheckPublicKeyAddress(netParams, publicKey, entityOrder.BuyerAddress)
				if err != nil {
					return "", errors.New(fmt.Sprintf("Check address err: %s. ", err.Error()))
				}
				if !verified {
					return "", errors.New(fmt.Sprintf("Check address verified: %v. ", verified))
				}
			} else {
				return "", errors.New(fmt.Sprintf("Wrong OrderState. "))
			}

			entityOrder.Fee = 0
			entityOrder.PsbtRawFinalBid = req.PsbtRaw
			state := model.DummyStateFinish
			if req.OrderState == model.OrderStateCancel {
				state = model.DummyStateCancel
			}
			UpdateForOrderBidDummy(entityOrder.OrderId, state)
			break
		}
		entityOrder.DealTime = tool.MakeTimestamp()
		_, err = mongo_service.SetOrderBrc20Model(entityOrder)
		if err != nil {
			return "", err
		}
		UpdateMarketPrice(req.Net, entityOrder.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(entityOrder.Tick)))
	} else {
		return "", errors.New("Wrong state. ")
	}

	return req.OrderId, nil
}

func CheckBrc20(req *request.CheckBrc20InscriptionReq) (*respond.CheckBrc20InscriptionReq, error) {
	var (
		inscriptionResp        *oklink_service.OklinkInscriptionDetails
		inscription            *oklink_service.InscriptionItem
		err                    error
		balanceDetail          *oklink_service.OklinkBrc20BalanceDetails
		availableTransferState string = "fail"
		amount                 string = "0"
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
		for _, v := range balanceDetail.TransferBalanceList {
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
		Amount:                 amount,
	}, nil
}
