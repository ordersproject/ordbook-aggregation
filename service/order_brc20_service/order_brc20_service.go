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
		netParams           *chaincfg.Params = GetNetParams(req.Net)
		entity              *model.OrderBrc20Model
		err                 error
		orderId             string = ""
		psbtBuilder         *PsbtBuilder
		sellerAddress       string = ""
		buyerAddress        string = ""
		coinAmount          uint64 = 0
		coinDec             int    = 18
		outAmount           uint64 = 0
		amountDec           int    = 8
		coinRatePrice       uint64 = 0
		coinPrice           int64  = 0
		coinPriceDecimalNum int32  = 0
		inscriptionId       string = ""
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
				inscriptionId = fmt.Sprintf("%si%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
				inscriptionBrc20BalanceItem, err = CheckBrc20Ordinals(v, req.Tick, sellerAddress)
				if err != nil {
					continue
				}
				has = true
			}

			if req.Net == "mainnet" || req.Net == "livenet" {
				if !has || inscriptionBrc20BalanceItem == nil {
					return "", errors.New("Wrong Psbt: Empty inscription. Please use a valid BRC20 token or re-inscribe the BRC20 token. ")
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

	coinPrice, coinPriceDecimalNum = MakePrice(int64(coinAmount), int64(outAmount))
	entity = &model.OrderBrc20Model{
		Net:                 req.Net,
		OrderId:             orderId,
		Tick:                req.Tick,
		Amount:              outAmount,
		DecimalNum:          amountDec,
		CoinAmount:          coinAmount,
		CoinDecimalNum:      coinDec,
		CoinRatePrice:       coinRatePrice,
		CoinPrice:           coinPrice,
		CoinPriceDecimalNum: coinPriceDecimalNum,
		OrderState:          req.OrderState,
		OrderType:           req.OrderType,
		SellerAddress:       sellerAddress,
		BuyerAddress:        buyerAddress,
		PsbtRawPreAsk:       req.PsbtRaw,
		InscriptionId:       inscriptionId,
		Timestamp:           tool.MakeTimestamp(),
		PlatformDummy:       req.PlatformDummy,
	}
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return "", err
	}
	UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	UpdateMarketPriceV2(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
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
		marketPrice := GetMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
		marketCoinPrice := GetMarketPriceV2(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

		for _, v := range poolOrderList {
			finishCount, _ := mongo_service.FindUsedInscriptionPoolFinish(v.InscriptionId)
			if finishCount != 0 {
				//fmt.Printf("finishCount InscriptionPool: [%s]\n", v.InscriptionId)
				continue
			}

			if req.Address != "" {
				bidCountOwner := checkPoolBidCount(v.OrderId, req.Address)
				if bidCountOwner > 0 {
					continue
				}

				poolOwner := checkPoolAddress(v.OrderId, req.Address)
				if poolOwner > 0 {
					continue
				}
			}

			//if marketPrice > v.CoinRatePrice || marketCoinPrice > uint64(v.CoinPrice) {
			if marketCoinPrice > uint64(v.CoinPrice) {
				fmt.Printf("marketPrice not enrough OrderId: [%s] [%d - %d][%d - %d]\n", v.OrderId, marketPrice, v.CoinRatePrice, marketCoinPrice, v.CoinPrice)
				//fmt.Printf("marketPrice not enrough OrderId: [%s]\n", v.OrderId)
				continue
			}

			bidCount := checkPoolBidCount(v.OrderId, "")
			list = append(list, &respond.AvailableItem{
				InscriptionId:       v.InscriptionId,
				InscriptionNumber:   v.InscriptionNumber,
				CoinAmount:          strconv.FormatUint(v.CoinAmount, 10),
				PoolOrderId:         v.OrderId,
				CoinRatePrice:       v.CoinRatePrice,
				CoinPrice:           v.CoinPrice,
				CoinPriceDecimalNum: v.CoinPriceDecimalNum,
				PoolType:            v.PoolType,
				BtcPoolMode:         v.BtcPoolMode,
				BidCount:            bidCount,
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
		coinPrice                                                         int64    = 0
		coinPriceDecimalNum                                               int32    = 0
		marketCoinPrice                                                   uint64   = 0
		marketCoinPriceDecimalNum                                         int32    = coinPriceDecimalNumDefault
	)
	_ = platformPrivateKeyReceiveBidValue

	req.InscriptionId = strings.ReplaceAll(req.InscriptionId, ":", "i")
	marketPrice = GetMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	marketCoinPrice = GetMarketPriceV2(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

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

			if marketPrice > poolOrder.CoinRatePrice || marketCoinPrice > uint64(poolOrder.CoinPrice) {
				return nil, errors.New("Market price is higher than this lp. ")
			}

			marketPrice = poolOrder.CoinRatePrice
			marketCoinPrice = uint64(poolOrder.CoinPrice)
			bidBalanceItem = &oklink_service.BalanceItem{
				InscriptionId:     poolOrder.InscriptionId,
				InscriptionNumber: poolOrder.InscriptionNumber,
				Amount:            strconv.FormatUint(poolOrder.CoinAmount, 10),
			}
			inscriptionId = bidBalanceItem.InscriptionId
			inscriptionId = strings.ReplaceAll(inscriptionId, ":", "i")
			inscriptionNumber = bidBalanceItem.InscriptionNumber
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
		inscriptions, err = oklink_service.GetInscriptions("", bidBalanceItem.InscriptionId, "", 1, 50)
		if err != nil {
			return nil, err
		}
		for _, v := range inscriptions.InscriptionsList {
			//if req.InscriptionId == v.InscriptionId && req.InscriptionNumber == v.InscriptionNumber {
			if req.InscriptionId == v.InscriptionId {
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

		//if inscriptionTxId == "" {
		//
		//}

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

	if req.SwitchPrice == 1 {
		fmt.Printf("marketCoinPrice:%d\n", marketCoinPrice)
		marketPriceInt, err := GetPrice(int64(coinAmountInt), int64(marketCoinPrice), marketCoinPriceDecimalNum)
		if err != nil {
			return nil, err
		}
		marketPrice = uint64(marketPriceInt)
		fmt.Printf("coinAmountInt:%d， finalSellPrice:%d\n", coinAmountInt, marketPrice)

	}

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
	inscriptionId = strings.ReplaceAll(inscriptionId, ":", "i")
	fmt.Printf("[BID]build-inscriptionId:%s, address:%s\n", inscriptionId, req.Address)
	orderId = fmt.Sprintf("%s_%s_%s_%s_%d", req.Net, req.Tick, inscriptionId, req.Address, coinAmountInt)
	orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
	coinPrice, coinPriceDecimalNum = MakePrice(int64(coinAmountInt), int64(req.Amount))
	entityOrder = &model.OrderBrc20Model{
		Net:                 req.Net,
		OrderId:             orderId,
		Tick:                req.Tick,
		Amount:              req.Amount,
		DecimalNum:          amountDec,
		CoinAmount:          coinAmountInt,
		CoinDecimalNum:      coinDec,
		CoinRatePrice:       coinRatePrice,
		CoinPrice:           coinPrice,
		CoinPriceDecimalNum: coinPriceDecimalNum,
		OrderState:          model.OrderStatePreCreate,
		OrderType:           model.OrderTypeBuy,
		SellerAddress:       "",
		BuyerAddress:        req.Address,
		MarketAmount:        marketPrice,
		PlatformTx:          inscriptionTxId,
		InscriptionId:       inscriptionId,
		InscriptionNumber:   inscriptionNumber,
		PsbtRawPreAsk:       "",
		PsbtRawPreBid:       psbtRaw,
		PoolOrderId:         poolOrderId,
		Timestamp:           tool.MakeTimestamp(),
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

func FetchBidPsbtByPlatform(req *request.OrderBrc20GetBidPlatformReq) (*respond.BidPsbt, error) {
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
		platformPrivateKeyDummy, platformAddressDummy                     string   = GetPlatformKeyAndAddressForDummy(req.Net)
		_, platformAddressReceiveBidValueToReturn                         string   = GetPlatformKeyAndAddressReceiveBidValueToReturn(req.Net)
		inputSignsExchangePriHex                                          string   = platformPrivateKeySendBrc20
		inputSignsExchangePkScript                                        string   = ""
		inputSignsUtxoType                                                UtxoType = Witness
		inputSignsTxHex                                                   string   = ""
		inputSignsAmount                                                  uint64   = 0
		poolOrderId                                                       string   = ""
		marketPrice                                                       uint64   = 0
		utxoDummy1200List                                                 []*model.OrderUtxoModel
		utxoDummyList                                                     []*model.OrderUtxoModel
		buyerAddress                                                      string = ""
		poolInputIndex                                                    int    = 2
		poolPsbtBuilder                                                   *PsbtBuilder
		poolPsbtInput                                                     Input
		poolPsbtSigIn                                                     SigIn
		poolOutput                                                        Output
		poolBrc20InputValue                                               uint64 = 0
		coinPrice                                                         int64  = 0
		coinPriceDecimalNum                                               int32  = 0
		marketCoinPrice                                                   uint64 = 0
		marketCoinPriceDecimalNum                                         int32  = coinPriceDecimalNumDefault
	)
	_ = platformPrivateKeyReceiveBidValue
	buyerAddress = req.Address
	req.InscriptionId = strings.ReplaceAll(req.InscriptionId, ":", "i")
	marketPrice = GetMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	marketCoinPrice = GetMarketPriceV2(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

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
			marketCoinPrice = uint64(poolOrder.CoinPrice)
			bidBalanceItem = &oklink_service.BalanceItem{
				InscriptionId:     poolOrder.InscriptionId,
				InscriptionNumber: poolOrder.InscriptionNumber,
				Amount:            strconv.FormatUint(poolOrder.CoinAmount, 10),
			}
			inscriptionId = bidBalanceItem.InscriptionId
			inscriptionId = strings.ReplaceAll(inscriptionId, ":", "i")
			inscriptionNumber = bidBalanceItem.InscriptionNumber
			poolPsbtRaw = poolOrder.CoinPsbtRaw
			req.CoinAmount = bidBalanceItem.Amount

			poolBrc20InputValue = poolOrder.CoinInputValue
			poolPsbtBuilder, err = NewPsbtBuilder(netParams, poolOrder.CoinPsbtRaw)
			if err != nil {
				return nil, err
			}
			poolPreOutList := poolPsbtBuilder.GetInputs()
			if poolPreOutList == nil || len(poolPreOutList) == 0 {
				return nil, errors.New("Wrong pool Psbt: empty inputs in brc20 psbt. ")
			}
			poolOutputList := poolPsbtBuilder.GetOutputs()
			if poolOutputList == nil || len(poolOutputList) == 0 {
				return nil, errors.New("Wrong pool Psbt: empty outputs in brc20 psbt. ")
			}
			poolPsbtInput = Input{
				OutTxId:  poolPreOutList[0].PreviousOutPoint.Hash.String(),
				OutIndex: poolPreOutList[0].PreviousOutPoint.Index,
			}
			finalScriptWitness := poolPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].FinalScriptWitness
			witnessUtxo := poolPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].WitnessUtxo
			sighashType := poolPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].SighashType
			poolPsbtSigIn = SigIn{
				WitnessUtxo:        witnessUtxo,
				SighashType:        sighashType,
				FinalScriptWitness: finalScriptWitness,
				Index:              poolInputIndex,
			}
			poolOutput = Output{
				Amount: uint64(poolOutputList[0].Value),
				Script: hex.EncodeToString(poolOutputList[0].PkScript),
			}

		} else {
			brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressSendBrc20, req.Tick, 1, 50)
			if err != nil {
				return nil, err
			}
			inscriptionId = req.InscriptionId
			inscriptionNumber = req.InscriptionNumber
			for _, v := range brc20BalanceResult.TransferBalanceList {
				if req.InscriptionId == v.InscriptionId && req.CoinAmount == v.Amount {
					//if req.InscriptionId == v.InscriptionId && req.InscriptionNumber == v.InscriptionNumber && req.CoinAmount == v.Amount {
					bidBalanceItem = v
					break
				}
			}
		}

		if bidBalanceItem == nil {
			return nil, errors.New("No Available bid. ")
		}
		inscriptions, err = oklink_service.GetInscriptions("", bidBalanceItem.InscriptionId, "", 1, 50)
		if err != nil {
			return nil, err
		}
		for _, v := range inscriptions.InscriptionsList {
			//if req.InscriptionId == v.InscriptionId && req.InscriptionNumber == v.InscriptionNumber {
			if req.InscriptionId == v.InscriptionId {
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

	//find dummy utxo
	utxoDummyList, err = GetUnoccupiedUtxoList(req.Net, 2, 0, model.UtxoTypeDummyBidX, "", 0)
	defer ReleaseUtxoList(utxoDummyList)
	if err != nil {
		return nil, err
	}
	utxoDummy1200List, err = GetUnoccupiedUtxoList(req.Net, 1, 0, model.UtxoTypeDummy1200BidX, "", 0)
	defer ReleaseUtxoList(utxoDummy1200List)
	if err != nil {
		return nil, err
	}

	inputs := make([]Input, 0)
	dummyOutValue := uint64(0)
	//add dummy input: 0,1
	for _, dummy := range utxoDummyList {
		inputs = append(inputs, Input{
			OutTxId:  dummy.TxId,
			OutIndex: uint32(dummy.Index),
		})
		dummyOutValue = dummyOutValue + dummy.Amount
	}

	//add brc20 input: 2
	if req.IsPool && poolPsbtInput.OutTxId != "" {
		inputs = append(inputs, poolPsbtInput)
	} else {
		inputs = append(inputs, Input{
			OutTxId:  inscriptionTxId,
			OutIndex: uint32(inscriptionTxIndex),
		})
	}

	//add dummy1200 input: 3
	for _, dummy := range utxoDummy1200List {
		inputs = append(inputs, Input{
			OutTxId:  dummy.TxId,
			OutIndex: uint32(dummy.Index),
		})
	}

	coinAmountInt, _ := strconv.ParseUint(req.CoinAmount, 10, 64)
	fmt.Printf("marketPrice:%d\n", marketPrice)
	marketPrice = marketPrice * coinAmountInt
	fmt.Printf("coinAmountInt:%d， finalSellPrice:%d\n", coinAmountInt, marketPrice)
	if req.SwitchPrice == 1 {
		if marketCoinPrice <= 0 {
			return nil, errors.New("Empty Market. ")
		}
		fmt.Printf("marketCoinPrice:%d\n", marketCoinPrice)
		marketPriceInt, err := GetPrice(int64(coinAmountInt), int64(marketCoinPrice), marketCoinPriceDecimalNum)
		if err != nil {
			return nil, err
		}
		marketPrice = uint64(marketPriceInt)
		fmt.Printf("coinAmountInt:%d， finalSellPrice:%d\n", coinAmountInt, marketPrice)
	} else {
		if marketPrice <= 0 {
			return nil, errors.New("Empty Market. ")
		}
	}

	//todo check marketPrice >= req.amount

	outputs := make([]Output, 0)
	// add dummy1200 output: 0
	outputs = append(outputs, Output{
		Address: platformAddressDummy,
		Amount:  dummyOutValue,
	})

	// add buyer receive brc20 output: 1
	outputs = append(outputs, Output{
		Address: buyerAddress,
		Amount:  poolBrc20InputValue,
	})

	// add platform receive btc output: 2
	if req.IsPool && poolOutput.Script != "" {
		outputs = append(outputs, poolOutput)
	} else {
		outputs = append(outputs, Output{
			Address: platformAddressReceiveBidValue,
			Amount:  marketPrice,
		})
	}

	// add 10000 fee output: 3,4
	fee10000Out := Output{
		Address: platformAddressReceiveBidValueToReturn,
		Amount:  10000,
	}
	outputs = append(outputs, fee10000Out)
	outputs = append(outputs, fee10000Out)

	// add dummy output: 5,6
	dummyOut600 := Output{
		Address: platformAddressDummy,
		Amount:  600,
	}
	outputs = append(outputs, dummyOut600)
	outputs = append(outputs, dummyOut600)

	// add change output: 7
	if model.PlatformDummy(req.PlatformDummy) == model.PlatformDummyYes && req.BidTxSpec != nil && len(req.BidTxSpec.Outputs) != 0 {
		for _, v := range req.BidTxSpec.Outputs {
			if v.Type == "change" {
				changeOut := Output{
					Address: v.Address,
					Amount:  uint64(v.Value),
				}
				outputs = append(outputs, changeOut)
			}
		}
	}

	inputSigns := make([]*InputSign, 0)
	platformDummyPkScript, err := AddressToPkScript(req.Net, platformAddressDummy)
	if err != nil {
		return nil, errors.New("AddressToPkScript err: " + err.Error())
	}
	//add dummy inputSign: 0,1
	inputSigns = append(inputSigns, &InputSign{
		Index:       0,
		OutRaw:      "",
		PkScript:    platformDummyPkScript,
		SighashType: txscript.SigHashAll | txscript.SigHashAnyOneCanPay,
		PriHex:      platformPrivateKeyDummy,
		UtxoType:    Witness,
		Amount:      600,
	})
	inputSigns = append(inputSigns, &InputSign{
		Index:       1,
		OutRaw:      "",
		PkScript:    platformDummyPkScript,
		SighashType: txscript.SigHashAll | txscript.SigHashAnyOneCanPay,
		PriHex:      platformPrivateKeyDummy,
		UtxoType:    Witness,
		Amount:      600,
	})

	if !req.IsPool {
		inputSigns = append(inputSigns, &InputSign{
			Index:       2,
			OutRaw:      inputSignsTxHex,
			PkScript:    inputSignsExchangePkScript,
			SighashType: txscript.SigHashSingle | txscript.SigHashAnyOneCanPay,
			PriHex:      inputSignsExchangePriHex,
			UtxoType:    inputSignsUtxoType,
			Amount:      inputSignsAmount,
		})
	}

	inputSigns = append(inputSigns, &InputSign{
		Index:       3,
		OutRaw:      "",
		PkScript:    platformDummyPkScript,
		SighashType: txscript.SigHashAll | txscript.SigHashAnyOneCanPay,
		PriHex:      platformPrivateKeyDummy,
		UtxoType:    Witness,
		Amount:      1200,
	})

	builder, err = CreatePsbtBuilder(netParams, inputs, outputs)
	if err != nil {
		return nil, err
	}

	if req.IsPool && poolPsbtSigIn.WitnessUtxo != nil {
		err = builder.AddSinInStruct(&poolPsbtSigIn)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("PSBT(X): AddPartialSigIn err:%s", err.Error()))
		}
	}

	err = builder.UpdateAndSignInput(inputSigns)
	if err != nil {
		return nil, err
	}

	psbtRaw, err = builder.ToString()
	if err != nil {
		return nil, err
	}

	if req.IsPool && model.PlatformDummy(req.PlatformDummy) != model.PlatformDummyYes {
		psbtRaw = poolPsbtRaw
	}

	//save
	outAmountDe := decimal.NewFromInt(int64(req.Amount))
	coinAmountDe := decimal.NewFromInt(int64(coinAmountInt))
	coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
	coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)
	//orderId = fmt.Sprintf("%s_%s_%s_%s_%d_%d", req.Net, req.Tick, inscriptionId, req.Address, req.Amount, coinAmountInt)
	inscriptionId = strings.ReplaceAll(inscriptionId, ":", "i")
	fmt.Printf("[BID]build-inscriptionId:%s, address:%s\n", inscriptionId, buyerAddress)
	orderId = fmt.Sprintf("%s_%s_%s_%s_%d", req.Net, req.Tick, inscriptionId, buyerAddress, coinAmountInt)
	orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
	coinPrice, coinPriceDecimalNum = MakePrice(int64(coinAmountInt), int64(req.Amount))
	entityOrder = &model.OrderBrc20Model{
		Net:                 req.Net,
		OrderId:             orderId,
		Tick:                req.Tick,
		Amount:              req.Amount,
		DecimalNum:          amountDec,
		CoinAmount:          coinAmountInt,
		CoinDecimalNum:      coinDec,
		CoinRatePrice:       coinRatePrice,
		CoinPrice:           coinPrice,
		CoinPriceDecimalNum: coinPriceDecimalNum,
		OrderState:          model.OrderStatePreCreate,
		OrderType:           model.OrderTypeBuy,
		SellerAddress:       "",
		BuyerAddress:        buyerAddress,
		MarketAmount:        marketPrice,
		PlatformTx:          inscriptionTxId,
		InscriptionId:       inscriptionId,
		InscriptionNumber:   inscriptionNumber,
		PsbtRawPreAsk:       "",
		PsbtRawPreBid:       psbtRaw,
		PoolOrderId:         poolOrderId,
		PlatformDummy:       model.PlatformDummy(req.PlatformDummy),
		Timestamp:           tool.MakeTimestamp(),
		Version:             2,
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
		entityOrder                               *model.OrderBrc20Model
		err                                       error
		psbtBuilder                               *PsbtBuilder
		netParams                                 *chaincfg.Params        = GetNetParams(req.Net)
		coinRatePrice                             uint64                  = 0
		buyerAddress                              string                  = ""
		_, platformAddressReceiveBidValue         string                  = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		_, platformAddressReceiveBidValueToReturn string                  = GetPlatformKeyAndAddressReceiveBidValueToReturn(req.Net)
		utxoDummy1200BidXList                     []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		utxoDummyBidXList                         []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		coinPrice                                 int64                   = 0
		coinPriceDecimalNum                       int32                   = 0
		buyerTotalFee                             int64                   = 0
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

	insPsbtX := psbtBuilder.GetInputs()
	for k, v := range insPsbtX {
		if entityOrder.PlatformDummy == model.PlatformDummyYes {
			utxoId := fmt.Sprintf("%s_%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
			if k == 0 || k == 1 {
				dummyUtxo, _ := mongo_service.FindOrderUtxoModelByUtxorId(utxoId)
				if dummyUtxo != nil {
					utxoDummyBidXList = append(utxoDummyBidXList, dummyUtxo)
				}
			} else if k == 3 {
				dummyUtxo, _ := mongo_service.FindOrderUtxoModelByUtxorId(utxoId)
				if dummyUtxo != nil {
					utxoDummy1200BidXList = append(utxoDummy1200BidXList, dummyUtxo)
				}
			}
		}
	}

	//check platform brc20 utxo
	exchangeInput := preOutList[2]
	if exchangeInput.PreviousOutPoint.Hash.String() != entityOrder.PlatformTx {
		return "", errors.New("Wrong Psbt: No inscription input. ")
	}
	//check buyer pay utxo
	buyerInput := preOutList[3]
	if entityOrder.PlatformDummy == model.PlatformDummyYes {
		buyerInput = preOutList[4]
	}
	buyerInputTxId := buyerInput.PreviousOutPoint.Hash.String()
	buyerInputIndex := buyerInput.PreviousOutPoint.Index
	buyerInAmount := uint64(0)
	if strings.ToLower(req.Net) != "testnet" {
		if entityOrder.PlatformDummy == model.PlatformDummyYes {
			buyerInAmount = req.BuyerInValue
			buyerAddress = req.Address
		} else {
			buyerTx, err := oklink_service.GetTxDetail(buyerInputTxId)
			if err != nil {
				buyerTx, err = GetTxDetail(req.Net, buyerInputTxId)
				if err != nil {
					return "", errors.New(fmt.Sprintf("preTx of buyer not found:%s. Please wait for a block's confirmation, which should take approximately 10 to 30 minutes.", err.Error()))
				}
			}
			//buyerInAmount, _ = strconv.ParseUint(buyerTx.OutputDetails[buyerInputIndex].Amount, 10, 64)
			buyerAmountDe, err := decimal.NewFromString(buyerTx.OutputDetails[buyerInputIndex].Amount)
			if err != nil {
				return "", errors.New("Wrong Psbt: The value of buyer input decimal parse err. ")
			}
			buyerInAmount = uint64(buyerAmountDe.Mul(decimal.New(1, 8)).IntPart())
			fmt.Printf("buyerInputIndex:%d, buyerInAmount:%d, req.Amount:%d\n", buyerInputIndex, buyerInAmount, req.Amount)
			if buyerInAmount <= req.Amount {
				return "", errors.New("Wrong Psbt: The value of buyer input dose not match. Please try again or contact customer service for assistance. ")
			}
			buyerAddress = buyerTx.OutputDetails[buyerInputIndex].OutputHash
		}
	} else {
		buyerInAmount = req.BuyerInValue
		buyerAddress = req.Address
	}

	//check out: len-6 for no buyer changeWallet
	outList := psbtBuilder.GetOutputs()
	if len(outList) != 6 && len(outList) != 7 && len(outList) != 8 {
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
			return "", errors.New(fmt.Sprintf("PSBT(X): Recheck pool order is empty. Please select a different liquidity and place a new order. "))
		} else if poolOrder.PoolState != model.PoolStateAdd {
			entityOrder.OrderState = model.OrderStateErr
			_, err := mongo_service.SetOrderBrc20Model(entityOrder)
			if err != nil {
				return "", err
			}
			return "", errors.New(fmt.Sprintf("PSBT(X): Recheck pool order status not AddState. Please select a different liquidity and place a new order. "))
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

	//check out for platform return fee amount and address
	platformFeePkScript, err := AddressToPkScript(req.Net, platformAddressReceiveBidValueToReturn)
	if err != nil {
		return "", errors.New("AddressToPkScript err: " + err.Error())
	}
	platformFeeOut := outList[3]
	if uint64(platformFeeOut.Value) != 10000 {
		return "", errors.New("Wrong Psbt: wrong value of platform fee, please contact customer service. ")
	}
	platformFee := uint64(platformFeeOut.Value)
	platformFeeOut2 := outList[4]
	if uint64(platformFeeOut2.Value) != 10000 {
		return "", errors.New("Wrong Psbt: wrong value of platform fee2 please contact customer service. ")
	}
	platformFee = platformFee + uint64(platformFeeOut2.Value)
	if hex.EncodeToString(platformFeeOut.PkScript) != platformFeePkScript && hex.EncodeToString(platformFeeOut2.PkScript) != platformFeePkScript {
		return "", errors.New("Wrong Psbt: wrong address of platform fee, please contact customer service. ")
	}
	buyerTotalFee = int64(platformFee)

	//check out for buyer changeWallet
	changeAmount := uint64(0)
	if len(outList) == 7 {
		//changeOut := outList[6]
		//_, changeAddrs, _, err := txscript.ExtractPkScriptAddrs(changeOut.PkScript, netParams)
		//if err != nil {
		//	return "", errors.New("Wrong Psbt: Extract address from out for buyer changeWallet. ")
		//}
		//if changeAddrs[0].EncodeAddress() != buyerAddress {
		//	return "", errors.New("Wrong Psbt: wrong address of out for buyer changeWallet, please contact customer service. ")
		//}
		//changeAmount = uint64(changeOut.Value)
	} else if len(outList) == 8 {
		changeOut := outList[7]
		_, changeAddrs, _, err := txscript.ExtractPkScriptAddrs(changeOut.PkScript, netParams)
		if err != nil {
			return "", errors.New("Wrong Psbt: Extract address from out for buyer changeWallet. ")
		}
		if changeAddrs[0].EncodeAddress() != buyerAddress {
			return "", errors.New("Wrong Psbt: wrong address of out for buyer changeWallet, please contact customer service. ")
		}
		changeAmount = uint64(changeOut.Value)
	}

	//check amount
	buyAmount := int64(buyerInAmount) - int64(platformFee) - int64(changeAmount) - int64(req.Fee) - (600 * 2)
	if entityOrder.PlatformDummy == model.PlatformDummyYes {
		buyAmount = int64(buyerInAmount) - int64(platformFee) - int64(changeAmount) - int64(req.Fee)
	}
	fmt.Printf("buyerInAmount:%d, platformFee:%d, changeAmount:%d, req.Fee:%d\n", buyerInAmount, platformFee, changeAmount, req.Fee)
	fmt.Printf("buyAmount:%d, req.Amount:%d\n", buyAmount, req.Amount)
	if buyAmount <= 0 {
		return "", errors.New("Wrong Psbt: The purchase amount is less than 0. Please raise your order price. ")
	}
	if buyAmount != int64(req.Amount) {
		return "", errors.New("Wrong Psbt: The purchase amount dose not match, please contact customer service. ")
	}
	supplementaryAmount := int64(entityOrder.MarketAmount) - int64(buyAmount)
	fmt.Printf("supplementaryAmount:%d, entityOrder.MarketAmount:%d\n", supplementaryAmount, entityOrder.MarketAmount)
	//if supplementaryAmount < 600 {
	if supplementaryAmount <= 0 {
		return "", errors.New("Wrong Psbt: The purchase amount exceeds the market price. Please adjust your purchase price to below the market price or directly buy from the Ask orders. ")
	}

	if entityOrder.PlatformDummy == model.PlatformDummyYes {
		bidUtxo := preOutList[4]
		SaveForUserBidUtxo(entityOrder.Net, entityOrder.Tick, entityOrder.BuyerAddress, entityOrder.OrderId, bidUtxo.PreviousOutPoint.Hash.String(), int64(bidUtxo.PreviousOutPoint.Index), model.DummyStateLive)
	} else {
		for i := 0; i < 2; i++ {
			dummy := preOutList[i]
			SaveForUserBidDummy(entityOrder.Net, entityOrder.Tick, entityOrder.BuyerAddress, entityOrder.OrderId, dummy.PreviousOutPoint.Hash.String(), int64(dummy.PreviousOutPoint.Index), model.DummyStateLive)
		}
	}

	outAmountDe := decimal.NewFromInt(int64(req.Amount))
	coinAmountDe := decimal.NewFromInt(int64(entityOrder.CoinAmount))
	coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
	coinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)
	if coinRatePrice == 0 {
		coinRatePrice = 1
	}
	coinPrice, coinPriceDecimalNum = MakePrice(int64(entityOrder.CoinAmount), int64(req.Amount))
	entityOrder.PlatformFee = platformFee
	entityOrder.ChangeAmount = changeAmount
	entityOrder.Fee = req.Fee
	entityOrder.FeeRate = req.Rate
	entityOrder.SupplementaryAmount = uint64(supplementaryAmount)
	entityOrder.CoinRatePrice = coinRatePrice
	entityOrder.CoinPrice = coinPrice
	entityOrder.CoinPriceDecimalNum = coinPriceDecimalNum
	entityOrder.BuyerAddress = buyerAddress
	entityOrder.Amount = req.Amount
	entityOrder.OrderState = model.OrderStateCreate
	entityOrder.PsbtRawMidBid = req.PsbtRaw
	entityOrder.BuyerTotalFee = buyerTotalFee
	_, err = mongo_service.SetOrderBrc20Model(entityOrder)
	if err != nil {
		//fmt.Printf("SetOrderBrc20Model: %+v\n", entityOrder)
		return "", err
	}
	UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	UpdateMarketPriceV2(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))

	// set occupied
	SetOccupiedDummyUtxo(utxoDummyBidXList, entityOrder.OrderId)
	SetOccupiedDummyUtxo(utxoDummy1200BidXList, entityOrder.OrderId)

	return req.OrderId, nil
}

func CalFeeAmount(req *request.OrderBrc20CalFeeReq) (*respond.CalFeeResp, error) {
	var (
		feeAmountForReleaseInscription int64 = 5000
		feeAmountForRewardInscription  int64 = 4000
		feeAmountForRewardSend         int64 = 4000
		feeAmountForPlatform           int64 = 3000
	)
	if req.Version == 2 {
		feeAmountForReleaseInscription, feeAmountForRewardInscription, feeAmountForRewardSend = GenerateBidTakerFee(req.NetworkFeeRate)
	}
	return &respond.CalFeeResp{
		ReleaseInscriptionFee: feeAmountForReleaseInscription,
		RewardInscriptionFee:  feeAmountForRewardInscription,
		RewardSendFee:         feeAmountForRewardSend,
		PlatformFee:           feeAmountForPlatform,
	}, nil
}

func DoBid(req *request.OrderBrc20DoBidReq) (*respond.DoBidResp, error) {
	var (
		entity                *model.OrderBrc20Model
		err                   error
		psbtBuilder           *PsbtBuilder
		btcPsbtBuilder        *PsbtBuilder
		netParams             *chaincfg.Params = GetNetParams(req.Net)
		utxoDummy1200List     []*model.OrderUtxoModel
		utxoDummyList         []*model.OrderUtxoModel
		utxoBidYList          []*model.OrderUtxoModel
		utxoDummy1200BidXList []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		utxoDummyBidXList     []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		addressUtxoMap        map[string][]*wire.TxIn = make(map[string][]*wire.TxIn)

		//startIndexDummy int64 = -1
		//startIndexBidY int64 = -1
		newPsbtBuilder                                                                        *PsbtBuilder
		marketPrice                                                                           uint64 = 0
		inscriptionId                                                                         string = ""
		sellInscriptionId                                                                     string = ""
		inscriptionBrc20BalanceItem                                                           *oklink_service.BalanceItem
		coinAmount                                                                            uint64 = 0
		brc20ReceiveValue                                                                     uint64 = 0
		platformPrivateKeyReceiveBidValueToX, platformAddressReceiveBidValueToX               string = GetPlatformKeyAndAddressReceiveBidValueToX(req.Net)
		_, platformAddressReceiveBidValue                                                     string = GetPlatformKeyAndAddressReceiveBidValue(req.Net)
		platformPrivateKeyReceiveBidValueToReturn, platformAddressReceiveBidValueToReturn     string = GetPlatformKeyAndAddressReceiveBidValueToReturn(req.Net)
		_, platformAddressReceiveBrc20                                                        string = GetPlatformKeyAndAddressReceiveBrc20(req.Net)
		platformPrivateKeyReceiveDummyValue, platformAddressReceiveDummyValue                 string = GetPlatformKeyAndAddressReceiveDummyValue(req.Net)
		_, platformAddressSendBrc20                                                           string = GetPlatformKeyAndAddressSendBrc20(req.Net)
		platformPrivateKeyReceiveBidValueForPoolBtc, platformAddressReceiveBidValueForPoolBtc string = GetPlatformKeyAndAddressReceiveValueForPoolBtc(req.Net)
		platformPrivateKeyDummy, platformAddressDummy                                         string = GetPlatformKeyAndAddressForDummy(req.Net)
		_, platformAddressForMultiSigInscription                                              string = GetPlatformKeyAndAddressForMultiSigInscription(req.Net)
		_, platformAddressForRewardBrc20FeeUtxos                                              string = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(req.Net)
		_, platformAddressReceiveFee                                                          string = GetPlatformKeyAndAddressReceiveFee(req.Net)

		sellerPreBrc20Output *wire.TxIn
		sellerSendAddress    string = req.Address
		inValue              uint64 = req.Value
		sellerPayFee         uint64 = 0

		platformPayPerAmount int64  = 10000
		addressSendBrc20     string = platformAddressSendBrc20
		multiSigScript       string = ""

		inscriptionOutputValue uint64                  = 0
		poolBtcUtxoList        []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		poolBtcAmount          uint64                  = 0
		poolBtcPsbtInput       Input
		poolBtcPsbtSigIn       SigIn
		poolBtcOutput          Output

		feeOutputForReleaseInscription Output
		feeOutputForRewardInscription  Output
		feeOutputForRewardSend         Output
		feeOutputForPlatform           Output
		sellerChangeOutput             Output
		feeAmountForReleaseInscription int64 = 5000
		feeAmountForRewardInscription  int64 = 4000
		feeAmountForRewardSend         int64 = 4000
		feeAmountForPlatform           int64 = 3000
		sellerNetworkFeeAmount         int64 = 0
		sellerTotalFee                 int64 = 0

		dealTxIndex, dealTxOutValue         int64 = 0, 0 // receive btc
		dealCoinTxIndex, dealCoinTxOutValue int64 = 0, 0 // receive brc20

		bidYUtxoOffsetIndex   int   = 4
		bidYOffsetIndex       int   = 3
		poolBtcInputIndex     int   = 3
		dummy1200InputIndex   int   = 3
		newDummy600Index      int64 = 4
		newBidXUtxoOuputIndex int64 = 3

		newBidXUtxoOuputIndexForReleaseInscription int64 = 0
		newBidXUtxoOuputIndexForRewardInscription  int64 = 0
		newBidXUtxoOuputIndexForRewardSend         int64 = 0
	)
	if req.Version == 2 {
		feeAmountForReleaseInscription, feeAmountForRewardInscription, feeAmountForRewardSend = GenerateBidTakerFee(req.NetworkFeeRate)
		fmt.Printf("feeAmountForReleaseInscription:%d, feeAmountForRewardInscription:%d, feeAmountForRewardSend:%d\n", feeAmountForReleaseInscription, feeAmountForRewardInscription, feeAmountForRewardSend)
	}

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
	if len(preOutList) != 1 && len(preOutList) != 2 {
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

		sellerPreBrc20Output = preOutList[0]

		time.Sleep(1000 * time.Millisecond)

		if len(preOutList) == 2 {
			preSellFeeTx, err := oklink_service.GetTxDetail(preOutList[1].PreviousOutPoint.Hash.String())
			if err != nil {
				preSellFeeTx, err = GetTxDetail(entity.Net, preOutList[1].PreviousOutPoint.Hash.String())
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Wrong Psbt: sellPayFee input is empty preTx. [%s:%d] err:%s", preOutList[1].PreviousOutPoint.Hash.String(), preOutList[1].PreviousOutPoint.Index, err.Error()))
				}
			}
			feeInValueDe, err := decimal.NewFromString(preSellFeeTx.OutputDetails[preOutList[1].PreviousOutPoint.Index].Amount)
			if err != nil {
				return nil, errors.New("Wrong Psbt: The value of sellPayFee input decimal parse err. ")
			}
			feeInValue := uint64(feeInValueDe.Mul(decimal.New(1, 8)).IntPart())
			if feeInValue == 0 {
				return nil, errors.New("Wrong Psbt: sellPayFee out of preTx is empty amount. ")
			}
			if sellerSendAddress != preSellFeeTx.OutputDetails[preOutList[1].PreviousOutPoint.Index].OutputHash {
				return nil, errors.New("Wrong Psbt: sellPayFee out of address dose not match. ")
			}
			sellerPayFee = feeInValue
		}
	}

	if sellerSendAddress == entity.BuyerAddress {
		return nil, errors.New("You cannot sell on your own order. ")
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

	//check seller brc20 utxo
	sellerInscriptionBrc20BalanceItem, err := CheckBrc20Ordinals(sellerPreBrc20Output, entity.Tick, sellerSendAddress)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Wrong Psbt: CheckBrc20Ordinals err：%s. Please confirm that your address has BRC20 tokens and they have been inscribed. ", err.Error()))
	}
	sellInscriptionId = fmt.Sprintf("%si%d", sellerPreBrc20Output.PreviousOutPoint.Hash.String(), sellerPreBrc20Output.PreviousOutPoint.Index)

	if req.Net == "mainnet" || req.Net == "livenet" {
		if sellerInscriptionBrc20BalanceItem == nil {
			return nil, errors.New("Wrong Psbt: Empty inscription from Seller. Please confirm that your address has BRC20 tokens and they have been inscribed. ")
		}
	}

	//if seller's psbt has changeWallet out
	if len(sellOuts) == 2 {
		//add seller's changeWallet out
		sellerChangeOut := sellOuts[1]
		sellerChangeOutput = Output{
			Amount: uint64(sellerChangeOut.Value),
			Script: hex.EncodeToString(sellerChangeOut.PkScript),
		}

		if int64(sellerPayFee)-int64(sellerChangeOutput.Amount) < (feeAmountForReleaseInscription + feeAmountForRewardInscription + feeAmountForRewardSend + feeAmountForPlatform) {
			return nil, errors.New("Wrong Psbt: seller's payFee amount dose not match. ")
		}
		sellerTotalFee = int64(sellerPayFee) - int64(sellerChangeOutput.Amount)
		sellerNetworkFeeAmount = (int64(sellerPayFee) - int64(sellerChangeOutput.Amount)) - (feeAmountForReleaseInscription + feeAmountForRewardInscription + feeAmountForRewardSend + feeAmountForPlatform)
		if req.NetworkFee != 0 {
			if req.NetworkFee != sellerNetworkFeeAmount {
				fmt.Printf("sellerPayFee:%d, sellerChangeOutput.Amount:%d\n", sellerPayFee, sellerChangeOutput.Amount)
				fmt.Printf("sellerNetworkFeeAmount:%d, req.NetworkFee:%d\n", sellerNetworkFeeAmount, req.NetworkFee)
				return nil, errors.New("Wrong Psbt: seller's network Fee amount dose not match. ")
			}
		}

		//add fee output for release inscription and reward inscription and platform
		platformPkScriptForMultiSigInscription, err := AddressToPkScript(req.Net, platformAddressForMultiSigInscription)
		if err != nil {
			return nil, errors.New("AddressToPkScript err: " + err.Error())
		}
		platformPkScriptForRewardBrc20FeeUtxos, err := AddressToPkScript(req.Net, platformAddressForRewardBrc20FeeUtxos)
		if err != nil {
			return nil, errors.New("AddressToPkScript err: " + err.Error())
		}
		platformPkScriptForReceiveFee, err := AddressToPkScript(req.Net, platformAddressReceiveFee)
		if err != nil {
			return nil, errors.New("AddressToPkScript err: " + err.Error())
		}

		feeOutputForReleaseInscription = Output{
			Amount:  uint64(feeAmountForReleaseInscription),
			Script:  platformPkScriptForMultiSigInscription,
			Address: platformAddressForMultiSigInscription,
		}
		feeOutputForRewardInscription = Output{
			Amount:  uint64(feeAmountForRewardInscription),
			Script:  platformPkScriptForRewardBrc20FeeUtxos,
			Address: platformAddressForRewardBrc20FeeUtxos,
		}
		feeOutputForRewardSend = Output{
			Amount:  uint64(feeAmountForRewardSend),
			Script:  platformPkScriptForRewardBrc20FeeUtxos,
			Address: platformAddressForRewardBrc20FeeUtxos,
		}
		feeOutputForPlatform = Output{
			Amount:  uint64(feeAmountForPlatform),
			Script:  platformPkScriptForReceiveFee,
			Address: platformAddressReceiveFee,
		}

		poolBtcInputIndex = poolBtcInputIndex + 1
		bidYUtxoOffsetIndex = bidYUtxoOffsetIndex + 1
		bidYOffsetIndex = bidYOffsetIndex + 1
		dummy1200InputIndex = dummy1200InputIndex + 1
	}

	has := false
	for _, v := range preOutList {
		inscriptionId = fmt.Sprintf("%si%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
		inscriptionBrc20BalanceItem, err = CheckBrc20Ordinals(v, entity.Tick, sellerSendAddress)
		if err != nil {
			continue
		}
		has = true
		break
	}
	_ = inscriptionId

	if req.Net == "mainnet" || req.Net == "livenet" {
		if !has || inscriptionBrc20BalanceItem == nil {
			return nil, errors.New("Wrong Psbt: Empty inscription. Please use a valid BRC20 token or re-inscribe the BRC20 token. ")
		}
		coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
	} else {
		coinAmount, _ = strconv.ParseUint(req.CoinAmount, 10, 64)
	}
	if coinAmount != entity.CoinAmount {
		return nil, errors.New("Wrong Psbt: brc20 coin amount dose not match. Please use a valid BRC20 token or re-inscribe the BRC20 token. ")
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
			return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck pool order is empty. Please select a different liquidity and place a new order. "))
		} else if poolOrder.PoolState != model.PoolStateAdd {
			entity.OrderState = model.OrderStateErr
			_, err := mongo_service.SetOrderBrc20Model(entity)
			if err != nil {
				return nil, err
			}
			return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck pool order status not AddState. Please select a different liquidity and place a new order. "))
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
					//check btc psbt PreOut
					liveBtcUtxoList := make([]*oklink_service.UtxoItem, 0)
					utxoResp, err := oklink_service.GetAddressUtxo(poolOrder.Address, 1, 100)
					if err != nil {
						return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck address utxo list for pool err:%s", err.Error()))
					}
					if utxoResp.UtxoList != nil && len(utxoResp.UtxoList) != 0 {
						liveBtcUtxoList = append(liveBtcUtxoList, utxoResp.UtxoList...)
					}
					utxoList, err := unisat_service.GetAddressUtxo(poolOrder.Address)
					_ = err
					//if err != nil {
					//	return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck address utxo list err:%s", err.Error()))
					//}
					if utxoList != nil && len(utxoList) != 0 {
						for _, uu := range utxoList {
							liveBtcUtxoList = append(liveBtcUtxoList, &oklink_service.UtxoItem{
								TxId:          uu.TxId,
								Index:         strconv.FormatInt(uu.OutputIndex, 10),
								Height:        "",
								BlockTime:     "",
								Address:       uu.ScriptPk,
								UnspentAmount: strconv.FormatInt(uu.Satoshis, 10),
							})
						}
					}

					hasPoolUtxo := false
					poolBtcUtxoId := fmt.Sprintf("%s_%d", poolBtcPsbtInput.OutTxId, poolBtcPsbtInput.OutIndex)
					for _, utxo := range liveBtcUtxoList {
						uId := fmt.Sprintf("%s_%s", utxo.TxId, utxo.Index)
						if uId == poolBtcUtxoId {
							hasPoolUtxo = true
							break
						}
					}
					if !hasPoolUtxo {
						//set bid order err
						entity.OrderState = model.OrderStateErr
						_, err := mongo_service.SetOrderBrc20Model(entity)
						if err != nil {
							return nil, err
						}
						if entity.PlatformDummy == model.PlatformDummyYes {
							dummyUtxoList, _ := mongo_service.FindOccupiedUtxoListByOrderId(entity.Net, entity.OrderId, 1000, model.UsedOccupied)
							ReleaseOccupiedDummyUtxo(dummyUtxoList)
							UpdateForOrderLiveUtxo(entity.OrderId, model.DummyStateFinish)
						}

						//set pool order err
						poolOrder.PoolState = model.PoolStateRemove
						poolOrder.PoolCoinState = model.PoolStateRemove
						_, err = mongo_service.SetPoolBrc20Model(poolOrder)
						if err != nil {
							return nil, err
						}
						updatePoolInfo(poolOrder)
						return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck address utxo list, pool utxo had been spent: %s. Please select a different liquidity and place a new order. ", poolBtcUtxoId))
					}

					finalScriptWitness := btcPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].FinalScriptWitness
					witnessUtxo := btcPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].WitnessUtxo
					sighashType := btcPsbtBuilder.PsbtUpdater.Upsbt.Inputs[0].SighashType
					poolBtcPsbtSigIn = SigIn{
						WitnessUtxo:        witnessUtxo,
						SighashType:        sighashType,
						FinalScriptWitness: finalScriptWitness,
						Index:              poolBtcInputIndex,
					}
					bidYUtxoOffsetIndex = bidYUtxoOffsetIndex + 1
					bidYOffsetIndex = bidYOffsetIndex + 1
					dummy1200InputIndex = dummy1200InputIndex + 1
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

	utxoDummy1200List, err = GetUnoccupiedUtxoList(req.Net, 1, 0, model.UtxoTypeDummy1200, "", 0)
	defer ReleaseUtxoList(utxoDummy1200List)
	if err != nil {
		return nil, err
	}

	utxoDummyList, err = GetUnoccupiedUtxoList(req.Net, 2, 0, model.UtxoTypeDummy, "", 0)
	defer ReleaseUtxoList(utxoDummyList)
	if err != nil {
		return nil, err
	}

	//get bidY pay utxo
	supplementaryAmount := entity.SupplementaryAmount
	if supplementaryAmount <= 600 {
		supplementaryAmount = 600
	}
	networkFee := entity.Fee
	if sellerNetworkFeeAmount != 0 {
		networkFee = uint64(sellerNetworkFeeAmount)
	}

	totalNeedAmount := int64(supplementaryAmount+sellerReceiveValue+networkFee+inscriptionOutputValue) - sellerNetworkFeeAmount
	limit := totalNeedAmount/platformPayPerAmount + 1
	changeAmount := platformPayPerAmount*limit - totalNeedAmount
	fmt.Printf("[DO][not-pool]totalNeedAmount: %d， supplementaryAmount：%d, sellerReceiveValue: %d, entity.Fee: %d, inscriptionOutputValue: %d, sellerNetworkFeeAmount:%d\n", totalNeedAmount, supplementaryAmount, sellerReceiveValue, entity.Fee, inscriptionOutputValue, sellerNetworkFeeAmount)
	if poolBtcAmount != 0 {
		totalNeedAmount = totalNeedAmount - int64(poolBtcAmount)
		fmt.Printf("[DO][pool]totalNeedAmount: %d， poolBtcAmount：%d\n", totalNeedAmount, poolBtcAmount)
		limit = totalNeedAmount/platformPayPerAmount + 1
		changeAmount = platformPayPerAmount*limit - totalNeedAmount
	}
	fmt.Printf("[DO]totalNeedAmount: %d， poolBtcAmount：%d, changeAmount: %d, sellerNetworkFeeAmount: %d\n", totalNeedAmount, poolBtcAmount, changeAmount, sellerNetworkFeeAmount)
	if totalNeedAmount <= 0 {
		return nil, errors.New("Wrong Psbt: totalNeedAmount is less than 0. ")
	}

	utxoBidYList, err = GetUnoccupiedUtxoList(req.Net, int64(limit), int64(totalNeedAmount), model.UtxoTypeBidY, "", 0)
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

	//add seller's payFee ins - index: 3
	if len(preOutList) == 2 {
		inputs = append(inputs, Input{
			OutTxId:  preOutList[1].PreviousOutPoint.Hash.String(),
			OutIndex: preOutList[1].PreviousOutPoint.Index,
		})
	}

	////add dummy1200 ins - index: 3
	//if feeOutput.Script != "" {
	//	for _, dummy := range utxoDummy1200List {
	//		inputs = append(inputs, Input{
	//			OutTxId:  dummy.TxId,
	//			OutIndex: uint32(dummy.Index),
	//		})
	//	}
	//}

	//add btc pool psbt ins - index: 3/4
	if poolBtcPsbtInput.OutTxId != "" {
		inputs = append(inputs, poolBtcPsbtInput)
	}

	//add dummy1200 outs - idnex: 4/5
	for _, dummy := range utxoDummy1200List {
		inputs = append(inputs, Input{
			OutTxId:  dummy.TxId,
			OutIndex: uint32(dummy.Index),
		})
	}

	//add Exchange pay value ins - index: 3,3+/4,4+/5,5+/6,6+
	for _, payBid := range utxoBidYList {
		inputs = append(inputs, Input{
			OutTxId:  payBid.TxId,
			OutIndex: uint32(payBid.Index),
		})
	}

	//add dummy outs - idnex: 0
	newDummy1200Out := Output{
		Address: platformAddressReceiveDummyValue,
		Amount:  dummyOutValue,
	}
	outputs = append(outputs, newDummy1200Out)
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

	//add receive seller change outs - idnex: 3
	if sellerChangeOutput.Script != "" || sellerChangeOutput.Address != "" {
		outputs = append(outputs, sellerChangeOutput)
		newDummy600Index++
	}

	//add receive pool btc outs - idnex: 3/4
	if poolBtcOutput.Script != "" || poolBtcOutput.Address != "" {
		outputs = append(outputs, poolBtcOutput)
		newDummy600Index++
	}

	_ = marketPrice
	//add receive exchange psbtX outs - idnex: 3/4/5
	//psbtXValue := entity.MarketAmount - sellerReceiveValue
	psbtXValue := supplementaryAmount
	exchangePsbtXOut := Output{
		Address: platformAddressReceiveBidValueToX,
		Amount:  psbtXValue,
	}
	outputs = append(outputs, exchangePsbtXOut)
	//add new dummy outs - idnex: 4,5 / 5,6 / 6,7
	newDummyOut := Output{
		Address: newDummyOutSegwitAddress,
		Amount:  600,
	}
	outputs = append(outputs, newDummyOut)
	outputs = append(outputs, newDummyOut)

	//add fee output for release inscription and reward inscription
	if feeOutputForReleaseInscription.Script != "" || feeOutputForReleaseInscription.Address != "" {
		outputs = append(outputs, feeOutputForReleaseInscription)
		newBidXUtxoOuputIndexForReleaseInscription = int64(len(outputs)) - 1
	}
	if feeOutputForRewardInscription.Script != "" || feeOutputForRewardInscription.Address != "" {
		outputs = append(outputs, feeOutputForRewardInscription)
		newBidXUtxoOuputIndexForRewardInscription = int64(len(outputs)) - 1
	}
	if feeOutputForRewardSend.Script != "" || feeOutputForRewardSend.Address != "" {
		outputs = append(outputs, feeOutputForRewardSend)
		newBidXUtxoOuputIndexForRewardSend = int64(len(outputs)) - 1
	}
	if feeOutputForPlatform.Script != "" || feeOutputForPlatform.Address != "" {
		outputs = append(outputs, feeOutputForPlatform)
	}

	if changeAmount >= 546 {
		outputs = append(outputs, Output{
			Address: platformAddressReceiveBidValue,
			Amount:  uint64(changeAmount),
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

	if len(preOutList) == 2 {
		finalScriptWitness2 := psbtBuilder.PsbtUpdater.Upsbt.Inputs[1].FinalScriptWitness
		witnessUtxo2 := psbtBuilder.PsbtUpdater.Upsbt.Inputs[1].WitnessUtxo
		sighashType2 := psbtBuilder.PsbtUpdater.Upsbt.Inputs[1].SighashType
		err = newPsbtBuilder.AddSigIn(witnessUtxo2, sighashType2, finalScriptWitness2, 3)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("PSBT(Y): AddPartialSigIn2 err:%s", err.Error()))
		}
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

	//add dummy ins sign - index: 3
	for k, dummy := range utxoDummy1200List {
		inSigns = append(inSigns, &InputSign{
			Index:       k + dummy1200InputIndex,
			PkScript:    dummy.PkScript,
			Amount:      dummy.Amount,
			SighashType: txscript.SigHashAll,
			PriHex:      dummy.PrivateKeyHex,
			UtxoType:    Witness,
		})
	}

	//add Exchange pay value ins - index: 3,3+ / 4,4+ / 5,5+
	for k, payBid := range utxoBidYList {
		inSigns = append(inSigns, &InputSign{
			Index:       k + bidYUtxoOffsetIndex,
			PkScript:    payBid.PkScript,
			Amount:      payBid.Amount,
			SighashType: txscript.SigHashAll,
			PriHex:      payBid.PrivateKeyHex,
			UtxoType:    Witness,
		})
	}
	//
	//txPsbt_t := newPsbtBuilder.PsbtUpdater.Upsbt.UnsignedTx
	//fmt.Printf("Tx:\n")
	//fmt.Printf("%+v\n", txPsbt_t)
	//for _, in := range txPsbt_t.TxIn {
	//	fmt.Printf("%+v\n", *in)
	//}
	//for _, out := range txPsbt_t.TxOut {
	//	fmt.Printf("%+v\n", *out)
	//}
	//fmt.Printf("\n")

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
	addressUtxoMap[entity.BuyerAddress] = make([]*wire.TxIn, 0)
	addressUtxoMap[addressSendBrc20] = make([]*wire.TxIn, 0)
	for k, v := range insPsbtX {
		if entity.PlatformDummy == model.PlatformDummyYes {
			if k == 2 {
				addressUtxoMap[addressSendBrc20] = append(addressUtxoMap[addressSendBrc20], v)
			} else if k == 4 {
				addressUtxoMap[entity.BuyerAddress] = append(addressUtxoMap[entity.BuyerAddress], v)
			}

			utxoId := fmt.Sprintf("%s_%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
			if k == 0 || k == 1 {
				dummyUtxo, _ := mongo_service.FindOrderUtxoModelByUtxorId(utxoId)
				if dummyUtxo != nil {
					if dummyUtxo.ConfirmStatus == model.Unconfirmed {
						return nil, errors.New(fmt.Sprintf("PSBT(X):dummy Utxo still not confirmed. Please wait for the confirmation of the dummy Utxo. "))
					}
					utxoDummyBidXList = append(utxoDummyBidXList, dummyUtxo)
				}
			} else if k == 3 {
				dummyUtxo, _ := mongo_service.FindOrderUtxoModelByUtxorId(utxoId)
				if dummyUtxo != nil {
					if dummyUtxo.ConfirmStatus == model.Unconfirmed {
						return nil, errors.New(fmt.Sprintf("PSBT(X):dummy Utxo still not confirmed. Please wait for the confirmation of the dummy Utxo. "))
					}
					utxoDummy1200BidXList = append(utxoDummy1200BidXList, dummyUtxo)
				}
			}
		} else {
			if k == 2 {
				addressUtxoMap[addressSendBrc20] = append(addressUtxoMap[addressSendBrc20], v)
			} else {
				addressUtxoMap[entity.BuyerAddress] = append(addressUtxoMap[entity.BuyerAddress], v)
			}
		}
	}
	liveUtxoList := make([]*oklink_service.UtxoItem, 0)
	//liveUtxoList := make([]*unisat_service.UtxoDetailItem, 0)
	if entity.Net != "testnet" {
		for address, _ := range addressUtxoMap {
			fmt.Printf("[DO][Check live utxo] address:%s\n", address)
			utxoResp, err := oklink_service.GetAddressUtxo(address, 1, 100)
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
					height := "0"
					confirmation := GetTxConfirm(uu.TxId)
					if confirmation > 0 {
						height = "1"
					}
					liveUtxoList = append(liveUtxoList, &oklink_service.UtxoItem{
						TxId:          uu.TxId,
						Index:         strconv.FormatInt(uu.OutputIndex, 10),
						Height:        height,
						BlockTime:     "",
						Address:       uu.ScriptPk,
						UnspentAmount: strconv.FormatInt(uu.Satoshis, 10),
					})
				}
			}

			utxoInscription, err := unisat_service.GetAddressInscriptions(address)
			if utxoInscription != nil && len(utxoInscription) != 0 {
				for _, ui := range utxoInscription {
					output := ui.Output
					outputStrs := strings.Split(output, ":")
					if len(outputStrs) <= 2 {
						continue
					}
					height := "0"
					confirmation := ui.UtxoConfirmation
					if confirmation > 0 {
						height = "1"
					}
					liveUtxoList = append(liveUtxoList, &oklink_service.UtxoItem{
						TxId:          outputStrs[0],
						Index:         outputStrs[1],
						Height:        height,
						BlockTime:     "",
						Address:       ui.Address,
						UnspentAmount: strconv.FormatInt(ui.OutputValue, 10),
					})
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}

	for k, v := range insPsbtX {
		if entity.PlatformDummy == model.PlatformDummyYes {
			if k != 2 && k != 4 {
				continue
			}
		}

		bidInId := fmt.Sprintf("%s_%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
		has := false
		for _, u := range liveUtxoList {
			uId := fmt.Sprintf("%s_%s", u.TxId, u.Index)
			//uId := fmt.Sprintf("%s_%d", u.TxId, u.OutputIndex)
			fmt.Printf("liveUtxo:[%s]\n", uId)
			if bidInId == uId {
				if u.Height == "0" || u.Height == "" {
					return nil, errors.New(fmt.Sprintf("PSBT(X):buyer or brc20 Utxo still not confirmed. Please wait for the confirmation of the Utxo. "))
				}
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
			if entity.PlatformDummy == model.PlatformDummyYes {
				dummyUtxoList, _ := mongo_service.FindOccupiedUtxoListByOrderId(entity.Net, entity.OrderId, 1000, model.UsedOccupied)
				ReleaseOccupiedDummyUtxo(dummyUtxoList)
				UpdateForOrderLiveUtxo(entity.OrderId, model.DummyStateCancel)
			}
			return nil, errors.New(fmt.Sprintf("PSBT(X): Recheck address utxo list, utxo had been spent: %s. Please select a different liquidity and place a new order. ", bidInId))
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
	entity.SellInscriptionId = sellInscriptionId
	entity.SellerTotalFee = sellerTotalFee
	entity.BidValueToXUtxoId = fmt.Sprintf("%s_%d", psbtYTxId, bidYOffsetIndex)
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return nil, err
	}

	//check fee output for new bidUtxo from X psbt
	//check out for platform return fee amount and address
	platformFeePkScript, err := AddressToPkScript(req.Net, platformAddressReceiveBidValueToReturn)
	if err != nil {
		return nil, errors.New("AddressToPkScript err: " + err.Error())
	}
	platformDummyPkScript, err := AddressToPkScript(req.Net, platformAddressDummy)
	if err != nil {
		return nil, errors.New("AddressToPkScript err: " + err.Error())
	}
	newBidXUtxoOuts := make([]Output, 0)
	newBidXDummy1200UtxoOuts := make([]Output, 0)
	newBidXDummy600UtxoOuts := make([]Output, 0)
	bidXOutputs := bidPsbtBuilder.GetOutputs()

	newBidXDummy1200Out := bidXOutputs[0]
	if newBidXDummy1200Out.Value == 1200 && hex.EncodeToString(newBidXDummy1200Out.PkScript) == platformDummyPkScript {
		newBidXDummy1200UtxoOuts = append(newBidXDummy1200UtxoOuts, Output{
			Address: platformAddressDummy,
			Amount:  uint64(newBidXDummy1200Out.Value),
		})
	}
	newBidXDummy600Out1 := bidXOutputs[5]
	if newBidXDummy600Out1.Value == 600 && hex.EncodeToString(newBidXDummy600Out1.PkScript) == platformDummyPkScript {
		newBidXDummy600UtxoOuts = append(newBidXDummy600UtxoOuts, Output{
			Address: platformAddressDummy,
			Amount:  uint64(newBidXDummy600Out1.Value),
		})
	}
	newBidXDummy600Out2 := bidXOutputs[6]
	if newBidXDummy600Out2.Value == 600 && hex.EncodeToString(newBidXDummy600Out2.PkScript) == platformDummyPkScript {
		newBidXDummy600UtxoOuts = append(newBidXDummy600UtxoOuts, Output{
			Address: platformAddressDummy,
			Amount:  uint64(newBidXDummy600Out2.Value),
		})
	}

	if len(bidXOutputs) == 7 {
		newBidXFeeOut := bidXOutputs[3]
		if newBidXFeeOut.Value == 10000 && hex.EncodeToString(newBidXFeeOut.PkScript) == platformFeePkScript {
			newBidXUtxoOuts = append(newBidXUtxoOuts, Output{
				Address: platformAddressReceiveBidValueToReturn,
				Amount:  uint64(newBidXFeeOut.Value),
			})
		}
		newBidXFeeOut2 := bidXOutputs[4]
		if newBidXFeeOut2.Value == 10000 && hex.EncodeToString(newBidXFeeOut2.PkScript) == platformFeePkScript {
			newBidXUtxoOuts = append(newBidXUtxoOuts, Output{
				Address: platformAddressReceiveBidValueToReturn,
				Amount:  uint64(newBidXFeeOut2.Value),
			})
		}
	} else if len(bidXOutputs) == 8 {
		newBidXFeeOut := bidXOutputs[3]
		if newBidXFeeOut.Value == 10000 && hex.EncodeToString(newBidXFeeOut.PkScript) == platformFeePkScript {
			newBidXUtxoOuts = append(newBidXUtxoOuts, Output{
				Address: platformAddressReceiveBidValueToReturn,
				Amount:  uint64(newBidXFeeOut.Value),
			})
		}
		newBidXFeeOut2 := bidXOutputs[4]
		if newBidXFeeOut2.Value == 10000 && hex.EncodeToString(newBidXFeeOut2.PkScript) == platformFeePkScript {
			newBidXUtxoOuts = append(newBidXUtxoOuts, Output{
				Address: platformAddressReceiveBidValueToReturn,
				Amount:  uint64(newBidXFeeOut2.Value),
			})
		}
	}

	fmt.Printf("PsbtY:%s\n", txRawPsbtY)
	fmt.Printf("psbtYTxId: %s\n", psbtYTxId)
	fmt.Printf("PsbtX:%s\n", txRawPsbtX)
	fmt.Printf("psbtXTxId: %s\n", psbtXTxId)

	txPsbtXRespTxId := ""
	txPsbtYRespTxId := ""
	entity.DealTime = tool.MakeTimestamp()

	if req.Net == "mainnet" || req.Net == "livenet" {
		txPsbtYResp, err := unisat_service.BroadcastTx(req.Net, txRawPsbtY)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s. Please try again or contact customer service for assistance. ", req.Net, err.Error()))
		}
		SetUsedDummyUtxo(utxoDummy1200List, txPsbtYResp.Result)
		SetUsedDummyUtxo(utxoDummyList, txPsbtYResp.Result)
		setUsedBidYUtxo(utxoBidYList, txPsbtYResp.Result)

		SaveNewDummy1200FromBid(req.Net, newDummy1200Out, platformPrivateKeyReceiveDummyValue, 0, psbtYTxId)
		SaveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, newDummy600Index, psbtYTxId)
		SaveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, newDummy600Index+1, psbtYTxId)
		SaveNewUtxoFromBid(req.Net, feeOutputForReleaseInscription, "", newBidXUtxoOuputIndexForReleaseInscription, psbtYTxId, model.UtxoTypeMultiInscription, entity.PoolOrderId, req.NetworkFeeRate)
		SaveNewUtxoFromBid(req.Net, feeOutputForRewardInscription, "", newBidXUtxoOuputIndexForRewardInscription, psbtYTxId, model.UtxoTypeRewardInscription, entity.PoolOrderId, req.NetworkFeeRate)
		SaveNewUtxoFromBid(req.Net, feeOutputForRewardSend, "", newBidXUtxoOuputIndexForRewardSend, psbtYTxId, model.UtxoTypeRewardSend, entity.PoolOrderId, req.NetworkFeeRate)

		if entity.PoolOrderId != "" {
			setCoinStatusPoolBrc20Order(entity, model.PoolStateUsed, dealCoinTxIndex, dealCoinTxOutValue, entity.DealTime)
		}

		time.Sleep(800 * time.Millisecond)
		txPsbtXResp, err := unisat_service.BroadcastTx(req.Net, txRawPsbtX)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(X) %s err:%s. Please try again or contact customer service for assistance. ", req.Net, err.Error()))
		}
		SetUsedDummyUtxo(utxoDummyBidXList, txPsbtXResp.Result)     // set used dummy600 utxo for bidX
		SetUsedDummyUtxo(utxoDummy1200BidXList, txPsbtXResp.Result) // set used dummy1200 utxo for bidX

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
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s. Please try again or contact customer service for assistance. ", req.Net, err.Error()))
		}
		SetUsedDummyUtxo(utxoDummy1200List, txPsbtYResp.Result)
		SetUsedDummyUtxo(utxoDummyList, txPsbtYResp.Result)
		setUsedBidYUtxo(utxoBidYList, txPsbtYResp.Result)

		SaveNewDummy1200FromBid(req.Net, newDummy1200Out, platformPrivateKeyReceiveDummyValue, 0, psbtYTxId)
		SaveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, newDummy600Index, psbtYTxId)
		SaveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, newDummy600Index+1, psbtYTxId)
		SaveNewUtxoFromBid(req.Net, feeOutputForReleaseInscription, "", newBidXUtxoOuputIndexForReleaseInscription, psbtYTxId, model.UtxoTypeMultiInscription, entity.PoolOrderId, req.NetworkFeeRate)
		SaveNewUtxoFromBid(req.Net, feeOutputForRewardInscription, "", newBidXUtxoOuputIndexForRewardInscription, psbtYTxId, model.UtxoTypeRewardInscription, entity.PoolOrderId, req.NetworkFeeRate)
		SaveNewUtxoFromBid(req.Net, feeOutputForRewardSend, "", newBidXUtxoOuputIndexForRewardSend, psbtYTxId, model.UtxoTypeRewardSend, entity.PoolOrderId, req.NetworkFeeRate)

		time.Sleep(800 * time.Millisecond)
		txPsbtXResp, err := unisat_service.BroadcastTx(req.Net, txRawPsbtX)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Broadcast Psbt(X) %s err:%s. Please try again or contact customer service for assistance. ", req.Net, err.Error()))
		}
		SetUsedDummyUtxo(utxoDummyBidXList, txPsbtXResp.Result)     // set used dummy600 utxo for bidX
		SetUsedDummyUtxo(utxoDummy1200BidXList, txPsbtXResp.Result) // set used dummy1200 utxo for bidX

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

	UpdateForOrderLiveUtxo(entity.OrderId, model.DummyStateFinish)
	for k, v := range newBidXUtxoOuts {
		SaveNewBidYUtxo10000FromBid(req.Net, v, platformPrivateKeyReceiveBidValueToReturn, newBidXUtxoOuputIndex+int64(k), psbtXTxId)
	}
	if entity.PlatformDummy == model.PlatformDummyYes {
		for k, v := range newBidXDummy1200UtxoOuts {
			SaveNewDummy1200FromBidX(req.Net, v, platformPrivateKeyDummy, int64(k), psbtXTxId)
		}
		for k, v := range newBidXDummy600UtxoOuts {
			SaveNewDummyFromBidX(req.Net, v, platformPrivateKeyDummy, int64(k+5), psbtXTxId)
		}
	}

	//SaveNewDummy1200FromBid(req.Net, newDummy1200Out, platformPrivateKeyReceiveDummyValue, 0, psbtYTxId)
	//SaveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, newDummy600Index, psbtYTxId)
	//SaveNewDummyFromBid(req.Net, newDummyOut, newDummyOutPriKeyHex, newDummy600Index+1, psbtYTxId)
	//SaveNewUtxoFromBid(req.Net, feeOutputForReleaseInscription, "", newBidXUtxoOuputIndexForReleaseInscription, psbtYTxId, model.UtxoTypeMultiInscription)
	//SaveNewUtxoFromBid(req.Net, feeOutputForRewardInscription, "", newBidXUtxoOuputIndexForRewardInscription, psbtYTxId, model.UtxoTypeRewardInscription)
	//SaveNewUtxoFromBid(req.Net, feeOutputForRewardInscription, "", newBidXUtxoOuputIndexForRewardSend, psbtYTxId, model.UtxoTypeRewardSend)

	entity.OrderState = model.OrderStateFinish
	entity.DealTxBlockState = model.ClaimTxBlockStateUnconfirmed
	_, err = mongo_service.SetOrderBrc20Model(entity)
	if err != nil {
		return nil, err
	}

	//todo use DB-Transaction
	if entity.PoolOrderId != "" {
		setStatusPoolBrc20Order(entity, model.PoolStateUsed, dealTxIndex, dealTxOutValue, entity.DealTime)
		setCoinStatusPoolBrc20Order(entity, model.PoolStateUsed, dealCoinTxIndex, dealCoinTxOutValue, entity.DealTime)
		removeInvalidBidByPoolOrderId(entity.PoolOrderId)
		if req.Version == 2 {
			mongo_service.SetPoolBrc20ModelForVersion(entity.PoolOrderId, req.Version)
		}
		//lp which is used
		UpdateForOrderLiveUtxo(entity.PoolOrderId, model.DummyStateFinish)
	}

	UpdateMarketPrice(req.Net, entity.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(entity.Tick)))
	UpdateMarketPriceV2(req.Net, entity.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(entity.Tick)))

	AddNotificationForOrderFinish(entity.BuyerAddress)

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
				finalAskPsbtRaw, err := finalAskPsbtBuilder.ToString()
				if err != nil {
					return "", errors.New(fmt.Sprintf("PSBT(X): ToString err:%s", err.Error()))
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
						buyerTx, err = GetTxDetail(req.Net, buyerInputTxId)
						if err != nil {
							return "", errors.New(fmt.Sprintf("preTx of buyer not found:%s. Please wait for a block's confirmation, which should take approximately 10 to 30 minutes. ", err.Error()))
						}
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
						fmt.Printf("[ERR][ask][%s]-%s-%s\n", entityOrder.OrderId, entityOrder.InscriptionId, entityOrder.SellerAddress)
						if (strings.Contains(err.Error(), "missingorspent") ||
							strings.Contains(err.Error(), "mempool-conflict")) &&
							!CheckInscriptionExist(entityOrder.SellerAddress, entityOrder.InscriptionId) {

							entityOrder.OrderState = model.OrderStateErr
							_, err := mongo_service.SetOrderBrc20Model(entityOrder)
							if err != nil {
								return "", err
							}
							UpdateMarketPrice(req.Net, entityOrder.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(entityOrder.Tick)))
							UpdateMarketPriceV2(req.Net, entityOrder.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(entityOrder.Tick)))
						}
						return "", errors.New(fmt.Sprintf("Broadcast Psbt %s, orderId-%s err:%s", entityOrder.Net, entityOrder.OrderId, err.Error()))
					}

					entityOrder.PsbtAskTxId = txPsbtResp.Result
					entityOrder.OrderState = model.OrderStateFinish

					if entityOrder.PlatformDummy == model.PlatformDummyYes {
						UpdateAndNewDummyForAsk(entityOrder.Net, finalAskPsbtRaw, txPsbtResp.Result)
					}

					//setWhitelist(entityOrder.BuyerAddress, model.WhitelistTypeClaim, 1, 0)
				}
				AddNotificationForOrderFinish(entityOrder.SellerAddress)
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
			UpdateForOrderLiveUtxo(entityOrder.OrderId, state)
			if entityOrder.PlatformDummy == model.PlatformDummyYes {
				dummyUtxoList, _ := mongo_service.FindOccupiedUtxoListByOrderId(entityOrder.Net, entityOrder.OrderId, 1000, model.UsedOccupied)
				ReleaseOccupiedDummyUtxo(dummyUtxoList)
			}
			break
		}
		entityOrder.DealTime = tool.MakeTimestamp()
		_, err = mongo_service.SetOrderBrc20Model(entityOrder)
		if err != nil {
			return "", err
		}
		UpdateMarketPrice(req.Net, entityOrder.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(entityOrder.Tick)))
		UpdateMarketPriceV2(req.Net, entityOrder.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(entityOrder.Tick)))
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
