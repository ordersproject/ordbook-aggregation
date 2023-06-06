package inscription_service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"log"
	"ordbook-aggregation/service/create_key"
	"ordbook-aggregation/service/inscription_service/internal/ord"
	"ordbook-aggregation/service/inscription_service/pkg/btcapi"
	"ordbook-aggregation/service/inscription_service/pkg/btcapi/mempool"
)

var (
	testnetPriKey = "5457216ea9134624eb667c68de54ddca9dcb626c9a978a4bb52ba616d1d1285f"
	testnetTaprootAddress = "tb1plc2nakpmxjkp3uqzva3we6hv0txjwv3wfcadxf5ydjslkn86c4hqpvm9y6"

	fakerHash, _ = chainhash.NewHashFromStr("f3125da7f0e0894ae51b1b7b25996026aac45617fa518113f2d28b8277f5a9da")
	fakerPkScript, _ = hex.DecodeString("fe153ed83b34ac18f0026762eceaec7acd27322e4e3ad326846ca1fb4cfac56e")
	unspentFakerList = []*btcapi.UnspentOutput{
			&btcapi.UnspentOutput{
				Outpoint: &wire.OutPoint{
					Hash:  *fakerHash,
					Index: 3,
				},
				Output:   &wire.TxOut{
					Value:    1000000,
					PkScript: fakerPkScript,
				},
			},
		}


)

func CreateKeyAndCalculateInscribe(netParams *chaincfg.Params, toTaprootAddress, content string) (string, string, int64, error) {
	fromPriKeyHex, fromTaprootAddress, err := create_key.CreateTaprootKey(netParams)
	if err != nil {
		return "", "", 0, err
	}

	testnetNetParams := &chaincfg.SigNetParams
	btcApiClient := mempool.NewClient(testnetNetParams)
	contentType := "text/plain;charset=utf-8"
	//dataMap := make(map[string]interface{})

	utxoPrivateKeyHex := testnetPriKey
	destination := testnetTaprootAddress

	commitTxOutPointList := make([]*wire.OutPoint, 0)
	commitTxPrivateKeyList := make([]*btcec.PrivateKey, 0)

	{
		utxoPrivateKeyBytes, err := hex.DecodeString(utxoPrivateKeyHex)
		if err != nil {
			return "", "", 0, err
		}
		utxoPrivateKey, _ := btcec.PrivKeyFromBytes(utxoPrivateKeyBytes)

		utxoTaprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(utxoPrivateKey.PubKey())), testnetNetParams)
		if err != nil {
			return "", "", 0, err
		}
		_ = utxoTaprootAddress

		unspentList, err := btcApiClient.ListUnspent(utxoTaprootAddress)
		if err != nil {
			return "", "", 0, errors.New(fmt.Sprintf("list unspent err %v", err))
		}
		if unspentList == nil || len(unspentList) == 0 {
			return "", "", 0, errors.New(fmt.Sprintf("list unspent is empty"))
		}

		//unspentList := unspentFakerList
		for i := range unspentList {
			//fmt.Println(i)
			//fmt.Println(unspentList[i].Outpoint.String())
			//fmt.Println(unspentList[i].Output.Value)
			//fmt.Println(hex.EncodeToString(unspentList[i].Output.PkScript))
			commitTxOutPointList = append(commitTxOutPointList, unspentList[i].Outpoint)
			commitTxPrivateKeyList = append(commitTxPrivateKeyList, utxoPrivateKey)
		}
	}

	request := ord.InscriptionRequest{
		CommitTxOutPointList:   commitTxOutPointList,
		CommitTxPrivateKeyList: commitTxPrivateKeyList,
		//CommitFeeRate:          2,
		//FeeRate:                1,
		CommitFeeRate:          50,
		FeeRate:                50,
		DataList: []ord.InscriptionData{
			{
				ContentType: contentType,
				Body:        []byte(content),
				Destination: destination,
			},
		},
		SingleRevealTxOnly: false,
	}

	tool, err := ord.NewInscriptionToolWithBtcApiClient(testnetNetParams, btcApiClient, &request)
	if err != nil {
		//log.Fatalf("Failed to create inscription tool: %v", err)
		return "", "", 0, err
	}
	fee := tool.CalculateFee()
	return fromPriKeyHex, fromTaprootAddress, fee, nil
}

func InscribeOneData(netParams *chaincfg.Params, fromPriKeyHex, toTaprootAddress, content string) (string, string, string, error) {
	//netParams := &chaincfg.SigNetParams
	btcApiClient := mempool.NewClient(netParams)
	contentType := "text/plain;charset=utf-8"
	//dataMap := make(map[string]interface{})

	utxoPrivateKeyHex := fromPriKeyHex
	destination := toTaprootAddress

	commitTxOutPointList := make([]*wire.OutPoint, 0)
	commitTxPrivateKeyList := make([]*btcec.PrivateKey, 0)

	{
		utxoPrivateKeyBytes, err := hex.DecodeString(utxoPrivateKeyHex)
		if err != nil {
			return "", "", "", err
		}
		utxoPrivateKey, _ := btcec.PrivKeyFromBytes(utxoPrivateKeyBytes)

		utxoTaprootAddress, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(txscript.ComputeTaprootKeyNoScript(utxoPrivateKey.PubKey())), netParams)
		if err != nil {
			return "", "", "", err
		}

		unspentList, err := btcApiClient.ListUnspent(utxoTaprootAddress)

		if err != nil {
			return "", "", "", errors.New(fmt.Sprintf("list unspent err %v", err))
		}
		if unspentList == nil || len(unspentList) == 0 {
			return "", "", "", errors.New(fmt.Sprintf("list unspent is empty"))
		}

		for i := range unspentList {
			commitTxOutPointList = append(commitTxOutPointList, unspentList[i].Outpoint)
			commitTxPrivateKeyList = append(commitTxPrivateKeyList, utxoPrivateKey)
		}
	}

	request := ord.InscriptionRequest{
		CommitTxOutPointList:   commitTxOutPointList,
		CommitTxPrivateKeyList: commitTxPrivateKeyList,
		//CommitFeeRate:          2,
		//FeeRate:                1,
		CommitFeeRate:          50,
		FeeRate:                50,
		DataList: []ord.InscriptionData{
			{
				ContentType: contentType,
				Body:        []byte(content),
				Destination: destination,
			},
		},
		SingleRevealTxOnly: false,
	}

	tool, err := ord.NewInscriptionToolWithBtcApiClient(netParams, btcApiClient, &request)
	if err != nil {
		return "", "", "", errors.New(fmt.Sprintf("Failed to create inscription tool: %v", err))
	}
	commitTxHash, revealTxHashList, inscriptions, fees, err := tool.Inscribe()
	if err != nil {
		return "", "", "", errors.New(fmt.Sprintf("send tx errr, %v", err))
	}
	log.Println("commitTxHash, " + commitTxHash.String())
	revealTxHash := ""
	for i := range revealTxHashList {
		revealTxHash = revealTxHashList[i].String()
		log.Println("revealTxHash, " + revealTxHashList[i].String())
	}
	inscription := ""
	for i := range inscriptions {
		inscription = inscriptions[i]
		log.Println("inscription, " + inscriptions[i])
	}
	log.Println("fees: ", fees)
	return commitTxHash.String(), revealTxHash, inscription, nil
}