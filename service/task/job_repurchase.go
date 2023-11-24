package task

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
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/create_key"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
	"time"
)

func jobForRepurchase() {
	var (
		net                                                     string                     = "livenet"
		tick                                                    string                     = "orxc"
		platformPrivateKeyReceiveFee, platformAddressReceiveFee string                     = order_brc20_service.GetPlatformKeyAndAddressReceiveFee(net)
		utxoList                                                []*oklink_service.UtxoItem = make([]*oklink_service.UtxoItem, 0)
		totalAmount                                             int64                      = 0
		entityList                                              []*model.OrderBrc20Model
		utxoEntityList                                          []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		netParams                                               *chaincfg.Params        = order_brc20_service.GetNetParams(net)
	)
	_ = platformAddressReceiveFee

	//return

	for i := int64(0); i < 1; i++ {
		utxoResp, err := oklink_service.GetAddressUtxo(platformAddressReceiveFee, i+1, 50)
		if err != nil {
			fmt.Printf("[JOB-Repurchase]-%s\n", fmt.Sprintf("Recheck address utxo list err:%s", err.Error()))
			return
		}
		if utxoResp.UtxoList != nil && len(utxoResp.UtxoList) != 0 {
			utxoList = append(utxoList, utxoResp.UtxoList...)
		} else {
			break
		}
		time.Sleep(10 * time.Second)
	}
	for _, u := range utxoList {
		amountDe, _ := decimal.NewFromString(u.UnspentAmount)
		amount := amountDe.Mul(decimal.New(1, 8)).IntPart()
		totalAmount = totalAmount + amount

		addr, err := btcutil.DecodeAddress(platformAddressReceiveFee, netParams)
		if err != nil {
			return
		}
		pkScriptBtc, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return
		}
		index, _ := strconv.ParseInt(u.Index, 10, 64)
		utxoEntityList = append(utxoEntityList, &model.OrderUtxoModel{
			TxId:          u.TxId,
			Index:         index,
			Amount:        uint64(amount),
			PrivateKeyHex: platformPrivateKeyReceiveFee,
			PkScript:      hex.EncodeToString(pkScriptBtc),
		})
	}
	major.Println(fmt.Sprintf("[JOB-Repurchase] check utxo len[%d], totalFees[%d]", len(utxoList), totalAmount))

	entityList, _ = mongo_service.FindOrderBrc20ModelList(net, tick, "", "",
		model.OrderTypeSell, model.OrderStateCreate,
		100, 0, 0, "coinRatePrice", 1, 0, 0)

	if entityList == nil || len(entityList) == 0 {
		return
	}
	for _, v := range entityList {

		remainingFeesUtxoList, err := repurchaseAsk(v, utxoEntityList)
		if err != nil {
			major.Println(fmt.Sprintf("[JOB-Repurchase] orderId:[%s][%d][%d][%d] err:%s", v.OrderId, v.CoinRatePrice, v.CoinPrice, v.CoinAmount, err.Error()))
			return
		}
		major.Println(fmt.Sprintf("[JOB-Repurchase] orderId:[%s][%d][%d][%d] success", v.OrderId, v.CoinRatePrice, v.CoinPrice, v.CoinAmount))
		utxoEntityList = remainingFeesUtxoList
		break
	}
}

