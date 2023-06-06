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
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/tool"
	"strings"
)

type UtxoType int

const (
	NonWitness UtxoType = 1
	Witness    UtxoType = 2
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
	UtxoType    UtxoType             `json:"utxo_type"`
	Index       int                  `json:"index"`
	OutRaw      string               `json:"out_raw"`
	PkScript    string               `json:"pk_script"`
	Amount      uint64               `json:"amount"`
	SighashType txscript.SigHashType `json:"sighash_type"`
	PriHex      string               `json:"pri_hex"`
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
		privateKeyBytes, err := hex.DecodeString(v.PriHex)
		if err != nil {
			return err
		}
		privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
		sigScript := []byte{}
		switch v.UtxoType {
		case NonWitness:
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
			sigScript, err = txscript.RawTxInSignature(s.PsbtUpdater.Upsbt.UnsignedTx, v.Index, s.PsbtUpdater.Upsbt.Inputs[v.Index].NonWitnessUtxo.TxOut[s.PsbtUpdater.Upsbt.UnsignedTx.TxIn[v.Index].PreviousOutPoint.Index].PkScript, v.SighashType, privateKey)
			if err != nil {
				return err
			}
			break
		case Witness:
			witnessUtxoScriptHex, err := hex.DecodeString(
				v.PkScript)
			if err != nil {
				return err
			}
			txout := wire.TxOut{Value: int64(v.Amount), PkScript: witnessUtxoScriptHex}
			err = s.PsbtUpdater.AddInWitnessUtxo(&txout, v.Index)
			if err != nil {
				return err
			}
			err = s.PsbtUpdater.AddInSighashType(v.SighashType, v.Index)
			if err != nil {
				return err
			}
			prevOutputFetcher := NewPrevOutputFetcher(s.PsbtUpdater.Upsbt.Inputs[v.Index].WitnessUtxo.PkScript, s.PsbtUpdater.Upsbt.Inputs[v.Index].WitnessUtxo.Value)
			sigHashes := txscript.NewTxSigHashes(s.PsbtUpdater.Upsbt.UnsignedTx, prevOutputFetcher)
			sigScript, err = txscript.RawTxInWitnessSignature(s.PsbtUpdater.Upsbt.UnsignedTx, sigHashes, v.Index, s.PsbtUpdater.Upsbt.Inputs[v.Index].WitnessUtxo.Value, s.PsbtUpdater.Upsbt.Inputs[v.Index].WitnessUtxo.PkScript, v.SighashType, privateKey)
			if err != nil {
				return err
			}
			break
		}
		fmt.Printf("sigScript: %s\n", hex.EncodeToString(sigScript))

		publicKey := hex.EncodeToString(privateKey.PubKey().SerializeCompressed())
		pubByte, err := hex.DecodeString(publicKey)
		if err != nil {
			return err
		}
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

func (s *PsbtBuilder) AddPartialSigIn(partialSigs []*psbt.PartialSig, index int) error {
	s.PsbtUpdater.Upsbt.Inputs[index].PartialSigs = partialSigs
	if err := s.PsbtUpdater.Upsbt.SanityCheck(); err != nil {
		return err
	}
	return nil
}

func (s *PsbtBuilder) ToString() (string, error) {
	var b bytes.Buffer
	err := s.PsbtUpdater.Upsbt.Serialize(&b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b.Bytes()), nil
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
	txHash, err := chainhash.NewHashFromStr(in.OutTxId)
	if err != nil {
		return err
	}
	s.PsbtUpdater.Upsbt.UnsignedTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: *wire.NewOutPoint(txHash, in.OutIndex),
		Sequence:         wire.MaxTxInSequenceNum,
	})
	s.PsbtUpdater.Upsbt.Inputs = append(s.PsbtUpdater.Upsbt.Inputs, psbt.PInput{})



	privateKeyBytes, err := hex.DecodeString(signIn.PriHex)
	if err != nil {
		return err
	}
	privateKey, _ := btcec.PrivKeyFromBytes(privateKeyBytes)
	sigScript := []byte{}
	switch signIn.UtxoType {
	case NonWitness:
		tx := wire.NewMsgTx(2)
		nonWitnessUtxoHex, err := hex.DecodeString(signIn.OutRaw)
		if err != nil {
			return err
		}
		err = tx.Deserialize(bytes.NewReader(nonWitnessUtxoHex))
		if err != nil {
			return err
		}
		err = s.PsbtUpdater.AddInNonWitnessUtxo(tx, signIn.Index)
		if err != nil {
			return err
		}
		err = s.PsbtUpdater.AddInSighashType(signIn.SighashType, signIn.Index)
		if err != nil {
			return err
		}
		sigScript, err = txscript.RawTxInSignature(s.PsbtUpdater.Upsbt.UnsignedTx, signIn.Index, s.PsbtUpdater.Upsbt.Inputs[signIn.Index].NonWitnessUtxo.TxOut[s.PsbtUpdater.Upsbt.UnsignedTx.TxIn[signIn.Index].PreviousOutPoint.Index].PkScript, signIn.SighashType, privateKey)
		if err != nil {
			return err
		}
		break
	case Witness:
		witnessUtxoScriptHex, err := hex.DecodeString(
			signIn.PkScript)
		if err != nil {
			return err
		}
		txout := wire.TxOut{Value: int64(signIn.Amount), PkScript: witnessUtxoScriptHex}
		err = s.PsbtUpdater.AddInWitnessUtxo(&txout, signIn.Index)
		if err != nil {
			return err
		}
		err = s.PsbtUpdater.AddInSighashType(signIn.SighashType, signIn.Index)
		if err != nil {
			return err
		}
		prevOutputFetcher := NewPrevOutputFetcher(s.PsbtUpdater.Upsbt.Inputs[signIn.Index].WitnessUtxo.PkScript, s.PsbtUpdater.Upsbt.Inputs[signIn.Index].WitnessUtxo.Value)
		sigHashes := txscript.NewTxSigHashes(s.PsbtUpdater.Upsbt.UnsignedTx, prevOutputFetcher)
		sigScript, err = txscript.RawTxInWitnessSignature(s.PsbtUpdater.Upsbt.UnsignedTx, sigHashes, signIn.Index, s.PsbtUpdater.Upsbt.Inputs[signIn.Index].WitnessUtxo.Value, s.PsbtUpdater.Upsbt.Inputs[signIn.Index].WitnessUtxo.PkScript, signIn.SighashType, privateKey)
		if err != nil {
			return err
		}
		break
	}
	fmt.Printf("sigScript: %s\n", hex.EncodeToString(sigScript))
	publicKey := hex.EncodeToString(privateKey.PubKey().SerializeCompressed())
	pubByte, err := hex.DecodeString(publicKey)
	if err != nil {
		return err
	}
	res, err := s.PsbtUpdater.Sign(signIn.Index, sigScript, pubByte, nil, nil)
	if err != nil || res != 0 {
		return err
	}
	_, err = psbt.MaybeFinalize(s.PsbtUpdater.Upsbt, signIn.Index)
	if err != nil {
		return err
	}
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

