package order_brc20_service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	// spendSize is the largest number of bytes of a sigScript
	// which spends a p2pkh output: OP_DATA_73 <sig> OP_DATA_33 <pubkey>
	//SpendSize = 1 + 73 + 1 + 33
	SpendSize = 1 + 73 + 1 + 33

	OutSize   = 31
	OtherSize = 10

	TX_EMPTY_SIZE               = 4 + 1 + 1 + 4
	TX_INPUT_BASE               = 32 + 4 + 1 + 4
	TX_INPUT_PUBKEYHASH         = 107
	TX_INPUT_SEGWIT             = 27
	TX_INPUT_TAPROOT            = 17 // round up 16.5 bytes
	TX_OUTPUT_BASE              = 8 + 1
	TX_OUTPUT_PUBKEYHASH        = 25
	TX_OUTPUT_SCRIPTHASH        = 23
	TX_OUTPUT_SEGWIT            = 22
	TX_OUTPUT_SEGWIT_SCRIPTHASH = 34
)

type TxInputUtxo struct {
	TxId     string
	TxIndex  int64
	PkScript string
	Amount   uint64
	PriHex   string
}

type TxOutput struct {
	Address string
	Amount  int64
}

func BuildCommonTx(netParam *chaincfg.Params, ins []*TxInputUtxo, outs []*TxOutput, changeAddress string, fee int64) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(2)
	totalAmount := int64(0)
	outAmount := int64(0)
	for _, out := range outs {
		addr, err := btcutil.DecodeAddress(out.Address, netParam)
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(out.Amount, pkScript))
		outAmount = outAmount + out.Amount
	}

	for _, in := range ins {
		hash, err := chainhash.NewHashFromStr(in.TxId)
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(hash, uint32(in.TxIndex))
		txIn := wire.NewTxIn(prevOut, nil, nil)
		tx.AddTxIn(txIn)
		totalAmount = totalAmount + int64(in.Amount)
	}

	txSize := tx.SerializeSize() + SpendSize*len(tx.TxIn)
	//txSize := tx.SerializeSize()

	reqFee := btcutil.Amount(txSize * int(fee))
	fmt.Printf("txSize:%d, txSizeOnly:%d, reqFee:%d, totalAmount:%d, outAmount:%d\n", txSize, tx.SerializeSize(), reqFee, totalAmount, outAmount)
	if totalAmount-outAmount < int64(reqFee) {
		return nil, errors.New("Insufficient fee")
	}

	changeVal := totalAmount - outAmount - int64(reqFee)
	if changeVal >= 600 && changeAddress != "" {
		addr, err := btcutil.DecodeAddress(changeAddress, netParam)
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(changeVal, pkScript))
	}

	for i, in := range ins {
		privateKeyBytes, err := hex.DecodeString(in.PriHex)
		if err != nil {
			return nil, err
		}
		privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

		pkScriptByte, err := hex.DecodeString(in.PkScript)
		if err != nil {
			return nil, err
		}
		prevOutputFetcher := NewPrevOutputFetcher(pkScriptByte, int64(in.Amount))
		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

		witnessScript, err := txscript.WitnessSignature(
			tx, sigHashes, i, int64(in.Amount), pkScriptByte,
			txscript.SigHashAll, privateKey, true,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		//tx.TxIn[i].SignatureScript = sigScript
		tx.TxIn[i].Witness = witnessScript
	}

	return tx, nil
}

