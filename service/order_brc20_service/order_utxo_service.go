package order_brc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/create_key"
	"ordbook-aggregation/service/inscription_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
)

func ColdDownUtxo(req *request.ColdDownUtxo) (string, error) {
	var (
		netParams                        *chaincfg.Params = GetNetParams(req.Net)
		err                              error
		fromPriKeyHex, fromSegwitAddress string = "", ""
		txRaw                            string = ""
		//latestUtxo *model.OrderUtxoModel
		utxoList   []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		startIndex int64                   = GetSaveStartIndex(req.Net, req.UtxoType, int64(req.PerAmount))
	)

	if req.UtxoType == model.UtxoTypeMultiInscription {
		_, fromSegwitAddress = GetPlatformKeyAndAddressForMultiSigInscription(req.Net)
	} else if req.UtxoType == model.UtxoTypeRewardInscription {
		_, fromSegwitAddress = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(req.Net)
	} else if req.UtxoType == model.UtxoTypeRewardSend {
		_, fromSegwitAddress = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(req.Net)
	} else if req.UtxoType == model.UtxoTypeDummy1200 {
		fromPriKeyHex, fromSegwitAddress = GetPlatformKeyAndAddressReceiveDummyValue(req.Net)
	} else {
		fromPriKeyHex, fromSegwitAddress, err = create_key.CreateSegwitKey(netParams)
		if err != nil {
			return "", err
		}
	}

	inputs := make([]*TxInputUtxo, 0)
	inputs = append(inputs, &TxInputUtxo{
		TxId:     req.TxId,
		TxIndex:  req.Index,
		PkScript: req.PkScript,
		Amount:   req.Amount,
		PriHex:   req.PriKeyHex,
	})
	addr, err := btcutil.DecodeAddress(fromSegwitAddress, netParams)
	if err != nil {
		return "", err
	}
	//addrHash, err := btcutil.NewAddressWitnessPubKeyHash(addr.ScriptAddress(), netParams)
	//if err != nil {
	//	fmt.Printf("NewAddressWitnessPubKeyHash err: %s\n", err.Error())
	//	return "", err
	//}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}
	pkScript := hex.EncodeToString(pkScriptByte)
	//count := req.Amount/req.PerAmount
	outputs := make([]*TxOutput, 0)
	for i := int64(0); i < req.Count; i++ {
		outputs = append(outputs, &TxOutput{
			Address: fromSegwitAddress,
			Amount:  int64(req.PerAmount),
		})

		utxoList = append(utxoList, &model.OrderUtxoModel{
			//UtxoId:     "",
			Net:           req.Net,
			UtxoType:      req.UtxoType,
			Amount:        req.PerAmount,
			Address:       fromSegwitAddress,
			PrivateKeyHex: fromPriKeyHex,
			TxId:          "",
			Index:         i,
			PkScript:      pkScript,
			UsedState:     model.UsedNo,
			//UseTx:      "",
			SortIndex: startIndex + i + 1,
			Timestamp: tool.MakeTimestamp(),
		})
	}

	if req.ChangeAddress == "" {
		req.ChangeAddress = req.Address
	}
	tx, err := BuildCommonTx(netParams, inputs, outputs, req.ChangeAddress, req.FeeRate)
	if err != nil {
		fmt.Printf("BuildCommonTx err:%s\n", err.Error())
		return "", err
	}
	txRaw, err = ToRaw(tx)
	if err != nil {
		return "", err
	}
	for _, u := range utxoList {
		u.TxId = tx.TxHash().String()
		u.UtxoId = fmt.Sprintf("%s_%d", u.TxId, u.Index)

		_, err := mongo_service.SetOrderUtxoModel(u)
		if err != nil {
			major.Println(fmt.Sprintf("SetOrderUtxoModel for cold down err:%s", err.Error()))
			return "", err
		}
	}

	txId := ""
	//if req.Net == "testnet" {
	//	txResp, err := mempool_space_service.BroadcastTx(req.Net, txRaw)
	//	if err != nil {
	//		return "", err
	//	}
	//	txId = txResp
	//}else {
	//	txResp, err := oklink_service.BroadcastTx(txRaw)
	//	if err != nil {
	//		return "", err
	//	}
	//	txId = txResp.TxId
	//}

	txResp, err := unisat_service.BroadcastTx(req.Net, txRaw)
	if err != nil {
		return "", err
	}
	txId = txResp.Result

	//txResp, err := node.BroadcastTx(req.Net, txRaw)
	//if err != nil {
	//	return "", err
	//}
	//txId = txResp

	return txId, nil
}

