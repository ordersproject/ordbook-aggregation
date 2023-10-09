package parse

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/wire"
)

func parseTx(txRaw string) {
	txRawByte, _ := hex.DecodeString(txRaw)
	tx := wire.NewMsgTx(2)
	err := tx.Deserialize(bytes.NewReader(txRawByte))
	if err != nil {
		fmt.Printf(fmt.Sprintf("PSBT(Y): txRawPsbtY Deserialize err:%s", err.Error()))
		return
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
}