type PrevOutputFetcher struct{
	pkScript []byte
	value int64
}


func NewPrevOutputFetcher(pkScript []byte, value int64) *PrevOutputFetcher {
	return &PrevOutputFetcher{
		pkScript,
		value,
	}
}


func (d *PrevOutputFetcher) FetchPrevOutput(wire.OutPoint) *wire.TxOut{
	return &wire.TxOut{
		Value:    d.value,
		PkScript: d.pkScript,
	}
}




func CheckOrdinals(preTxOut *wire.TxIn) (*hiro_service.HiroInscription, error) {
	output := fmt.Sprintf("%s:%d", preTxOut.PreviousOutPoint.Hash.String(), preTxOut.PreviousOutPoint.Index)
	inscription, err := hiro_service.GetOutInscription(output)
	return inscription, err
}

func CheckBrc20Ordinals(preTxOut *wire.TxIn, tick, address string) (*oklink_service.BalanceItem, error) {
	inscriptionId := fmt.Sprintf("%si%d", preTxOut.PreviousOutPoint.Hash.String(), preTxOut.PreviousOutPoint.Index)
	inscriptionResp, err := oklink_service.GetInscriptions("", inscriptionId, "", 1, 5)
	if err != nil {
		return nil, err
	}
	has := false
	for _, v := range inscriptionResp.InscriptionsList {
		if inscriptionId == v.InscriptionId &&
			v.State == "success" &&
			strings.ToLower(v.TokenType) == "brc20" &&
			address == v.OwnerAddress{

			has = true
			break
		}
	}
	if !has {
		return nil, errors.New("Not a valid inscription. ")
	}

	brc20Resp, err := oklink_service.GetAddressBrc20BalanceResult(address, tick, 1, 50)
	if err != nil {
		return nil, err
	}
	item := &oklink_service.BalanceItem{}
	for _, v := range brc20Resp.TransferBalanceList {
		if inscriptionId == v.InscriptionId {
			item = v
			break
		}
	}
	if item.Amount == "" {
		return nil, errors.New("Not a valid brc20. ")
	}
	return item, err
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