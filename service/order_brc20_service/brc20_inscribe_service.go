package order_brc20_service

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"ordbook-aggregation/controller/request"
	"ordbook-aggregation/controller/respond"
	"ordbook-aggregation/service/cache_service"
	"ordbook-aggregation/service/inscription_service"
)

func PreInscribe(req *request.Brc20PreReq) (*respond.Brc20PreResp, error) {
	var (
		fromPrivateKeyHex, fromTaprootAddress string = "", ""
		fee int64 = 0
		err error
		netParams *chaincfg.Params = GetNetParams(req.Net)
	)
	fromPrivateKeyHex, fromTaprootAddress, fee, err = inscription_service.CreateKeyAndCalculateInscribe(netParams, req.ReceiveAddress, req.Content, req.FeeRate)
	if err != nil {
		return nil, err
	}
	_ = fromPrivateKeyHex

	cache_service.GetInscribeItemMap().Set(fromTaprootAddress, &cache_service.InscribeInfo{
		FromPrivateKeyHex: fromPrivateKeyHex,
		Content:           req.Content,
		ToAddress:         req.ReceiveAddress,
		Fee:               fee,
		FeeRate:           req.FeeRate,
	})

	return &respond.Brc20PreResp{
		FeeAddress: fromTaprootAddress,
		Fee:        fee,
	}, nil
}

func CommitInscribe(req *request.Brc20CommitReq) (*respond.Brc20CommitResp, error) {
	var (
		commitTxHash, revealTxHash, inscriptionId string = "", "", ""
		err                         error
		fromPriKeyHex, toAddress, content = "", "", ""
		netParams *chaincfg.Params = GetNetParams(req.Net)
		inscribeUtxoList []*inscription_service.InscribeUtxo = make([]*inscription_service.InscribeUtxo, 0)
	)
	inscribeInfo, isExist := cache_service.GetInscribeItemMap().Get(req.FeeAddress)
	if !isExist {
		return nil, errors.New("pre request has not been done")
	}
	fromPriKeyHex, toAddress, content = inscribeInfo.FromPrivateKeyHex, inscribeInfo.ToAddress, inscribeInfo.Content
	fmt.Println(fromPriKeyHex)
	fmt.Println(toAddress)
	fmt.Println(content)
	fmt.Println(inscribeInfo.FeeRate)

	if req.Utxos != nil && len(req.Utxos) != 0 {
		for _, v := range req.Utxos {
			inscribeUtxoList = append(inscribeUtxoList, &inscription_service.InscribeUtxo{
				OutTx:     v.OutTx,
				OutIndex:  v.OutIndex,
				OutAmount: v.OutAmount,
			})
		}
		commitTxHash, revealTxHash, inscriptionId, err = inscription_service.InscribeOneDataFromUtxo(netParams, fromPriKeyHex, toAddress, content, inscribeInfo.FeeRate, "", inscribeUtxoList)
		if err != nil {
			return nil, err
		}
	}else {
		commitTxHash, revealTxHash, inscriptionId, err = inscription_service.InscribeOneData(netParams, fromPriKeyHex, toAddress, content, inscribeInfo.FeeRate, "")
		if err != nil {
			return nil, err
		}
	}

	return &respond.Brc20CommitResp{
		CommitTxHash:  commitTxHash,
		RevealTxHash:  revealTxHash,
		InscriptionId: inscriptionId,
	}, nil
}

