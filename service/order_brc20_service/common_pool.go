package order_brc20_service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/shopspring/decimal"
	"ordbook-aggregation/config"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/inscription_service"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/unisat_service"
	"ordbook-aggregation/tool"
	"strconv"
	"strings"
)

type addressToKey struct {
	key        *btcec.PrivateKey
	compressed bool
}

func mkGetKey(keys map[string]addressToKey) txscript.KeyDB {
	if keys == nil {
		return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
			bool, error) {
			return nil, false, errors.New("nope")
		})
	}
	return txscript.KeyClosure(func(addr btcutil.Address) (*btcec.PrivateKey,
		bool, error) {
		a2k, ok := keys[addr.EncodeAddress()]
		if !ok {
			return nil, false, errors.New("nope")
		}
		return a2k.key, a2k.compressed, nil
	})
}

func mkGetScript(scripts map[string][]byte) txscript.ScriptDB {
	if scripts == nil {
		return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
			return nil, errors.New("nope")
		})
	}
	return txscript.ScriptClosure(func(addr btcutil.Address) ([]byte, error) {
		script, ok := scripts[addr.EncodeAddress()]
		if !ok {
			return nil, errors.New("nope")
		}
		return script, nil
	})
}

func createMultiSigAddress(net *chaincfg.Params, pubKey ...string) (string, string, string, error) {
	var (
		pubKeys = make([]*btcutil.AddressPubKey, 0)
	)
	for _, v := range pubKey {
		pubByte, err := hex.DecodeString(v)
		if err != nil {
			return "", "", "", err
		}
		pub, err := btcutil.NewAddressPubKey(pubByte, net)
		if err != nil {
			return "", "", "", err
		}
		pubKeys = append(pubKeys, pub)
	}

	requiredSigs := len(pubKey)
	multiSigScript, err := txscript.MultiSigScript(pubKeys, requiredSigs)
	if err != nil {
		fmt.Println("Failed to create multi-sig script:", err)
		return "", "", "", err
	}
	address, err := btcutil.NewAddressScriptHash(multiSigScript, net)
	if err != nil {
		fmt.Println("Failed to create  native address:", err)
		return "", "", "", err

	}

	h := sha256.New()
	h.Write(multiSigScript)

	nativeSegwitAddress, err := btcutil.NewAddressWitnessScriptHash(h.Sum(nil), net)
	if err != nil {
		fmt.Println("Failed to create native SegWit address:", err)
		return "", "", "", err
	}

	fmt.Println("Multi-Sig Address:", address)
	return hex.EncodeToString(multiSigScript), address.EncodeAddress(), nativeSegwitAddress.EncodeAddress(), nil
}

func signMultiSigScript(net *chaincfg.Params, tx *wire.MsgTx, i int, pkScript []byte, hashType txscript.SigHashType, priKey string, preSigScript []byte) ([]byte, error) {
	privateKeyBytes, err := hex.DecodeString(priKey)
	if err != nil {
		return nil, err
	}
	privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

	publicKey := privateKey.PubKey().SerializeCompressed()
	address, err := btcutil.NewAddressPubKey(publicKey, net)
	if err != nil {
		return nil, err
	}

	scriptAddr, err := btcutil.NewAddressScriptHash(pkScript, net)
	if err != nil {
		return nil, err
	}

	// Sign with the other key and merge
	sigScript, err := txscript.SignTxOutput(net,
		tx, i, pkScript, hashType,
		mkGetKey(map[string]addressToKey{
			address.EncodeAddress(): {privateKey, true},
		}), mkGetScript(map[string][]byte{
			scriptAddr.EncodeAddress(): pkScript,
		}), preSigScript)
	return sigScript, err
}

func getPoolBrc20PsbtOrder(net, tick string, limit, page, flag int64) ([]*model.PoolBrc20Model, int64, error) {
	var (
		entityList []*model.PoolBrc20Model
		total      int64 = 0
	)
	total, _ = mongo_service.CountPoolBrc20ModelList(net, tick, "", "", model.PoolTypeAll, model.PoolStateAdd)
	entityList, _ = mongo_service.FindPoolBrc20ModelList(net, tick, "", "", model.PoolTypeAll, model.PoolStateAdd,
		limit, flag, page, "coinRatePrice", 1)
	return entityList, total, nil
}

