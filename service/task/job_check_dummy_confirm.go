package task

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/order_brc20_service"
)

func JobForCheckUtxoBlock() {
	var (
		net        string                  = "livenet"
		startIndex int64                   = -1
		maxLimit   int64                   = 3000
		utxoList   []*model.OrderUtxoModel = make([]*model.OrderUtxoModel, 0)
	)
	utxoList, _ = mongo_service.FindAllTypeUtxoList(net, startIndex, maxLimit, 0, -1)
	if len(utxoList) == 0 {
		return
	}
	major.Println(fmt.Sprintf("[JOP-UTXO-BLOCK]  len [%d]", len(utxoList)))

	for _, v := range utxoList {
		if v.UsedState != model.UsedNo {
			continue
		}
		if v.TxId == "" {
			continue
		}
		if v.ConfirmStatus == model.Confirmed {
			continue
		}
		block := order_brc20_service.GetTxConfirm(v.TxId)
		if block == 0 {
			continue
		}
		v.ConfirmStatus = model.Confirmed
		err := mongo_service.UpdateOrderUtxoModelForConfirm(v.UtxoId, v.ConfirmStatus)
		if err != nil {
			major.Println(fmt.Sprintf("[JOP-UTXO-BLOCK] UpdateOrderUtxoModelForConfirm err:%s", err))
			continue
		}
		major.Println(fmt.Sprintf("[JOP-UTXO-BLOCK] UpdateOrderUtxoModelForConfirm success [%s]", v.UtxoId))
	}
}
