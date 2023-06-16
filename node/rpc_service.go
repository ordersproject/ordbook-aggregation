package node

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