func getOnePoolBrc20OrderByOrderId(orderId string) (*model.PoolBrc20Model, error) {
	var (
		entity *model.PoolBrc20Model
	)
	entity, _ = mongo_service.FindPoolBrc20ModelByOrderId(orderId)
	if entity == nil || entity.Id == 0 {
		return nil, errors.New("order is empty")
	}
	return entity, nil
}

func setStatusPoolBrc20Order(bidOrder *model.OrderBrc20Model, poolState model.PoolState, dealTxIndex, dealTxOutValue, dealTime int64) {
	err := mongo_service.SetPoolBrc20ModelForStatus(bidOrder.PoolOrderId, poolState, bidOrder.PsbtBidTxId, dealTxIndex, dealTxOutValue, dealTime)
	if err != nil {
		major.Println(fmt.Sprintf("SetPoolBrc20ModelForStatus err:%s", err))
	}
}

func setCoinStatusPoolBrc20Order(bidOrder *model.OrderBrc20Model, poolCoinState model.PoolState, dealCoinTxIndex, dealCoinTxOutValue, dealCoinTime int64) {
	err := mongo_service.SetPoolBrc20ModelForCoinStatus(bidOrder.PoolOrderId, poolCoinState, bidOrder.PsbtAskTxId, dealCoinTxIndex, dealCoinTxOutValue, dealCoinTime)
	if err != nil {
		major.Println(fmt.Sprintf("SetPoolBrc20ModelForStatus err:%s", err))
	}
	setRewardForPoolBrc20Order(bidOrder.PoolOrderId)
}

func setRewardForPoolBrc20Order(poolOrderId string) {
	var (
		entityPool   *model.PoolBrc20Model
		rewardAmount int64 = 0
		dayTime      int64 = 1000 * 60 * 60 * 24
	)
	entityPool, _ = mongo_service.FindPoolBrc20ModelByOrderId(poolOrderId)
	if entityPool == nil {
		return
	}

	if entityPool.PoolType == model.PoolTypeTick {
		rewardAmount = getSinglePoolReward()
	} else {
		rewardAmount = getDoublePoolReward(entityPool.Ratio)
		dis := entityPool.DealTime - entityPool.Timestamp
		days := dis / dayTime
		count := days / config.PlatformRewardExtraRewardDuration

		entity, _ := mongo_service.FindPoolInfoModelByPair(entityPool.Net, strings.ToUpper(entityPool.Pair))
		if entity == nil || entity.Id == 0 {
			return
		}
		_, amountTotal, _, _ := getOwnPoolInfo(entityPool.Net, entityPool.Tick, strings.ToUpper(entityPool.Pair), entityPool.CoinAddress)
		ownerAmountTotalDe := decimal.NewFromInt(int64(amountTotal))
		amountTotalDe := decimal.NewFromInt(int64(entity.Amount))
		rewardAmountDe := decimal.NewFromInt(rewardAmount)
		extraReward := ownerAmountTotalDe.Div(amountTotalDe).Mul(rewardAmountDe).IntPart()
		extraReward = extraReward * count
		rewardAmount = rewardAmount + extraReward
	}
	entityPool.RewardAmount = rewardAmount
	err := mongo_service.SetPoolBrc20ModelForReward(entityPool)
	if err != nil {
		major.Println(fmt.Sprintf("SetPoolBrc20ModelForReward err: %s", err.Error()))
	}
	updatePoolInfo(entityPool)
}