func BuildTx(netParam *chaincfg.Params, ins []*TxInputUtxo, outs []*TxOutput, fee int64) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(2)
	totalAmount := int64(0)
	outAmount := int64(0)
	for _, out := range outs {
		addr, err := btcutil.DecodeAddress(out.Address, netParam)
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(out.Amount, pkScript))
		outAmount = outAmount + out.Amount
	}

	for _, in := range ins {
		hash, err := chainhash.NewHashFromStr(in.TxId)
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(hash, uint32(in.TxIndex))
		txIn := wire.NewTxIn(prevOut, nil, nil)
		tx.AddTxIn(txIn)
		totalAmount = totalAmount + int64(in.Amount)
	}

	txSize := tx.SerializeSize() + SpendSize*len(tx.TxIn)
	//txSize := tx.SerializeSize()

	reqFee := btcutil.Amount(txSize * int(fee))
	fmt.Printf("txSize:%d, txSizeOnly:%d, reqFee:%d, totalAmount:%d, outAmount:%d\n", txSize, tx.SerializeSize(), reqFee, totalAmount, outAmount)
	if totalAmount-outAmount < int64(reqFee) {
		return nil, errors.New("Insufficient fee")
	}

	for i, in := range ins {
		privateKeyBytes, err := hex.DecodeString(in.PriHex)
		if err != nil {
			return nil, err
		}
		privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

		pkScriptByte, err := hex.DecodeString(in.PkScript)
		if err != nil {
			return nil, err
		}
		prevOutputFetcher := NewPrevOutputFetcher(pkScriptByte, int64(in.Amount))
		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

		witnessScript, err := txscript.WitnessSignature(
			tx, sigHashes, i, int64(in.Amount), pkScriptByte,
			txscript.SigHashAll, privateKey, true,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		tx.TxIn[i].Witness = witnessScript
	}

	return tx, nil
}

func BuildCommonTxV2(netParam *chaincfg.Params, ins []*TxInputUtxo, outs []*TxOutput, changeAddress string, fee int64) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(2)
	totalAmount := int64(0)
	outAmount := int64(0)
	for _, out := range outs {
		addr, err := btcutil.DecodeAddress(out.Address, netParam)
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(out.Amount, pkScript))
		outAmount = outAmount + out.Amount
	}

	for _, in := range ins {
		hash, err := chainhash.NewHashFromStr(in.TxId)
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(hash, uint32(in.TxIndex))
		txIn := wire.NewTxIn(prevOut, nil, nil)
		tx.AddTxIn(txIn)
		totalAmount = totalAmount + int64(in.Amount)
	}

	txSize := tx.SerializeSize() + SpendSize*len(tx.TxIn)
	//txSize := tx.SerializeSize()

	reqFee := btcutil.Amount(txSize * int(fee))
	fmt.Printf("txSize:%d, txSizeOnly:%d, reqFee:%d, totalAmount:%d, outAmount:%d\n", txSize, tx.SerializeSize(), reqFee, totalAmount, outAmount)
	if totalAmount-outAmount < int64(reqFee) {
		return nil, errors.New("Insufficient fee")
	}

	changeVal := totalAmount - outAmount - int64(reqFee)
	if changeVal >= 600 && changeAddress != "" {
		addr, err := btcutil.DecodeAddress(changeAddress, netParam)
		if err != nil {
			return nil, err
		}
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(changeVal, pkScript))
	}

	for i, in := range ins {
		privateKeyBytes, err := hex.DecodeString(in.PriHex)
		if err != nil {
			return nil, err
		}
		privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

		pkScriptByte, err := hex.DecodeString(in.PkScript)
		if err != nil {
			return nil, err
		}
		prevOutputFetcher := NewPrevOutputFetcher(pkScriptByte, int64(in.Amount))
		sigHashes := txscript.NewTxSigHashes(tx, prevOutputFetcher)

		witnessScript, err := txscript.WitnessSignature(
			tx, sigHashes, i, int64(in.Amount), pkScriptByte,
			txscript.SigHashAll|txscript.SigHashAnyOneCanPay, privateKey, true,
		)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		//tx.TxIn[i].SignatureScript = sigScript
		tx.TxIn[i].Witness = witnessScript
	}

	return tx, nil
}

func ToRaw(tx *wire.MsgTx) (string, error) {
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	if err := tx.Serialize(buf); err != nil {
		return "", err
	}
	txHex := hex.EncodeToString(buf.Bytes())
	return txHex, nil
}