func SaveNewDummyFromBid(net string, out Output, priKeyHex string, index int64, txId string) error {
	startIndex := GetSaveStartIndex(net, model.UtxoTypeDummy, 0)
	netParams := GetNetParams(net)
	addr, err := btcutil.DecodeAddress(out.Address, netParams)
	if err != nil {
		return err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return err
	}
	pkScript := hex.EncodeToString(pkScriptByte)

	newDummy := &model.OrderUtxoModel{
		UtxoId:        fmt.Sprintf("%s_%d", txId, index),
		Net:           net,
		UtxoType:      model.UtxoTypeDummy,
		Amount:        out.Amount,
		Address:       out.Address,
		PrivateKeyHex: priKeyHex,
		TxId:          txId,
		Index:         index,
		PkScript:      pkScript,
		UsedState:     model.UsedNo,
		SortIndex:     startIndex + 1,
		Timestamp:     tool.MakeTimestamp(),
	}

	_, err = mongo_service.SetOrderUtxoModel(newDummy)
	if err != nil {
		major.Println(fmt.Sprintf("SetOrderUtxoModel from bid err:%s", err.Error()))
		return nil
	}
	return nil
}

func SaveNewDummy1200FromBid(net string, out Output, priKeyHex string, index int64, txId string) error {
	if out.Script == "" && out.Address == "" {
		return nil
	}
	startIndex := GetSaveStartIndex(net, model.UtxoTypeDummy1200, 0)
	netParams := GetNetParams(net)
	addr, err := btcutil.DecodeAddress(out.Address, netParams)
	if err != nil {
		return err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return err
	}
	pkScript := hex.EncodeToString(pkScriptByte)

	newDummy := &model.OrderUtxoModel{
		UtxoId:        fmt.Sprintf("%s_%d", txId, index),
		Net:           net,
		UtxoType:      model.UtxoTypeDummy1200,
		Amount:        out.Amount,
		Address:       out.Address,
		PrivateKeyHex: priKeyHex,
		TxId:          txId,
		Index:         index,
		PkScript:      pkScript,
		UsedState:     model.UsedNo,
		SortIndex:     startIndex + 1,
		Timestamp:     tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetOrderUtxoModel(newDummy)
	if err != nil {
		major.Println(fmt.Sprintf("SetOrderUtxoModel from bid err:%s", err.Error()))
		return nil
	}
	return nil
}

func SaveNewBidYUtxo10000FromBid(net string, out Output, priKeyHex string, index int64, txId string) error {
	if out.Script == "" && out.Address == "" {
		return nil
	}
	startIndex := GetSaveStartIndex(net, model.UtxoTypeBidY, 0)
	netParams := GetNetParams(net)
	addr, err := btcutil.DecodeAddress(out.Address, netParams)
	if err != nil {
		return err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return err
	}
	pkScript := hex.EncodeToString(pkScriptByte)

	newDummy := &model.OrderUtxoModel{
		UtxoId:        fmt.Sprintf("%s_%d", txId, index),
		Net:           net,
		UtxoType:      model.UtxoTypeBidY,
		Amount:        out.Amount,
		Address:       out.Address,
		PrivateKeyHex: priKeyHex,
		TxId:          txId,
		Index:         index,
		PkScript:      pkScript,
		UsedState:     model.UsedNo,
		SortIndex:     startIndex + 1,
		Timestamp:     tool.MakeTimestamp(),
	}
	_, err = mongo_service.SetOrderUtxoModel(newDummy)
	if err != nil {
		major.Println(fmt.Sprintf("SetOrderUtxoModel from bid err:%s", err.Error()))
		return nil
	}
	return nil
}

func CollectionUtxo(req *request.CollectionUtxo) (string, error) {
	var (
		netParams   *chaincfg.Params = GetNetParams(req.Net)
		err         error
		txRaw       string = ""
		totalIn     uint64 = 0
		totalAmount int64  = 0
	)
	inputs := make([]*TxInputUtxo, 0)
	for _, v := range req.UtxoList {
		inputs = append(inputs, &TxInputUtxo{
			TxId:     v.TxId,
			TxIndex:  v.Index,
			PkScript: v.PkScript,
			Amount:   v.Amount,
			PriHex:   req.PriKeyHex,
		})
		totalIn = totalIn + v.Amount
	}

	totalSize := int64(len(inputs))*SpendSize + 2*OutSize + OtherSize

	//totalAmount = int64(totalIn)- totalSize*req.FeeRate-546
	totalAmount = int64(totalIn) - totalSize*req.FeeRate - 140000

	fmt.Printf("totalSize:%d, totalIn:%d, totalSize*req.FeeRate:%d, totalAmount:%d\n", totalSize, totalIn, totalSize*req.FeeRate, totalAmount)

	outputs := make([]*TxOutput, 0)
	outputs = append(outputs, &TxOutput{
		Address: req.Address,
		Amount:  totalAmount,
	})

	tx, err := BuildCommonTx(netParams, inputs, outputs, req.Address, req.FeeRate)
	if err != nil {
		fmt.Printf("BuildCommonTx err:%s\n", err.Error())
		return "", err
	}
	txRaw, err = ToRaw(tx)
	if err != nil {
		return "", err
	}
	txId := ""
	//if req.Net == "testnet" {
	//	txResp, err := mempool_space_service.BroadcastTx(req.Net, txRaw)
	//	if err != nil {
	//		return "", err
	//	}
	//	txId = txResp
	//}else {
	//	txResp, err := oklink_service.BroadcastTx(txRaw)
	//	if err != nil {
	//		return "", err
	//	}
	//	txId = txResp.TxId
	//}

	txResp, err := unisat_service.BroadcastTx(req.Net, txRaw)
	if err != nil {
		return "", err
	}
	txId = txResp.Result

	//txResp, err := node.BroadcastTx(req.Net, txRaw)
	//if err != nil {
	//	return "", err
	//}
	//txId = txResp

	return txId, nil
}

// cold down the brc20 transfer
func ColdDownBrc20Transfer(req *request.ColdDownBrcTransfer) (*respond.Brc20TransferCommitResp, error) {
	var (
		netParams                                 *chaincfg.Params = GetNetParams(req.Net)
		_, platformAddressSendBrc20               string           = GetPlatformKeyAndAddressSendBrc20(req.Net)
		transferContent                           string           = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, req.Tick, req.InscribeTransferAmount)
		commitTxHash, revealTxHash, inscriptionId string           = "", "", ""
		err                                       error
		brc20BalanceResult                        *oklink_service.OklinkBrc20BalanceDetails
		availableBalance                          int64 = 0
	)

	fmt.Println(transferContent)
	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressSendBrc20, req.Tick, 1, 50)
	if err != nil {
		return nil, err
	}
	availableBalance, _ = strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	if availableBalance < req.InscribeTransferAmount {
		return nil, errors.New("AvailableBalance not enough. ")
	}
	commitTxHash, revealTxHash, inscriptionId, err = inscription_service.InscribeOneData(netParams, req.PriKeyHex, platformAddressSendBrc20, transferContent, req.FeeRate, req.ChangeAddress)
	if err != nil {
		return nil, err
	}
	return &respond.Brc20TransferCommitResp{
		CommitTxHash:  commitTxHash,
		RevealTxHash:  revealTxHash,
		InscriptionId: inscriptionId,
	}, nil
}