func repurchaseAsk(askOrder *model.OrderBrc20Model, allFeesUtxoList []*model.OrderUtxoModel) ([]*model.OrderUtxoModel, error) {
	var (
		psbtBuilder            *order_brc20_service.PsbtBuilder
		netParams              *chaincfg.Params = order_brc20_service.GetNetParams(askOrder.Net)
		utxoDummyList          []*model.OrderUtxoModel
		needFeesUtxoList       []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		remainingFeesUtxoList  []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		err                    error
		sellerSendAddress      string = askOrder.SellerAddress
		inValue                uint64 = 0
		coinAmount             uint64 = 0
		brc20ReceiveValue      uint64 = 0
		inscriptionOutputValue uint64 = 0

		inscriptionId               string = ""
		inscriptionBrc20BalanceItem *oklink_service.BalanceItem
		newPsbtBuilder              *order_brc20_service.PsbtBuilder

		_, platformAddressReceiveFee        string = order_brc20_service.GetPlatformKeyAndAddressReceiveFee(askOrder.Net)
		_, platformAddressReceiveDummyValue string = order_brc20_service.GetPlatformKeyAndAddressReceiveDummyValue(askOrder.Net)
		_, platformAddressRepurchaseAsk     string = order_brc20_service.GetPlatformKeyAndAddressForRepurchaseReceiveBrc20(askOrder.Net)
	)

	newDummyOutPriKeyHex, newDummyOutSegwitAddress, err := create_key.CreateSegwitKey(netParams)
	if err != nil {
		return nil, err
	}

	psbtBuilder, err = order_brc20_service.NewPsbtBuilder(netParams, askOrder.PsbtRawPreAsk)
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
	if strings.ToLower(askOrder.Net) != "testnet" {
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
	if sellerReceiveValue != askOrder.Amount {
		return nil, errors.New("Wrong Psbt: Seller receive value dose not match. ")
	}

	has := false
	for _, v := range preOutList {
		inscriptionId = fmt.Sprintf("%s:%d", v.PreviousOutPoint.Hash.String(), v.PreviousOutPoint.Index)
		inscriptionBrc20BalanceItem, err = order_brc20_service.CheckBrc20Ordinals(v, askOrder.Tick, sellerSendAddress)
		if err != nil {
			continue
		}
		has = true
	}
	_ = inscriptionId

	if askOrder.Net == "mainnet" || askOrder.Net == "livenet" {
		if !has || inscriptionBrc20BalanceItem == nil {
			return nil, errors.New("Wrong Psbt: Empty inscription. ")
		}
		coinAmount, _ = strconv.ParseUint(inscriptionBrc20BalanceItem.Amount, 10, 64)
	}
	if coinAmount != askOrder.CoinAmount {
		return nil, errors.New("Wrong Psbt: brc20 coin amount dose not match. ")
	}

	brc20ReceiveValue = inValue

	utxoDummyList, err = order_brc20_service.GetUnoccupiedUtxoList(askOrder.Net, 2, 0, model.UtxoTypeDummy, "", 0)
	defer order_brc20_service.ReleaseUtxoList(utxoDummyList)
	if err != nil {
		return nil, err
	}

	//get pay utxo
	if askOrder.Fee == 0 {
		askOrder.Fee = 7000
	}
	totalNeedAmount := sellerReceiveValue + askOrder.Fee + inscriptionOutputValue

	needFeesAmount := int64(0)
	for _, v := range allFeesUtxoList {
		if needFeesAmount >= int64(totalNeedAmount) {
			remainingFeesUtxoList = append(remainingFeesUtxoList, v)
		} else {
			needFeesAmount = needFeesAmount + int64(v.Amount)
			needFeesUtxoList = append(needFeesUtxoList, v)
		}
	}
	if needFeesAmount < int64(totalNeedAmount) {
		return nil, errors.New("utxos not enough")
	}

	changeAmount := needFeesAmount - int64(totalNeedAmount)

	fmt.Printf("changeAmount: %d\n", changeAmount)
	if err != nil {
		return nil, err
	}

	inputs := make([]order_brc20_service.Input, 0)
	outputs := make([]order_brc20_service.Output, 0)
	dummyOutValue := uint64(0)
	//add dummy ins - index: 0,1
	for _, dummy := range utxoDummyList {
		inputs = append(inputs, order_brc20_service.Input{
			OutTxId:  dummy.TxId,
			OutIndex: uint32(dummy.Index),
		})
		dummyOutValue = dummyOutValue + dummy.Amount
	}
	//add seller brc20 ins - index: 2
	inputs = append(inputs, order_brc20_service.Input{
		OutTxId:  preOutList[0].PreviousOutPoint.Hash.String(),
		OutIndex: preOutList[0].PreviousOutPoint.Index,
	})
	//add Exchange pay value ins - index: 3,3+
	for _, payBid := range needFeesUtxoList {
		inputs = append(inputs, order_brc20_service.Input{
			OutTxId:  payBid.TxId,
			OutIndex: uint32(payBid.Index),
		})
	}

	//add dummy outs - idnex: 0
	outputs = append(outputs, order_brc20_service.Output{
		Address: platformAddressReceiveDummyValue,
		Amount:  dummyOutValue,
	})
	//add receive brc20 outs - idnex: 1
	receiveBrc20 := order_brc20_service.Output{
		Address: platformAddressRepurchaseAsk,
		Amount:  brc20ReceiveValue + inscriptionOutputValue,
	}
	outputs = append(outputs, receiveBrc20)
	//add receive seller outs - idnex: 2
	outputs = append(outputs, order_brc20_service.Output{
		Address: sellerReceiveAddress,
		Amount:  sellerReceiveValue,
	})
	//add new dummy outs - idnex: 3,4
	newDummyOut := order_brc20_service.Output{
		Address: newDummyOutSegwitAddress,
		Amount:  600,
	}
	outputs = append(outputs, newDummyOut)
	outputs = append(outputs, newDummyOut)

	if changeAmount >= 546 {
		outputs = append(outputs, order_brc20_service.Output{
			Address: platformAddressReceiveFee,
			Amount:  uint64(changeAmount),
		})
	}

	//finish PSBT(Y)
	newPsbtBuilder, err = order_brc20_service.CreatePsbtBuilder(netParams, inputs, outputs)
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

	inSigns := make([]*order_brc20_service.InputSign, 0)
	//add dummy ins sign - index: 0,1
	for k, dummy := range utxoDummyList {
		inSigns = append(inSigns, &order_brc20_service.InputSign{
			Index:       k,
			PkScript:    dummy.PkScript,
			Amount:      dummy.Amount,
			SighashType: txscript.SigHashAll,
			PriHex:      dummy.PrivateKeyHex,
			UtxoType:    order_brc20_service.Witness,
		})
	}
	//add fees pay value ins - index: 3,3+
	for k, payBid := range needFeesUtxoList {
		inSigns = append(inSigns, &order_brc20_service.InputSign{
			Index:       k + 3,
			PkScript:    payBid.PkScript,
			Amount:      payBid.Amount,
			SighashType: txscript.SigHashAll,
			PriHex:      payBid.PrivateKeyHex,
			UtxoType:    order_brc20_service.Witness,
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
	askOrder.PsbtRawFinalAsk = psbtRawFinalAsk

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

	txPsbtYResp, err := unisat_service.BroadcastTx(askOrder.Net, txRawPsbtY)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Broadcast Psbt(Y) %s err:%s", askOrder.Net, err.Error()))
	}

	askOrder.PsbtAskTxId = psbtYTxId
	askOrder.BuyerAddress = platformAddressRepurchaseAsk
	askOrder.PsbtRawFinalAsk = psbtRawFinalAsk
	askOrder.DealTime = tool.MakeTimestamp()
	askOrder.OrderState = model.OrderStateFinish
	order_brc20_service.SetUsedDummyUtxo(utxoDummyList, txPsbtYResp.Result)
	order_brc20_service.SaveNewDummyFromBid(askOrder.Net, newDummyOut, newDummyOutPriKeyHex, 3, psbtYTxId)
	order_brc20_service.SaveNewDummyFromBid(askOrder.Net, newDummyOut, newDummyOutPriKeyHex, 4, psbtYTxId)
	_, err = mongo_service.SetOrderBrc20Model(askOrder)
	if err != nil {
		return nil, err
	}

	return remainingFeesUtxoList, nil
}