func inscriptionMultiSigTransfer(poolOrder *model.PoolBrc20Model) error {
	var (
		//poolOrder                   *model.PoolBrc20Model
		utxoMultiSigInscriptionList []*model.OrderUtxoModel
		inscriptionAddress          string = ""
		brc20BalanceResult          *oklink_service.OklinkBrc20BalanceDetails
		availableBalance            int64 = 0

		commitTxHash                                                              string = ""
		revealTxHashList                                                          []string
		inscriptionIdList                                                         []string
		err                                                                       error
		transferContent                                                           string                              = ""
		feeRate                                                                   int64                               = 10
		inscribeUtxoList                                                          []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
		_, platformAddressReceiveBidValue                                         string                              = GetPlatformKeyAndAddressReceiveBidValue(poolOrder.Net)
		platformPrivateKeyMultiSigInscription, platformAddressMultiSigInscription string                              = GetPlatformKeyAndAddressForMultiSigInscription(poolOrder.Net)
		netParams                                                                 *chaincfg.Params                    = GetNetParams(poolOrder.Net)
		changeAddress                                                             string                              = platformAddressReceiveBidValue

		dealInscriptionId, dealInscriptionTx                                   string = "", ""
		dealInscriptionTxIndex, dealInscriptionTxOutValue, dealInscriptionTime int64  = 0, 1000, 0
	)
	//poolOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(poolOrderId)
	//if poolOrder == nil {
	//	return errors.New("[POOL-INSCRIPTION]poolOrder is empty")
	//}
	if poolOrder.DealInscriptionId != "" {
		return errors.New("[POOL-INSCRIPTION]poolOrder DealInscription has been done")
	}
	if poolOrder.PoolCoinState != model.PoolStateUsed {
		return errors.New("[POOL-INSCRIPTION]poolOrder not used")
	}

	if poolOrder.MultiSigScriptAddress == "" {
		MultiSigScriptByte, err := hex.DecodeString(poolOrder.MultiSigScript)
		if err != nil {
			return errors.New(fmt.Sprintf("[POOL-INSCRIPTION] err:%s", err.Error()))
		}
		h := sha256.New()
		h.Write(MultiSigScriptByte)
		nativeSegwitAddress, err := btcutil.NewAddressWitnessScriptHash(h.Sum(nil), netParams)
		if err != nil {
			fmt.Println("Failed to create native SegWit address:", err)
			return errors.New(fmt.Sprintf("[POOL-INSCRIPTION] err:%s", err.Error()))
		}
		poolOrder.MultiSigScriptAddress = nativeSegwitAddress.EncodeAddress()
	}
	inscriptionAddress = poolOrder.MultiSigScriptAddress
	if inscriptionAddress == "" {
		return errors.New(fmt.Sprintf("[POOL-INSCRIPTION] inscriptionAddress is empty"))
	}

	brc20BalanceResult, err = oklink_service.GetAddressBrc20BalanceResult(inscriptionAddress, poolOrder.Tick, 1, 50)
	if err != nil {
		return err
	}
	availableBalance, _ = strconv.ParseInt(brc20BalanceResult.AvailableBalance, 10, 64)
	fmt.Printf("availableBalance:%d, coinAmount: %d\n", availableBalance, poolOrder.CoinAmount)
	if availableBalance < int64(poolOrder.CoinAmount) {
		return errors.New("[POOL-INSCRIPTION] AvailableBalance not enough. ")
	}

	has := false
	brc20TxResp, _ := oklink_service.GetAddressBrc20BalanceTransactionList(inscriptionAddress, poolOrder.Tick, 0, 100)
	if brc20TxResp != nil && brc20TxResp.InscriptionsList != nil {
		for _, tx := range brc20TxResp.InscriptionsList {
			if tx.TxId == poolOrder.DealCoinTx && tx.State == "success" {
				has = true
				break
			}
		}
	} else {
		return errors.New("[POOL-INSCRIPTION] get brc20 tx list err. ")
	}
	if !has {
		return errors.New("[POOL-INSCRIPTION] receive brc20 not confirm. ")
	}

	transferContent = fmt.Sprintf(`{"p":"brc-20", "op":"transfer", "tick":"%s", "amt":"%d"}`, poolOrder.Tick, poolOrder.CoinAmount)

	utxoMultiSigInscriptionList, err = GetUnoccupiedUtxoList(poolOrder.Net, 1, 0, model.UtxoTypeMultiInscription)
	defer ReleaseUtxoList(utxoMultiSigInscriptionList)
	if err != nil {
		return errors.New(fmt.Sprintf("[POOL-INSCRIPTION] get utxo err:%s", err.Error()))
	}
	for _, v := range utxoMultiSigInscriptionList {
		if v.Address != platformAddressMultiSigInscription {
			continue
		}
		fmt.Printf("%+v\n", *v)
		inscribeUtxoList = append(inscribeUtxoList, &inscription_service.InscribeUtxo{
			OutTx:     v.TxId,
			OutIndex:  v.Index,
			OutAmount: int64(v.Amount),
		})
	}
	if len(inscribeUtxoList) <= 0 {
		return errors.New(fmt.Sprintf("[POOL-INSCRIPTION] get utxo empty"))
	}

	commitTxHash, revealTxHashList, inscriptionIdList, _, err =
		inscription_service.InscribeMultiDataFromUtxo(netParams, platformPrivateKeyMultiSigInscription, inscriptionAddress,
			transferContent, feeRate, changeAddress, 1, inscribeUtxoList, "segwit", false, dealInscriptionTxOutValue)
	if err != nil {
		return errors.New(fmt.Sprintf("[POOL-INSCRIPTION] Inscribe err:%s", err.Error()))
	}

	dealInscriptionId = inscriptionIdList[0]
	dealInscriptionTx = revealTxHashList[0]
	dealInscriptionTxIndex = 0
	dealInscriptionTime = tool.MakeTimestamp()

	poolOrder.DealInscriptionId = dealInscriptionId
	poolOrder.DealInscriptionTx = dealInscriptionTx
	poolOrder.DealInscriptionTxIndex = dealInscriptionTxIndex
	poolOrder.DealInscriptionTime = dealInscriptionTime

	err = mongo_service.SetPoolBrc20ModelForDealInscription(poolOrder.OrderId, dealInscriptionId, dealInscriptionTx, dealInscriptionTxIndex, dealInscriptionTxOutValue, dealInscriptionTime)
	if err != nil {
		major.Println(fmt.Sprintf("SetPoolBrc20ModelForDealInscription err:%s", err))
		return errors.New(fmt.Sprintf("[POOL-INSCRIPTION] save data err:%s", err.Error()))
	}
	setUsedMultiSigInscriptionUtxo(utxoMultiSigInscriptionList, commitTxHash)
	major.Println(fmt.Sprintf("[POOL] inscription for multiSigAddress success"))
	return nil
}

