package order_brc20_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"testing"
)

func Test_createMultiSigAddress(t *testing.T) {
	net := &chaincfg.MainNetParams
	pubKeys := []string{
		"037651f0d9d5f5fd74aa04890168888ce01f26702faba2a5fbd820cbc1c638e7a8",
		"037355ad3caeacd0b8e69fd519bf7aac71c3c0227ae446f0c737e4616d7c1ac4f9",
	}
	multiSigScript, res, res2, err := createMultiSigAddress(net, pubKeys...)
	if err != nil {
		fmt.Printf("Err:%s\n", err.Error())
		return
	}
	fmt.Printf("MultiSigScript:%s\n", multiSigScript)
	fmt.Printf("Res:%s\n", res)
	fmt.Printf("Res2:%s\n", res2)

}

func Test_createMultiSigTx(t *testing.T) {
	var (
		address    string = "bc1qddrzfguuvfruelvnn2v5njvpmarj2zhnd76r9w"
		privateKey string = "d04f8bc50547f16dbd0d4337ffd7a11bf615e397c680cd447e9d31bce638b62f"
		tx         *wire.MsgTx

		changeAddress string = "bc1q98hfp00j259u93szt7cnfgfy38wy8xve3lh3qr"

		outAmount      int64 = 1000
		multiSigScript       = ""

		//utxoTxId     string = "52d7675ce73d88123ecd69f75e7f8f80ab0e102a08bd14edd8fc7323ca2f69b6"
		utxoTxId     string = "d57fb95c15b09b718a8a858f647b19361b1876a8334a1e405c65010f224faacf"
		utxoPkScript string = "00146b4624a39c6247ccfd939a9949c981df47250af3"
		utxoTxIndex  int64  = 0
		utxoValue    int64  = 10000
		fee          int64  = 14
		totalAmount  int64  = 0

		ins   []*TxInputUtxo = make([]*TxInputUtxo, 0)
		txRaw string         = ""
	)
	_ = address

	net := &chaincfg.MainNetParams
	pubKeys := []string{
		"037651f0d9d5f5fd74aa04890168888ce01f26702faba2a5fbd820cbc1c638e7a8",
		"037355ad3caeacd0b8e69fd519bf7aac71c3c0227ae446f0c737e4616d7c1ac4f9",
	}
	multiSigScript, addr1, addr2, err := createMultiSigAddress(net, pubKeys...)
	if err != nil {
		fmt.Printf("createMultiSigAddress err: %s\n", err.Error())
		return
	}
	_ = addr1
	_ = addr2
	multiSigScriptByte, err := hex.DecodeString(multiSigScript)
	if err != nil {
		fmt.Printf("DecodeString err: %s\n", err.Error())
		return
	}

	tx = wire.NewMsgTx(2)

	//add out
	tx.AddTxOut(wire.NewTxOut(outAmount, multiSigScriptByte))

	//add in
	hash, err := chainhash.NewHashFromStr(utxoTxId)
	if err != nil {
		fmt.Printf("NewHashFromStr err: %s\n", err.Error())
		return
	}
	prevOut := wire.NewOutPoint(hash, uint32(utxoTxIndex))
	txIn := wire.NewTxIn(prevOut, nil, nil)
	tx.AddTxIn(txIn)
	totalAmount = totalAmount + utxoValue

	txSize := tx.SerializeSize() + SpendSize*len(tx.TxIn)

	reqFee := btcutil.Amount(txSize * int(fee))
	fmt.Printf("txSize:%d, reqFee:%d, totalAmount:%d, outAmount:%d\n", txSize, reqFee, totalAmount, outAmount)
	if totalAmount-outAmount < int64(reqFee) {
		fmt.Printf("NewHashFromStr err: %s\n", errors.New("Insufficient fee"))
		return
	}

	changeVal := totalAmount - outAmount - int64(reqFee)
	fmt.Printf("changeVal:%d\n", changeVal)
	if changeVal >= 600 {
		addr, err := btcutil.DecodeAddress(changeAddress, net)
		if err != nil {
			fmt.Printf("DecodeAddress err: %s\n", err.Error())
			return
		}
		//addrHash, err := btcutil.NewAddressWitnessPubKeyHash(addr.ScriptAddress(), netParam)
		//if err != nil {
		//	return nil, err
		//}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			fmt.Printf("PayToAddrScript err: %s\n", err.Error())
			return
		}
		tx.AddTxOut(wire.NewTxOut(changeVal, pkScript))
	}

	ins = append(ins, &TxInputUtxo{
		TxId:     utxoTxId,
		TxIndex:  utxoTxIndex,
		PkScript: utxoPkScript,
		Amount:   uint64(utxoValue),
		PriHex:   privateKey,
	})
	for i, in := range ins {
		privateKeyBytes, err := hex.DecodeString(in.PriHex)
		if err != nil {
			fmt.Printf("DecodeString err: %s\n", err.Error())
			return
		}
		privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

		pkScriptByte, err := hex.DecodeString(in.PkScript)
		if err != nil {
			fmt.Printf("DecodeString err: %s\n", err.Error())
			return
		}

		prevOutputFetcher := NewPrevOutputFetcher(pkScriptByte, int64(in.Amount))
		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

		witnessScript, err := txscript.WitnessSignature(
			tx, sigHashes, i, int64(in.Amount), pkScriptByte,
			txscript.SigHashAll, privateKey, true,
		)
		if err != nil {
			fmt.Println(err)
			return
		}
		tx.TxIn[i].Witness = witnessScript
	}
	txRaw, err = ToRaw(tx)
	if err != nil {
		fmt.Printf("ToRaw err: %s\n", err.Error())
		return
	}
	fmt.Printf("Raw:%s\n", txRaw)

	//f4a2ff0b67ba81bf576113a605ab6c509c6204062ca2c75b47ab0ef9c5da6a86
	//https://mempool.space/zh/tx/f4a2ff0b67ba81bf576113a605ab6c509c6204062ca2c75b47ab0ef9c5da6a86
}
