package task

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/service/oklink_service"
	"ordbook-aggregation/service/order_brc20_service"
	"ordbook-aggregation/tool"
	"strings"
	"time"
)

func LoopForCheckAsk() {
	var (
		t                                                            = tool.MakeTimestamp()
		net                               string                     = "livenet"
		tick                              string                     = "rdex"
		limit                             int64                      = 20
		utxoList                          []*oklink_service.UtxoItem = make([]*oklink_service.UtxoItem, 0)
		_, platformAddressSendBrc20ForAsk string                     = order_brc20_service.GetPlatformKeyAndAddressSendBrc20ForAsk(net)
	)
	entityList, _ := mongo_service.FindOrderBrc20ModelList(net, tick, "", "",
		model.OrderTypeSell, model.OrderStateCreate,
		limit, 0, 0, "timestamp", 1, 0)
	if entityList == nil || len(entityList) == 0 {
		return
	}

	time.Sleep(15 * time.Second)

	for i := int64(0); i < 10; i++ {
		utxoResp, err := oklink_service.GetAddressUtxo(platformAddressSendBrc20ForAsk, i+1, 50)
		if err != nil {
			fmt.Printf("[LOOP-CHECK]-%s\n", fmt.Sprintf("PSBT(X): Recheck address utxo list err:%s", err.Error()))
			return
		}
		if utxoResp.UtxoList != nil && len(utxoResp.UtxoList) != 0 {
			utxoList = append(utxoList, utxoResp.UtxoList...)
		} else {
			break
		}
		time.Sleep(10 * time.Second)
	}

	for _, v := range entityList {
		inscriptionIdStrs := strings.Split(v.InscriptionId, "i")
		if len(inscriptionIdStrs) < 2 {
			continue
		}
		inscriptionTxId := inscriptionIdStrs[0]

		has := false
		for _, liveUtxo := range utxoList {
			if inscriptionTxId == liveUtxo.TxId {
				has = true
				break
			}
		}

		if !has {
			v.OrderState = model.OrderStateFinishButErr
			_, err := mongo_service.SetOrderBrc20Model(v)
			if err != nil {
				major.Println(fmt.Sprintf("[LOOP-CHECK]update for ask in orderId finishERR:%s, err:%s", v.OrderId, err.Error()))
				continue
			}
			major.Println(fmt.Sprintf("[LOOP-CHECK]update for ask in orderId:%s, finishERR, success", v.OrderId))
		}
	}
	major.Println(fmt.Sprintf("[LOOP-CHECK]check success, time:%d", tool.MakeTimestamp()-t))
}
