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
	fromPrivateKeyHex, fromTaprootAddress, fee, err = inscription_service.CreateKeyAndCalculateInscribe(netParams, req.ReceiveAddress, req.Content)
	if err != nil {
		return nil, err
	}
	_ = fromPrivateKeyHex

	cache_service.GetMetaNameOrderItemMap().Set(fromTaprootAddress, &cache_service.InscribeInfo{
		FromPrivateKeyHex: fromPrivateKeyHex,
		Content:           req.Content,
		ToTaprootAddress:  req.ReceiveAddress,
		Fee:               fee,
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
		fromPriKeyHex, toTaprootAddress, content = "", "", ""
		netParams *chaincfg.Params = GetNetParams(req.Net)
	)
	inscribeInfo, isExist := cache_service.GetMetaNameOrderItemMap().Get(req.FeeAddress)
	if !isExist {
		return nil, errors.New("pre request has not been done")
	}
	fromPriKeyHex, toTaprootAddress, content = inscribeInfo.FromPrivateKeyHex, inscribeInfo.ToTaprootAddress, inscribeInfo.Content
	fmt.Println(fromPriKeyHex)
	fmt.Println(toTaprootAddress)
	fmt.Println(content)

	commitTxHash, revealTxHash, inscriptionId, err = inscription_service.InscribeOneData(netParams, fromPriKeyHex, toTaprootAddress, content)
	if err != nil {
		return nil, err
	}
	return &respond.Brc20CommitResp{
		CommitTxHash:  commitTxHash,
		RevealTxHash:  revealTxHash,
		InscriptionId: inscriptionId,
	}, nil
}