func ColdDownBrc20TransferBatch(req *request.ColdDownBrcTransferBatch) (*respond.Brc20TransferCommitBatchResp, error) {
	var (
		netParams                           *chaincfg.Params = GetNetParams(req.Net)
		_, platformAddressSendBrc20         string           = GetPlatformKeyAndAddressSendBrc20(req.Net)
		transferContent                     string           = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, req.Tick, req.InscribeTransferAmount)
		commitTxHash                        string           = ""
		revealTxHashList, inscriptionIdList []string         = make([]string, 0), make([]string, 0)
		err                                 error
		brc20BalanceResult                  *oklink_service.OklinkBrc20BalanceDetails
		availableBalance                    int64                               = 0
		fees                                int64                               = 0
		inscribeUtxoList                    []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
	)
	inscribeUtxoList = append(inscribeUtxoList, &inscription_service.InscribeUtxo{
		OutTx:     req.TxId,
		OutIndex:  req.Index,
		OutAmount: int64(req.Amount),
	})

	fmt.Println(transferContent)
	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressSendBrc20, req.Tick, 1, 50)
	if err != nil {
		return nil, err
	}
	availableBalance, _ = strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	fmt.Printf("availableBalance:%d, req.InscribeTransferAmount*req.Count: %d\n", availableBalance, req.InscribeTransferAmount*req.Count)
	if availableBalance < req.InscribeTransferAmount*req.Count {
		return nil, errors.New("AvailableBalance not enough. ")
	}
	commitTxHash, revealTxHashList, inscriptionIdList, fees, err =
		inscription_service.InscribeMultiDataFromUtxo(netParams, req.PriKeyHex, platformAddressSendBrc20,
			transferContent, req.FeeRate, req.ChangeAddress, req.Count, inscribeUtxoList, "", req.IsOnlyCal, 0)
	if err != nil {
		return nil, err
	}
	return &respond.Brc20TransferCommitBatchResp{
		Fees:              fees,
		CommitTxHash:      commitTxHash,
		RevealTxHashList:  revealTxHashList,
		InscriptionIdList: inscriptionIdList,
	}, nil
}