func claimPoolBrc20Order(orderId, claimAddress string, poolType model.PoolType, preSigScript []byte) (*wire.MsgTx, string, error) {
	var (
		entity                     *model.PoolBrc20Model
		claimTxId                  string = ""
		claimTxIndex               int64  = 0
		claimTxValue               int64  = 0
		claimMultiSigScript        string = ""
		claimMultiSigScriptByte    []byte
		netParams                  *chaincfg.Params
		fee                        int64  = 14
		platformPrivateKeyMultiSig string = ""
		platformPublicKeyMultiSig  string = ""
		platformAddressMultiSig    string = ""
		changeAddress              string = ""
	)
	_ = claimMultiSigScriptByte
	_ = fee
	entity, _ = mongo_service.FindPoolBrc20ModelByOrderId(orderId)
	if entity == nil || entity.Id == 0 {
		return nil, "", errors.New("order is empty")
	}
	netParams = GetNetParams(entity.Net)
	platformPrivateKeyMultiSig, platformPublicKeyMultiSig = GetPlatformKeyMultiSig(entity.Net)
	//if poolType == model.PoolTypeBtc {
	//	platformPrivateKeyMultiSig, platformPublicKeyMultiSig = GetPlatformKeyMultiSigForBtc(entity.Net)
	//}
	platformPublicKeyMultiSigByte, err := hex.DecodeString(platformPublicKeyMultiSig)
	if err != nil {
		return nil, "", err
	}
	nativeSegwitAddress, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(platformPublicKeyMultiSigByte), netParams)
	if err != nil {
		return nil, "", err
	}
	platformAddressMultiSig = nativeSegwitAddress.EncodeAddress()
	changeAddress = platformAddressMultiSig
	_ = changeAddress

	claimTxId = entity.DealCoinTx
	claimTxIndex = entity.DealCoinTxIndex
	claimTxValue = entity.DealCoinTxOutValue
	claimMultiSigScript = entity.MultiSigScript

	if poolType == model.PoolTypeBtc {
		claimTxId = entity.DealTx
		claimTxIndex = entity.DealTxIndex
		claimTxValue = entity.DealTxOutValue
		claimMultiSigScript = entity.MultiSigScript
		//claimMultiSigScript = entity.MultiSigScriptBtc
	} else if poolType == model.PoolTypeMultiSigInscription {
		if entity.DealInscriptionTx == "" {
			//err := inscriptionMultiSigTransfer(entity.Net, entity.OrderId)
			err := inscriptionMultiSigTransfer(entity)
			if err != nil {
				fmt.Println(err)
				return nil, "", err
			}
		}
		claimTxId = entity.DealInscriptionTx
		claimTxIndex = entity.DealInscriptionTxIndex
		claimTxValue = entity.DealInscriptionTxOutValue
		claimMultiSigScript = entity.MultiSigScript

		//check brc20 valid
		//todo

	}

	if claimTxId == "" || claimMultiSigScript == "" {
		return nil, "", errors.New("txId or multiSigScript of pool order is empty")
	}

	claimMultiSigScriptByte, err = hex.DecodeString(claimMultiSigScript)
	if err != nil {
		return nil, "", err
	}

	tx := wire.NewMsgTx(2)
	totalAmount := int64(0)
	outAmount := int64(0)

	//add outs
	addr, err := btcutil.DecodeAddress(claimAddress, netParams)
	if err != nil {
		return nil, "", err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, "", err
	}
	tx.AddTxOut(wire.NewTxOut(int64(claimTxValue), pkScript))
	outAmount = outAmount + int64(claimTxValue)

	//add ins
	hash, err := chainhash.NewHashFromStr(claimTxId)
	if err != nil {
		return nil, "", err
	}
	prevOut := wire.NewOutPoint(hash, uint32(claimTxIndex))
	txIn := wire.NewTxIn(prevOut, nil, nil)
	tx.AddTxIn(txIn)
	totalAmount = totalAmount + int64(claimTxValue)

	inputs := make([]Input, 0)
	inputs = append(inputs, Input{
		OutTxId:  claimTxId,
		OutIndex: uint32(claimTxIndex),
	})

	outputs := make([]Output, 0)
	outputs = append(outputs, Output{
		Address: claimAddress,
		Amount:  uint64(claimTxValue),
	})

	builder, err := CreatePsbtBuilder(netParams, inputs, outputs)
	if err != nil {
		return nil, "", err
	}
	h := sha256.New()
	h.Write(claimMultiSigScriptByte)
	segwitAddress, err := btcutil.NewAddressWitnessScriptHash(h.Sum(nil), netParams)
	if err != nil {
		fmt.Println("Failed to create native SegWit address:", err)
		return nil, "", errors.New(fmt.Sprintf("Failed to create native SegWit address:%s", err.Error()))
	}
	segwitAddr, err := btcutil.DecodeAddress(segwitAddress.EncodeAddress(), netParams)
	if err != nil {
		return nil, "", err
	}
	segwitPkScript, err := txscript.PayToAddrScript(segwitAddr)
	if err != nil {
		return nil, "", err
	}
	fmt.Println(hex.EncodeToString(segwitPkScript))

	inSigns := make([]*InputSign, 0)
	inSigns = append(inSigns, &InputSign{
		Index: 0,
		//PkScript:       claimMultiSigScript,
		PkScript:       hex.EncodeToString(segwitPkScript),
		Amount:         uint64(claimTxValue),
		SighashType:    txscript.SigHashSingle | txscript.SigHashAnyOneCanPay,
		PriHex:         platformPrivateKeyMultiSig,
		UtxoType:       Witness,
		PreSigScript:   hex.EncodeToString(preSigScript),
		MultiSigScript: claimMultiSigScript,
	})

	err = builder.UpdateAndMultiSignInput(inSigns)
	if err != nil {
		return nil, "", err
	}

	psbtRaw, err := builder.ToString()
	if err != nil {
		return nil, "", err
	}

	return tx, psbtRaw, nil
}

