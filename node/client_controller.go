package node

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"ordbook-aggregation/config"
	"strings"
)

type ClientController struct {
	ClientMap map[string]*Client
}

var (
	RPC_url      string
	RPC_username string
	RPC_password string
)

var (
	MyClientController *ClientController
)

func getNetRpcParams(net string) (string, string, string) {
	if strings.ToLower(net) == "testnet" {
		return config.RpcUrlTestnet, config.RpcUsernameTestnet, config.RpcPasswordTestnet
	}
	return config.RpcUrlMainnet, config.RpcUsernameMainnet, config.RpcPasswordMainnet
}

func NewClientController(net string) *ClientController {
	if MyClientController != nil {
		if _, ok := MyClientController.ClientMap[net]; ok {
			return MyClientController
		}
	} else {
		MyClientController = &ClientController{
			ClientMap: make(map[string]*Client),
		}
	}

	RPC_url, RPC_username, RPC_password = getNetRpcParams(net)

	fmt.Println("*******RPC_url : [ ", RPC_url, " ]")

	accessToken := BasicAuth(RPC_username, RPC_password)
	MyClientController.ClientMap[net] = NewClientNode(RPC_url, accessToken, false)
	fmt.Println("****** Build new Client completed ******")

	return MyClientController
}

func (c *ClientController) BroadcastTx(net, txHexStr string) (string, error) {
	request := []interface{}{
		txHexStr,
		false,
	}

	result, err := c.ClientMap[net].Call("sendrawtransaction", request)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

func (c *ClientController) GetBlockHash(net string, height uint64) (string, error) {

	request := []interface{}{
		height,
	}

	result, err := c.ClientMap[net].Call("getblockhash", request)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func (c *ClientController) GetBlockHeight(net string) (uint64, error) {

	result, err := c.ClientMap[net].Call("getblockcount", nil)
	if err != nil {
		return 0, err
	}

	return result.Uint(), nil
}

func (c *ClientController) GetBlock(net string, hash string, format ...uint64) (*Block, error) {

	request := []interface{}{
		hash,
	}

	if len(format) > 0 {
		request = append(request, format[0])
	}

	result, err := c.ClientMap[net].Call("getblock", request)
	if err != nil {
		return nil, err
	}

	return NewBlock(result), nil
}

func (c *ClientController) GetTxIDsInMemPool(net string) ([]string, error) {

	var (
		txids = make([]string, 0)
	)

	result, err := c.ClientMap[net].Call("getrawmempool", nil)
	if err != nil {
		return nil, err
	}

	if !result.IsArray() {
		return nil, errors.New("no query record")
	}

	for _, txid := range result.Array() {
		txids = append(txids, txid.String())
	}

	return txids, nil
}

func (c *ClientController) GetTransaction(net string, txid string) (*Transaction, error) {

	var (
		result *gjson.Result
		err    error
	)

	request := []interface{}{
		txid,
		true,
	}

	result, err = c.ClientMap[net].Call("getrawtransaction", request)
	if err != nil {

		request = []interface{}{
			txid,
			1,
		}

		result, err = c.ClientMap[net].Call("getrawtransaction", request)
		if err != nil {
			return nil, err
		}
	}

	return newTxByCore(result), nil
}

func (c *ClientController) GetTransactionHex(net string, txid string) (string, error) {

	var (
		result *gjson.Result
		err    error
	)

	request := []interface{}{
		txid,
		false,
	}

	result, err = c.ClientMap[net].Call("getrawtransaction", request)
	if err != nil {

		request = []interface{}{
			txid,
			0,
		}

		result, err = c.ClientMap[net].Call("getrawtransaction", request)
		if err != nil {
			return "", err
		}
	}

	return result.String(), nil
}