func ColdDownBatchBrc20TransferAndMakeAsk(req *request.ColdDownBrcTransferBatch) (*respond.Brc20TransferCommitBatchResp, error) {
	var (
		netParams                                                         *chaincfg.Params = GetNetParams(req.Net)
		platformPrivateKeySendBrc20ForAsk, platformAddressSendBrc20ForAsk string           = GetPlatformKeyAndAddressSendBrc20ForAsk(req.Net)
		_, platformAddressReceiveValueForAsk                              string           = GetPlatformKeyAndAddressReceiveValueForAsk(req.Net)
		transferContent                                                   string           = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, req.Tick, req.InscribeTransferAmount)
		commitTxHash                                                      string           = ""
		revealTxHashList, inscriptionIdList                               []string         = make([]string, 0), make([]string, 0)
		err                                                               error
		brc20BalanceResult                                                *oklink_service.OklinkBrc20BalanceDetails
		availableBalance                                                  int64                               = 0
		fees                                                              int64                               = 0
		inscribeUtxoList                                                  []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
		commonCoinAmount                                                  uint64                              = uint64(req.InscribeTransferAmount)
		commonOutAmount                                                   uint64                              = 4000
		commonCoinRatePrice                                               uint64                              = 0
	)
	inscribeUtxoList = append(inscribeUtxoList, &inscription_service.InscribeUtxo{
		OutTx:     req.TxId,
		OutIndex:  req.Index,
		OutAmount: int64(req.Amount),
	})

	fmt.Println(transferContent)
	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressSendBrc20ForAsk, req.Tick, 1, 50)
	if err != nil {
		return nil, err
	}
	availableBalance, _ = strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	fmt.Printf("availableBalance:%d, req.InscribeTransferAmount*req.Count: %d\n", availableBalance, req.InscribeTransferAmount*req.Count)
	if availableBalance < req.InscribeTransferAmount*req.Count {
		return nil, errors.New("AvailableBalance not enough. ")
	}
	commitTxHash, revealTxHashList, inscriptionIdList, fees, err =
		inscription_service.InscribeMultiDataFromUtxo(netParams, req.PriKeyHex, platformAddressSendBrc20ForAsk,
			transferContent, req.FeeRate, req.ChangeAddress, req.Count, inscribeUtxoList, req.OutAddressType, req.IsOnlyCal, 0)
	if err != nil {
		return nil, err
	}

	//make ask order
	outAmountDe := decimal.NewFromInt(int64(commonOutAmount))
	coinAmountDe := decimal.NewFromInt(int64(commonCoinAmount))
	coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
	commonCoinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)

	addr, err := btcutil.DecodeAddress(platformAddressSendBrc20ForAsk, netParams)
	if err != nil {
		return nil, err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}
	for _, v := range inscriptionIdList {
		inscriptionIdStrs := strings.Split(v, "i")
		inscriptionTxId := inscriptionIdStrs[0]
		inscriptionTxIndex, _ := strconv.ParseInt(inscriptionIdStrs[1], 10, 64)

		inputs := make([]Input, 0)
		inputs = append(inputs, Input{
			OutTxId:  inscriptionTxId,
			OutIndex: uint32(inscriptionTxIndex),
		})

		outputs := make([]Output, 0)
		outputs = append(outputs, Output{
			Address: platformAddressReceiveValueForAsk,
			Amount:  commonOutAmount,
		})
		inputSigns := make([]*InputSign, 0)

		inputSigns = append(inputSigns, &InputSign{
			Index:       0,
			OutRaw:      "",
			PkScript:    hex.EncodeToString(pkScript),
			SighashType: txscript.SigHashSingle | txscript.SigHashAnyOneCanPay,
			PriHex:      platformPrivateKeySendBrc20ForAsk,
			UtxoType:    Witness,
			Amount:      546,
		})

		builder, err := CreatePsbtBuilder(netParams, inputs, outputs)
		if err != nil {
			return nil, err
		}
		err = builder.UpdateAndSignInput(inputSigns)
		if err != nil {
			return nil, err
		}
		psbtRaw, err := builder.ToString()
		if err != nil {
			return nil, err
		}

		orderId := fmt.Sprintf("%s_%s_%s_%s_%d_%d", req.Net, req.Tick, v, platformAddressSendBrc20ForAsk, commonOutAmount, commonCoinAmount)
		orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
		entity := &model.OrderBrc20Model{
			Net:            req.Net,
			OrderId:        orderId,
			Tick:           req.Tick,
			Amount:         commonOutAmount,
			DecimalNum:     8,
			CoinAmount:     commonCoinAmount,
			CoinDecimalNum: 18,
			CoinRatePrice:  commonCoinRatePrice,
			//OrderState:     model.OrderStateCreate,
			InscriptionId: v,
			//OrderState:    model.OrderStatePreAsk,
			OrderState: model.OrderStatePreClaim,
			//OrderState:    model.OrderStatePoolPreClaim,
			OrderType:     model.OrderTypeSell,
			SellerAddress: platformAddressSendBrc20ForAsk,
			BuyerAddress:  "",
			PsbtRawPreAsk: psbtRaw,
			FreeState:     model.FreeStateClaim,
			//FreeState: model.FreeStatePoolClaim,
			Timestamp: tool.MakeTimestamp(),
		}
		_, err = mongo_service.SetOrderBrc20Model(entity)
		if err != nil {
			return nil, err
		}
		UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	}

	return &respond.Brc20TransferCommitBatchResp{
		Fees:              fees,
		CommitTxHash:      commitTxHash,
		RevealTxHashList:  revealTxHashList,
		InscriptionIdList: inscriptionIdList,
	}, nil
}

