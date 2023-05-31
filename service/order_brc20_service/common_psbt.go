package order_brc20_service

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/hiro_service"
	"ordbook-aggregation/tool"
)

type PsbtBuilder struct {
	NetParams *chaincfg.Params
	PsbtUpdater *psbt.Updater
}

type Input struct {
	OutTxId string `json:"out_tx_id"`
	OutIndex uint32 `json:"out_index"`
}

type InputSign struct {
	Index       int                `json:"index"`
	OutRaw      string               `json:"out_raw"`
	SighashType txscript.SigHashType `json:"sighash_type"`
	PriHex string `json:"pri_hex"`
}

type Output struct {
	Address string `json:"address"`
	Amount uint64 `json:"amount"`
}

func CreatePsbtBuilder(netParams *chaincfg.Params, ins []Input, outs []Output) (*PsbtBuilder, error)  {
	var(
		txOuts []*wire.TxOut = make([]*wire.TxOut, 0)
		txIns []*wire.OutPoint = make([]*wire.OutPoint, 0)
		nSequences []uint32 = make([]uint32, 0)
	)
	for _, in := range ins{
		txHash, err := chainhash.NewHashFromStr(in.OutTxId)
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(txHash, in.OutIndex)
		txIns = append(txIns, prevOut)
		nSequences = append(nSequences, wire.MaxTxInSequenceNum)
	}

	for _, out := range outs {
		address, err := btcutil.DecodeAddress(out.Address, netParams)
		if err != nil {
			return nil, err
		}

		pkScript, err := txscript.PayToAddrScript(address)
		if err != nil {
			return nil, err
		}

		txOut := wire.NewTxOut(int64(out.Amount), pkScript)
		txOuts = append(txOuts, txOut)
	}

	cPsbt, err := psbt.New(txIns, txOuts, int32(2), uint32(0), nSequences)
	if err != nil {
		return nil, err
	}
	psbtBuilder := &PsbtBuilder{NetParams:netParams}

	psbtBuilder.PsbtUpdater, err = psbt.NewUpdater(cPsbt)
	if err != nil {
		return nil, err
	}
	return psbtBuilder, nil
}

func (s *PsbtBuilder) UpdateAndSignInput(signIns []InputSign) error {
	for _, v := range signIns {
		tx := wire.NewMsgTx(2)
		nonWitnessUtxoHex, err := hex.DecodeString(v.OutRaw)
		if err != nil {
			return err
		}
		err = tx.Deserialize(bytes.NewReader(nonWitnessUtxoHex))
		if err != nil {
			return err
		}
		err = s.PsbtUpdater.AddInNonWitnessUtxo(tx, v.Index)
		if err != nil {
			return err
		}
		err = s.PsbtUpdater.AddInSighashType(v.SighashType, v.Index)
		if err != nil {
			return err
		}

		privateKeyBytes, err := hex.DecodeString(v.PriHex)
		if err != nil {
			return err
		}
		privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)

		sigScript := []byte{}
		sigScript, err = txscript.RawTxInSignature(s.PsbtUpdater.Upsbt.UnsignedTx, v.Index, s.PsbtUpdater.Upsbt.Inputs[v.Index].NonWitnessUtxo.TxOut[s.PsbtUpdater.Upsbt.UnsignedTx.TxIn[v.Index].PreviousOutPoint.Index].PkScript, v.SighashType, privateKey)
		if err != nil {
			return err
		}
		publicKey := hex.EncodeToString(privateKey.PubKey().SerializeCompressed())

		fmt.Printf("sigScript: %s\n", hex.EncodeToString(sigScript))
		pubByte, err := hex.DecodeString(publicKey)
		res, err := s.PsbtUpdater.Sign(v.Index, sigScript, pubByte, nil, nil)
		if err != nil || res != 0 {
			return err
		}
		_, err = psbt.MaybeFinalize(s.PsbtUpdater.Upsbt, v.Index)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PsbtBuilder) ToString() (string, error) {
	var b bytes.Buffer
	err := s.PsbtUpdater.Upsbt.Serialize(&b)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}


func NewPsbtBuilder(netParams *chaincfg.Params, psbtHex string) (*PsbtBuilder, error) {
	psbtBuilder := &PsbtBuilder{NetParams:netParams}

	b, err := hex.DecodeString(psbtHex)
	if err != nil {
		return nil, err
	}
	p, err := psbt.NewFromRawBytes(bytes.NewReader(b), false)
	if err != nil {
		return nil, err
	}
	psbtBuilder.PsbtUpdater, err = psbt.NewUpdater(p)
	if err != nil {
		return nil, err
	}
	return psbtBuilder, nil
}

func (s *PsbtBuilder) GetInputs() []*wire.TxIn {
	return s.PsbtUpdater.Upsbt.UnsignedTx.TxIn
}

func (s *PsbtBuilder) GetOutputs() []*wire.TxOut {
	return s.PsbtUpdater.Upsbt.UnsignedTx.TxOut
}

func (s *PsbtBuilder) AddInput(in Input, signIn InputSign) error {
	return nil
}

func (s *PsbtBuilder) IsComplete() bool {
	return s.PsbtUpdater.Upsbt.IsComplete()
}

func (s *PsbtBuilder) ExtractPsbtTransaction() (string, error) {
	if !s.IsComplete() {
		err := psbt.MaybeFinalizeAll(s.PsbtUpdater.Upsbt)
		if err != nil {
			return "", err
		}
	}

	tx, err := psbt.Extract(s.PsbtUpdater.Upsbt)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = tx.Serialize(&b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b.Bytes()), nil
}



func CheckOrdinals(preTxOut *wire.TxIn) (*hiro_service.HiroInscription, error) {
	output := fmt.Sprintf("%s:%d", preTxOut.PreviousOutPoint.Hash.String(), preTxOut.PreviousOutPoint.Index)
	inscription, err := hiro_service.GetOutInscription(output)
	return inscription, err
}

func GetInscriptionContent(inscriptionId string) (interface{}, error) {
	return hiro_service.GetInscriptionContent(inscriptionId)
}

func GetBrc20Data(inscriptionId string) (*model.Brc20Protocol, error) {
	content, err := hiro_service.GetInscriptionContent(inscriptionId)
	if err != nil {
		return nil, err
	}
	data := &model.Brc20Protocol{}
	if err = tool.JsonToAny(content, &data) ; err != nil {
		return nil, errors.New(fmt.Sprintf("Parse Brc20 data err:%s", err))
	}
	return data, nil
}