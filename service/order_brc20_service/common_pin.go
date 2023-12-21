package order_brc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/unisat_service"
)

// make pin tx for preventing attacks
func MakePinTxForBidX(net, bidXTxId string, bidXTxIndex int64, fromPoolOrderId string) error {
	var (
		utxoList                                                                  []*model.OrderUtxoModel
		addUtxoList                                                               []*model.OrderUtxoModel
		bidXUtxo                                                                  *model.OrderUtxoModel
		pinTx                                                                     *wire.MsgTx
		err                                                                       error
		bidXUtxoId                                                                string = fmt.Sprintf("%s_%d", bidXTxId, bidXTxIndex)
		feeRate                                                                   int64  = 50
		pinTxRaw                                                                  string = ""
		pinTxId                                                                   string = ""
		platformPrivateKeyReceiveBidValueToReturn, _                              string = GetPlatformKeyAndAddressReceiveBidValueToReturn(net)
		platformPrivateKeyMultiSigInscription, platformAddressMultiSigInscription string = GetPlatformKeyAndAddressForMultiSigInscription(net)
	)
	bidXUtxo, _ = mongo_service.FindOrderUtxoModelByUtxorId(bidXUtxoId)
	if bidXUtxo == nil {
		return errors.New("bidXUtxo is nil")
	}
	if bidXUtxo.UsedState != model.UsedNo {
		return errors.New("bidXUtxo is used")
	}
	utxoList = append(utxoList, bidXUtxo)

	addUtxoList, err = GetUnoccupiedUtxoList(net, 1, 0, model.UtxoTypeMultiInscription, fromPoolOrderId, 0)
	defer ReleaseUtxoList(addUtxoList)
	if err != nil {
		return err
	}
	for _, v := range addUtxoList {
		if v.Address != platformAddressMultiSigInscription {
			continue
		}
		v.PrivateKeyHex = platformPrivateKeyMultiSigInscription
		if v.NetworkFeeRate != 0 {
			feeRate = v.NetworkFeeRate
		}
	}

	pinTx, err = MakePinTx(net, utxoList, addUtxoList, feeRate)
	if err != nil {
		return err
	}

	pinTxRaw, err = ToRaw(pinTx)
	if err != nil {
		return err
	}
	txResp, err := unisat_service.BroadcastTx(net, pinTxRaw)
	if err != nil {
		return err
	}
	pinTxId = txResp.Result
	setUsedBidYUtxo(utxoList, pinTxId)
	setUsedMultiSigInscriptionUtxo(addUtxoList, pinTxId)
	for k, v := range utxoList {
		newBidXUtxoOut := Output{
			Address: v.Address,
			Amount:  uint64(v.Amount),
		}
		SaveNewBidYUtxo10000FromBid(net, newBidXUtxoOut, platformPrivateKeyReceiveBidValueToReturn, int64(k), pinTxId)
	}
	return nil
}

// make pin tx
func MakePinTx(net string, utxoList []*model.OrderUtxoModel, addUtxoList []*model.OrderUtxoModel, feeRate int64) (*wire.MsgTx, error) {
	var (
		netParams *chaincfg.Params = GetNetParams(net)
		pinTx     *wire.MsgTx
		err       error
	)
	inputs := make([]*TxInputUtxo, 0)
	for _, v := range utxoList {
		addr, err := btcutil.DecodeAddress(v.Address, netParams)
		if err != nil {
			return nil, nil
		}
		pkScriptByte, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, nil
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
			return nil, nil
		}
		pkScriptByte, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, nil
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

	pinTx, err = BuildTx(netParams, inputs, outputs, feeRate)
	if err != nil {
		fmt.Printf("BuildTx err:%s\n", err.Error())
		return nil, err
	}
	return pinTx, nil
}