func ColdDownBatchBrc20TransferAndMakePool(req *request.ColdDownBrcTransferBatch) (*respond.Brc20TransferCommitBatchResp, error) {
	var (
		netParams                                                 *chaincfg.Params = GetNetParams(req.Net)
		platformPrivateKeyRewardBrc20, platformAddressRewardBrc20 string           = GetPlatformKeyAndAddressForRewardBrc20(req.Net)
		_, platformAddressReceiveValueForAsk                      string           = GetPlatformKeyAndAddressReceiveValueForAsk(req.Net)
		transferContent                                           string           = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, req.Tick, req.InscribeTransferAmount)
		commitTxHash                                              string           = ""
		revealTxHashList, inscriptionIdList                       []string         = make([]string, 0), make([]string, 0)
		err                                                       error
		brc20BalanceResult                                        *oklink_service.OklinkBrc20BalanceDetails
		availableBalance                                          int64                               = 0
		fees                                                      int64                               = 0
		inscribeUtxoList                                          []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
		commonCoinAmount                                          uint64                              = uint64(req.InscribeTransferAmount)
		commonOutAmount                                           uint64                              = 4000
		commonCoinRatePrice                                       uint64                              = 0
	)
	inscribeUtxoList = append(inscribeUtxoList, &inscription_service.InscribeUtxo{
		OutTx:     req.TxId,
		OutIndex:  req.Index,
		OutAmount: int64(req.Amount),
	})

	fmt.Println(transferContent)
	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(platformAddressRewardBrc20, req.Tick, 1, 50)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%+v\n", brc20BalanceResult)
	availableBalance, _ = strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	fmt.Printf("availableBalance:%d, req.InscribeTransferAmount*req.Count: %d\n", availableBalance, req.InscribeTransferAmount*req.Count)
	if availableBalance < req.InscribeTransferAmount*req.Count {
		return nil, errors.New("AvailableBalance not enough. ")
	}
	commitTxHash, revealTxHashList, inscriptionIdList, fees, err =
		inscription_service.InscribeMultiDataFromUtxo(netParams, req.PriKeyHex, platformAddressRewardBrc20,
			transferContent, req.FeeRate, req.ChangeAddress, req.Count, inscribeUtxoList, req.OutAddressType, req.IsOnlyCal, 0)
	if err != nil {
		return nil, err
	}

	//make ask order
	outAmountDe := decimal.NewFromInt(int64(commonOutAmount))
	coinAmountDe := decimal.NewFromInt(int64(commonCoinAmount))
	coinRatePriceStr := outAmountDe.Div(coinAmountDe).StringFixed(0)
	commonCoinRatePrice, _ = strconv.ParseUint(coinRatePriceStr, 10, 64)

	addr, err := btcutil.DecodeAddress(platformAddressRewardBrc20, netParams)
	if err != nil {
		return nil, err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}
	for _, v := range inscriptionIdList {
		inscriptionIdStrs := strings.Split(v, "i")
		inscriptionTxId := inscriptionIdStrs[0]
		inscriptionTxIndex, _ := strconv.ParseInt(inscriptionIdStrs[1], 10, 64)

		inputs := make([]Input, 0)
		inputs = append(inputs, Input{
			OutTxId:  inscriptionTxId,
			OutIndex: uint32(inscriptionTxIndex),
		})

		outputs := make([]Output, 0)
		outputs = append(outputs, Output{
			Address: platformAddressReceiveValueForAsk,
			Amount:  commonOutAmount,
		})
		inputSigns := make([]*InputSign, 0)

		inputSigns = append(inputSigns, &InputSign{
			Index:       0,
			OutRaw:      "",
			PkScript:    hex.EncodeToString(pkScript),
			SighashType: txscript.SigHashSingle | txscript.SigHashAnyOneCanPay,
			PriHex:      platformPrivateKeyRewardBrc20,
			UtxoType:    Witness,
			Amount:      546,
		})

		builder, err := CreatePsbtBuilder(netParams, inputs, outputs)
		if err != nil {
			return nil, err
		}
		err = builder.UpdateAndSignInput(inputSigns)
		if err != nil {
			return nil, err
		}
		psbtRaw, err := builder.ToString()
		if err != nil {
			return nil, err
		}

		orderId := fmt.Sprintf("%s_%s_%s_%s_%d_%d", req.Net, req.Tick, v, platformAddressRewardBrc20, commonOutAmount, commonCoinAmount)
		orderId = hex.EncodeToString(tool.SHA256([]byte(orderId)))
		entity := &model.OrderBrc20Model{
			Net:            req.Net,
			OrderId:        orderId,
			Tick:           req.Tick,
			Amount:         commonOutAmount,
			DecimalNum:     8,
			CoinAmount:     commonCoinAmount,
			CoinDecimalNum: 18,
			CoinRatePrice:  commonCoinRatePrice,
			//OrderState:     model.OrderStateCreate,
			InscriptionId: v,
			//OrderState:    model.OrderStatePreAsk,
			//OrderState:    model.OrderStatePreClaim,
			OrderState:    model.OrderStatePoolPreClaim,
			OrderType:     model.OrderTypeSell,
			SellerAddress: platformAddressRewardBrc20,
			BuyerAddress:  "",
			PsbtRawPreAsk: psbtRaw,
			//FreeState:     model.FreeStateClaim,
			FreeState: model.FreeStatePoolClaim,
			Timestamp: tool.MakeTimestamp(),
		}
		_, err = mongo_service.SetOrderBrc20Model(entity)
		if err != nil {
			return nil, err
		}
		//UpdateMarketPrice(req.Net, req.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(req.Tick)))
	}

	return &respond.Brc20TransferCommitBatchResp{
		Fees:              fees,
		CommitTxHash:      commitTxHash,
		RevealTxHashList:  revealTxHashList,
		InscriptionIdList: inscriptionIdList,
	}, nil
}