func updatePoolInfo(poolOrder *model.PoolBrc20Model) {
	if poolOrder == nil {
		return
	}

	var (
		entityPoolInfo *model.PoolInfoModel
	)
	entityPoolInfo, _ = mongo_service.FindPoolInfoModelByPair(poolOrder.Net, poolOrder.Pair)
	if entityPoolInfo == nil {
		entityPoolInfo = &model.PoolInfoModel{
			Net:            poolOrder.Net,
			Tick:           poolOrder.Tick,
			Pair:           poolOrder.Pair,
			CoinDecimalNum: poolOrder.CoinDecimalNum,
			DecimalNum:     poolOrder.DecimalNum,
			Timestamp:      poolOrder.Timestamp,
		}
	}

	//if poolOrder.PoolCoinState == model.PoolStateAdd {
	//	entityPoolInfo.CoinAmount = entityPoolInfo.CoinAmount + poolOrder.CoinAmount
	//} else {
	//	entityPoolInfo.CoinAmount = entityPoolInfo.CoinAmount - poolOrder.CoinAmount
	//}

	if poolOrder.PoolState == model.PoolStateAdd {
		entityPoolInfo.CoinAmount = entityPoolInfo.CoinAmount + poolOrder.CoinAmount
		entityPoolInfo.Amount = entityPoolInfo.Amount + poolOrder.Amount
	} else {
		entityPoolInfo.Amount = entityPoolInfo.Amount - poolOrder.Amount
		entityPoolInfo.CoinAmount = entityPoolInfo.CoinAmount - poolOrder.CoinAmount
	}

	_, err := mongo_service.SetPoolInfoModel(entityPoolInfo)
	if err != nil {
		major.Println(fmt.Sprintf("SetPoolInfoModel err:%s", err.Error()))
		return
	}
	return
}

