package node

import "fmt"

func BroadcastTx(net, txHex string) (string, error) {
	client := NewClientController(net)
	txId, err := client.BroadcastTx(net, txHex)
	return txId, err
}

func GetRawTx(net, txId string) (string, error) {
	client := NewClientController(net)
	txRaw, err := client.GetTransactionHex(net, txId)
	return txRaw, err
}

func CurrentBlockHeight(net string) (uint64, error) {
	client := NewClientController(net)
	return client.GetBlockHeight(net)
}

func GetTx(net, txId string) (*Transaction, error) {
	client := NewClientController(net)
	tx, err := client.GetTransaction(net, txId)
	if err != nil {
		fmt.Printf("[RPC]err:%s\n", err.Error())
		return nil, err
	}
	//fmt.Printf("[RPC]tx:%+v\n", tx)
	return tx, err
}

func GetBlockInfo(net string, blockHeight int64) (*Block, error) {
	client := NewClientController(net)
	blockHash, err := client.GetBlockHash(net, uint64(blockHeight))
	if err != nil {
		fmt.Printf("[RPC]err:%s\n", err.Error())
		return nil, err
	}
	blockInfo, err := client.GetBlock(net, blockHash)
	//fmt.Printf("[RPC]tx:%+v\n", tx)
	return blockInfo, err
}
