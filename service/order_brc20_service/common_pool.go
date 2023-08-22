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
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/unisat_service"
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

func makeMultiSigTx(outTxId string, outIndex int64) {

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
	total, _ = mongo_service.CountPoolBrc20ModelList(net, tick, "", "", model.PoolTypeTick, model.PoolStateAdd)
	entityList, _ = mongo_service.FindPoolBrc20ModelList(net, tick, "", "", model.PoolTypeTick, model.PoolStateAdd,
		limit, flag, page, "", 0)
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
}

func inscriptionMultiSigTransfer(poolOrderId string) {
	var (
		poolOrder          *model.PoolBrc20Model
		inscriptionAddress string           = ""
		netParams          *chaincfg.Params = &chaincfg.MainNetParams
	)
	poolOrder, _ = mongo_service.FindPoolBrc20ModelByOrderId(poolOrderId)
	if poolOrder == nil {
		return
	}
	netParams = GetNetParams(poolOrder.Net)
	if poolOrder.MultiSigScriptAddress == "" {
		MultiSigScriptByte, err := hex.DecodeString(poolOrder.MultiSigScript)
		if err != nil {
			return
		}
		h := sha256.New()
		h.Write(MultiSigScriptByte)
		nativeSegwitAddress, err := btcutil.NewAddressWitnessScriptHash(h.Sum(nil), netParams)
		if err != nil {
			fmt.Println("Failed to create native SegWit address:", err)
			return
		}
		poolOrder.MultiSigScriptAddress = nativeSegwitAddress.EncodeAddress()
	}
	inscriptionAddress = poolOrder.MultiSigScriptAddress
	if inscriptionAddress == "" {
		return
	}

	//commitTxHash, revealTxHashList, inscriptionIdList, fees, err =
	//	inscription_service.InscribeMultiDataFromUtxo(netParams, req.PriKeyHex, inscriptionAddress,
	//		transferContent, req.FeeRate, req.ChangeAddress, 1, inscribeUtxoList, "segwit", false)
	//if err != nil {
	//	return nil, err
	//}

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
	total, _ = mongo_service.CountPoolBrc20ModelList(net, tick, "", address, model.PoolTypeTick, model.PoolStateAdd)
	entityList, _ = mongo_service.FindPoolBrc20ModelList(net, tick, "", address, model.PoolTypeTick, model.PoolStateAdd,
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
	if poolOrder.PoolState == model.PoolStateUsed {
		poolOrder.PoolState = model.PoolStateClaim
	}
	if poolOrder.PoolCoinState == model.PoolStateUsed {
		poolOrder.PoolCoinState = model.PoolStateClaim
	}
	err = mongo_service.SetPoolBrc20ModelForClaim(poolOrder)
	if err != nil {
		return err
	}
	return nil
}
