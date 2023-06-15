package node

func BroadcastTx(net, txHex string) (string, error) {
	client := NewClientController(net)
	txId, err := client.BroadcastTx(net, txHex)
	return txId, err
}
