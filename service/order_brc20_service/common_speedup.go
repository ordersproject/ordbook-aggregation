package order_brc20_service

import (
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/unisat_service"
)

func speedupTx(net string, utxoList []*model.OrderUtxoModel, addUtxoList []*model.OrderUtxoModel, feeRate int64) (string, error) {
	var (
		netParams    *chaincfg.Params = GetNetParams(net)
		speedupTxRaw string           = ""
		speedupTxId  string           = ""
		err          error
	)
	inputs := make([]*TxInputUtxo, 0)
	for _, v := range utxoList {
		addr, err := btcutil.DecodeAddress(v.Address, netParams)
		if err != nil {
			return "", nil
		}
		pkScriptByte, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "", nil
		}
		inputs = append(inputs, &TxInputUtxo{
			TxId:     v.TxId,
			TxIndex:  v.Index,
			PkScript: hex.EncodeToString(pkScriptByte),
			Amount:   uint64(v.Amount),
			PriHex:   v.PrivateKeyHex,
		})
	}

	for _, v := range addUtxoList {
		addr, err := btcutil.DecodeAddress(v.Address, netParams)
		if err != nil {
			return "", nil
		}
		pkScriptByte, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return "", nil
		}
		inputs = append(inputs, &TxInputUtxo{
			TxId:     v.TxId,
			TxIndex:  v.Index,
			PkScript: hex.EncodeToString(pkScriptByte),
			Amount:   uint64(v.Amount),
			PriHex:   v.PrivateKeyHex,
		})
	}

	outputs := make([]*TxOutput, 0)
	for _, v := range utxoList {
		outputs = append(outputs, &TxOutput{
			Address: v.Address,
			Amount:  int64(v.Amount),
		})
	}

	tx, err := BuildTx(netParams, inputs, outputs, feeRate)
	if err != nil {
		fmt.Printf("[REWARD-SEND]BuildCommonTx err:%s\n", err.Error())
		return "", err
	}

	speedupTxRaw, err = ToRaw(tx)
	if err != nil {
		return "", err
	}
	txResp, err := unisat_service.BroadcastTx(net, speedupTxRaw)
	if err != nil {
		return "", err
	}
	speedupTxId = txResp.Result
	return speedupTxId, nil
}

func SpeedupPlatformUtxo(net, txId string, index, feeRate int64, utxoType model.UtxoType) {
	var (
		utxoEntity  *model.OrderUtxoModel
		oldUtxoList []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
		speedupTxId string                  = ""
		err         error

		privateKey string = ""
	)
	utxoEntity, _ = mongo_service.FindOrderUtxoModelByUtxorId(fmt.Sprintf("%s_%d", txId, index))
	if utxoEntity == nil {
		return
	}

	switch utxoType {
	case model.UtxoTypeDummy:
		break
	case model.UtxoTypeDummy1200:
		privateKey, _ = GetPlatformKeyAndAddressReceiveDummyValue(net)
		break
	case model.UtxoTypeDummyBidX:
		privateKey, _ = GetPlatformKeyAndAddressForDummy(net)
		break
	case model.UtxoTypeDummy1200BidX:
		privateKey, _ = GetPlatformKeyAndAddressForDummy(net)
		break
	case model.UtxoTypeMultiInscription:
		privateKey, _ = GetPlatformKeyAndAddressForMultiSigInscription(net)
		break
	case model.UtxoTypeMultiInscriptionFromRelease:
		break
	case model.UtxoTypeRewardInscription:
		privateKey, _ = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		break
	case model.UtxoTypeRewardSend:
		privateKey, _ = GetPlatformKeyAndAddressForRewardBrc20FeeUtxos(net)
		break
	default:
		return
	}

	if utxoEntity.PrivateKeyHex == "" && privateKey != "" {
		utxoEntity.PrivateKeyHex = privateKey
	} else {
		return
	}
	oldUtxoList = append(oldUtxoList, utxoEntity)

	//size * feeRate
	//339 * 50 = 16950
	speedupTxId, err = speedupTx(net, oldUtxoList, nil, feeRate*339)
	if err != nil {
		fmt.Printf("[SPEEDUP]speedupTx err:%s\n", err.Error())
		return
	}

	fmt.Printf("[SPEEDUP]speedupTx:%s\n", speedupTxId)
	return
}
