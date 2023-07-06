package unisat

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/pkg/errors"
	"ordbook-aggregation/node"
	"ordbook-aggregation/service/inscription_service/pkg/btcapi/mempool"
	"ordbook-aggregation/service/unisat_service"
	"strings"
)

func (c *UniSatClient) GetRawTransaction(txHash *chainhash.Hash) (*wire.MsgTx, error) {

	net := "livenet"
	netParams := &chaincfg.MainNetParams

	if strings.Contains(c.baseURL, "testnet") {
		net = "testnet"
		netParams = &chaincfg.TestNet3Params
	}
	txHex, err := node.GetRawTx(net, txHash.String())
	if err != nil {
		btcApiClient := mempool.NewClient(netParams)
		return btcApiClient.GetRawTransaction(txHash)
		//return nil, err
	}
	txByte, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	if err := tx.Deserialize(bytes.NewReader(txByte)); err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *UniSatClient) BroadcastTx(tx *wire.MsgTx) (*chainhash.Hash, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, err
	}

	//res, err := c.request(http.MethodPost, "/tx", strings.NewReader(hex.EncodeToString(buf.Bytes())))
	//if err != nil {
	//	return nil, err
	//}

	net := "livenet"
	if strings.Contains(c.baseURL, "testnet") {
		net = "testnet"
	}

	txPsbtXResp, err := unisat_service.BroadcastTx(net, hex.EncodeToString(buf.Bytes()))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Broadcast %s err:%s", net, err.Error()))
	}

	txHash, err := chainhash.NewHashFromStr(txPsbtXResp.Result)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse tx hash, %s", txPsbtXResp.Result))
	}
	return txHash, nil
}