func getOwnPoolInfo(net, tick, pair, address string) (uint64, uint64, uint64, error) {
	var (
		coinAmountTotal uint64 = 0
		amountTotal     uint64 = 0
		count           uint64 = 0
		entity          *model.PoolOrderCount
		entityBtc       *model.PoolOrderCount
		entityBoth      *model.PoolOrderCount
	)
	entity, _ = mongo_service.CountOwnPoolPair(net, tick, pair, address, model.PoolTypeTick)
	if entity != nil {
		coinAmountTotal = coinAmountTotal + uint64(entity.CoinAmountTotal)
		count = count + uint64(entity.OrderCounts)
	}

	entityBtc, _ = mongo_service.CountOwnPoolPair(net, tick, pair, address, model.PoolTypeBtc)
	if entityBtc != nil {
		amountTotal = amountTotal + uint64(entityBtc.AmountTotal)
		count = count + uint64(entityBtc.OrderCounts)
	}

	entityBoth, _ = mongo_service.CountOwnPoolPair(net, tick, pair, address, model.PoolTypeBoth)
	if entityBoth != nil {
		coinAmountTotal = coinAmountTotal + uint64(entityBoth.CoinAmountTotal)
		amountTotal = amountTotal + uint64(entityBoth.AmountTotal)
		count = count + uint64(entityBoth.OrderCounts)
	}

	return coinAmountTotal, amountTotal, count, nil
}

func getMyPoolInscription(net, tick, address string) ([]*model.PoolBrc20Model, int64) {
	var (
		total      int64 = 0
		entityList []*model.PoolBrc20Model
	)
	//total, _ = mongo_service.CountPoolBrc20ModelList(net, tick, "", address, model.PoolTypeTick, model.PoolStateAdd)
	//entityList, _ = mongo_service.FindPoolBrc20ModelList(net, tick, "", address, model.PoolTypeTick, model.PoolStateAdd,
	//	1000, 0, 0, "", 0)
	total, _ = mongo_service.CountPoolBrc20ModelList(net, tick, "", address, model.PoolTypeAll, model.PoolStateAdd)
	entityList, _ = mongo_service.FindPoolBrc20ModelList(net, tick, "", address, model.PoolTypeAll, model.PoolStateAdd,
		1000, 0, 0, "", 0)
	return entityList, total
}

func updateClaim(poolOrder *model.PoolBrc20Model, rawTx string) error {
	txPsbtResp, err := unisat_service.BroadcastTx(poolOrder.Net, rawTx)
	if err != nil {
		return errors.New(fmt.Sprintf("Broadcast Psbt %s, poolOrderId-%s err:%s", poolOrder.Net, poolOrder.OrderId, err.Error()))
	}
	_ = txPsbtResp

	poolOrder.ClaimTx = txPsbtResp.Result
	poolOrder.ClaimTime = tool.MakeTimestamp()
	if poolOrder.PoolState == model.PoolStateUsed {
		poolOrder.PoolState = model.PoolStateClaim
	}
	if poolOrder.PoolCoinState == model.PoolStateUsed {
		poolOrder.PoolCoinState = model.PoolStateClaim
	}

	rewardNowAmount := getRealNowReward(poolOrder)
	poolOrder.RewardRealAmount = rewardNowAmount
	poolOrder.ClaimTxBlockState = model.ClaimTxBlockStateUnconfirmed
	err = mongo_service.SetPoolBrc20ModelForClaim(poolOrder)
	if err != nil {
		return err
	}
	return nil
}

