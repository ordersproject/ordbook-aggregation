package parse

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/wire"
	"ordbook-aggregation/service/order_brc20_service"
)

func parseTx(txRaw string) *wire.MsgTx {
	txRawByte, _ := hex.DecodeString(txRaw)
	tx := wire.NewMsgTx(2)
	err := tx.Deserialize(bytes.NewReader(txRawByte))
	if err != nil {
		fmt.Printf(fmt.Sprintf("PSBT(Y): txRawPsbtY Deserialize err:%s", err.Error()))
		return nil
	}

	//
	fmt.Printf("Tx:\n")
	fmt.Printf("%+v\n", tx)
	for k, in := range tx.TxIn {
		fmt.Printf("Input-[%d] OutPoint:%s, Witness:%s\n", k, in.PreviousOutPoint.String(), hex.EncodeToString(in.Witness[0]))
	}
	for k, out := range tx.TxOut {
		fmt.Printf("Output-[%d] Value:%d PkScript:%s\n", k, out.Value, hex.EncodeToString(out.PkScript))
	}
	fmt.Printf("\n")
	return tx
}

func parsePsbt(psbtRaw string) {
	psbtBuilder, err := order_brc20_service.NewPsbtBuilder(order_brc20_service.GetNetParams("livenet"), psbtRaw)
	if err != nil {
		fmt.Printf(fmt.Sprintf("NewPsbtBuilder err:%s", err.Error()))
		return
	}
	txRawPsbtY, err := psbtBuilder.ExtractPsbtTransaction()
	if err != nil {
		fmt.Printf(fmt.Sprintf("ExtractPsbtTransaction err:%s", err.Error()))
		return
	}
	parseTx(txRawPsbtY)
}