func saveNewMultiSigInscriptionUtxo(net, txId string, txIndex int64, amount uint64) error {
	startIndex := GetSaveStartIndex(net, model.UtxoTypeMultiInscriptionFromRelease, 0)
	_, fromSegwitAddress := GetPlatformKeyAndAddressForMultiSigInscription(net)
	addr, err := btcutil.DecodeAddress(fromSegwitAddress, GetNetParams(net))
	if err != nil {
		return err
	}
	pkScriptByte, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return err
	}
	pkScript := hex.EncodeToString(pkScriptByte)
	newUtxo := &model.OrderUtxoModel{
		//UtxoId:     "",
		Net:           net,
		UtxoType:      model.UtxoTypeMultiInscriptionFromRelease,
		Amount:        amount,
		Address:       fromSegwitAddress,
		PrivateKeyHex: "",
		TxId:          "",
		Index:         txIndex,
		PkScript:      pkScript,
		UsedState:     model.UsedNo,
		//UseTx:      "",
		SortIndex: startIndex + 1,
		Timestamp: tool.MakeTimestamp(),
	}
	newUtxo.TxId = txId
	newUtxo.UtxoId = fmt.Sprintf("%s_%d", newUtxo.TxId, newUtxo.Index)

	_, err = mongo_service.SetOrderUtxoModel(newUtxo)
	if err != nil {
		major.Println(fmt.Sprintf("SetOrderUtxoModel for cold down err:%s", err.Error()))
		return err
	}
	return nil
}

func makePoolRewardPsbt(net, receiveAddress string) (string, error) {
	var (
		netParams                                                 *chaincfg.Params = GetNetParams(net)
		rewardTick                                                string           = "ORXC"
		rewardTickCoinAmount                                      int64            = 1500
		entityPoolClaimOrder                                      *model.OrderBrc20Model
		inscriptionId                                             string = ""
		platformPrivateKeyRewardBrc20, platformAddressRewardBrc20 string = GetPlatformKeyAndAddressForRewardBrc20(net)
		err                                                       error
	)
	entityPoolClaimOrder, err = GetUnoccupiedPoolClaimBrc20PsbtList(net, rewardTick, 1, rewardTickCoinAmount)
	if err != nil {
		return "", err
	}
	if entityPoolClaimOrder == nil {
		return "", errors.New("Pool Claim Order is empty. ")
	}
	inscriptionId = entityPoolClaimOrder.InscriptionId
	if inscriptionId == "" {
		return "", errors.New("InscriptionId of Pool Claim Order is empty. ")
	}
	fmt.Printf("PoolReward:[%s]\n", inscriptionId)
	addr, err := btcutil.DecodeAddress(platformAddressRewardBrc20, netParams)
	if err != nil {
		return "", err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return "", err
	}

	inscriptionIdStrs := strings.Split(inscriptionId, "i")
	inscriptionTxId := inscriptionIdStrs[0]
	inscriptionTxIndex, _ := strconv.ParseInt(inscriptionIdStrs[1], 10, 64)

	inputs := make([]Input, 0)
	inputs = append(inputs, Input{
		OutTxId:  inscriptionTxId,
		OutIndex: uint32(inscriptionTxIndex),
	})

	outputs := make([]Output, 0)
	outputs = append(outputs, Output{
		Address: receiveAddress,
		//Amount:  entityPoolClaimOrder.Amount,
		Amount: 546,
	})
	inputSigns := make([]*InputSign, 0)

	inputSigns = append(inputSigns, &InputSign{
		Index:       0,
		OutRaw:      "",
		PkScript:    hex.EncodeToString(pkScript),
		SighashType: txscript.SigHashSingle | txscript.SigHashAnyOneCanPay,
		PriHex:      platformPrivateKeyRewardBrc20,
		UtxoType:    Witness,
		//Amount:      entityPoolClaimOrder.Amount,
		Amount: 546,
	})

	builder, err := CreatePsbtBuilder(netParams, inputs, outputs)
	if err != nil {
		return "", err
	}
	err = builder.UpdateAndSignInputNoFinalize(inputSigns)
	if err != nil {
		return "", err
	}
	psbtRaw, err := builder.ToString()
	if err != nil {
		return "", err
	}

	return psbtRaw, nil
}

func addPoolRewardPsbt(net, receiveAddress, claimPsbtRaw string) (*PsbtBuilder, error) {
	var (
		netParams                                                 *chaincfg.Params = GetNetParams(net)
		rewardTick                                                string           = "ORXC"
		rewardTickCoinAmount                                      int64            = 1500
		entityPoolClaimOrder                                      *model.OrderBrc20Model
		inscriptionId                                             string = ""
		platformPrivateKeyRewardBrc20, platformAddressRewardBrc20 string = GetPlatformKeyAndAddressForRewardBrc20(net)
		err                                                       error
	)
	entityPoolClaimOrder, err = GetUnoccupiedPoolClaimBrc20PsbtList(net, rewardTick, 1, rewardTickCoinAmount)
	if err != nil {
		return nil, err
	}
	if entityPoolClaimOrder == nil {
		return nil, errors.New("Pool Claim Order is empty. ")
	}
	inscriptionId = entityPoolClaimOrder.InscriptionId
	if inscriptionId == "" {
		return nil, errors.New("InscriptionId of Pool Claim Order is empty. ")
	}
	addr, err := btcutil.DecodeAddress(platformAddressRewardBrc20, netParams)
	if err != nil {
		return nil, err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}

	inscriptionIdStrs := strings.Split(inscriptionId, "i")
	inscriptionTxId := inscriptionIdStrs[0]
	inscriptionTxIndex, _ := strconv.ParseInt(inscriptionIdStrs[1], 10, 64)

	inputs := make([]Input, 0)
	input := Input{
		OutTxId:  inscriptionTxId,
		OutIndex: uint32(inscriptionTxIndex),
	}
	inputs = append(inputs, input)

	outputs := make([]Output, 0)
	outputs = append(outputs, Output{
		Address: receiveAddress,
		Amount:  entityPoolClaimOrder.Amount,
	})
	inputSigns := make([]*InputSign, 0)
	inputSign := &InputSign{
		Index:       0,
		OutRaw:      "",
		PkScript:    hex.EncodeToString(pkScript),
		SighashType: txscript.SigHashAll | txscript.SigHashAnyOneCanPay,
		PriHex:      platformPrivateKeyRewardBrc20,
		UtxoType:    Witness,
		Amount:      546,
	}
	inputSigns = append(inputSigns, inputSign)

	builder, err := NewPsbtBuilder(netParams, claimPsbtRaw)
	if err != nil {
		return nil, err
	}
	err = builder.AddInput(input, inputSign)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("PSBT(X): AddInput err:%s", err.Error()))
	}

	return builder, nil
}

func makeBtcRefundTx(netParams *chaincfg.Params, refundUtxoList []*model.OrderUtxoModel, refundAmount uint64, refundAddress, changeAddress string) (*wire.MsgTx, error) {
	fee := int64(14)
	inputs := make([]*TxInputUtxo, 0)
	for _, u := range refundUtxoList {
		inputs = append(inputs, &TxInputUtxo{
			TxId:     u.TxId,
			TxIndex:  u.Index,
			PkScript: u.PkScript,
			Amount:   u.Amount,
			PriHex:   u.PrivateKeyHex,
		})
	}

	outputs := make([]*TxOutput, 0)
	outputs = append(outputs, &TxOutput{
		Address: refundAddress,
		Amount:  int64(refundAmount),
	})
	tx, err := BuildCommonTx(netParams, inputs, outputs, changeAddress, fee)
	if err != nil {
		fmt.Printf("BuildCommonTx err:%s\n", err.Error())
		return nil, err
	}
	return tx, nil
}

func getRewardRatio(ratio int64) int64 {
	var (
		rewardRatio int64 = 0
	)
	if ratio >= 12 && ratio < 15 {
		rewardRatio = 10
	} else if ratio >= 15 && ratio < 18 {
		rewardRatio = 12
	} else if ratio >= 18 {
		rewardRatio = 15
	}
	return rewardRatio
}
