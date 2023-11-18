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
		t           = tool.MakeTimestamp()
		net  string = "livenet"
		tick string = "rdex"
		//tick     string                     = ""
		limit    int64                      = 100
		utxoList []*oklink_service.UtxoItem = make([]*oklink_service.UtxoItem, 0)
	)
	entityList, _ := mongo_service.FindOrderBrc20ModelList(net, tick, "", "",
		model.OrderTypeSell, model.OrderStateCreate,
		limit, 0, 0, "timestamp", 1, 0, 0)
	if entityList == nil || len(entityList) == 0 {
		return
	}
	major.Println(fmt.Sprintf("[LOOP-CHECK-ASK]ask order len:%d", len(entityList)))
	time.Sleep(15 * time.Second)
	for _, v := range entityList {
		inscriptionId := v.InscriptionId
		if strings.Contains(inscriptionId, ":") {
			inscriptionId = strings.ReplaceAll(inscriptionId, ":", "i")
		}

		inscriptionIdStrs := strings.Split(inscriptionId, "i")
		if len(inscriptionIdStrs) < 2 {
			continue
		}
		inscriptionTxId := inscriptionIdStrs[0]

		for i := int64(0); i < 50; i++ {
			utxoResp, err := oklink_service.GetAddressUtxo(v.SellerAddress, i+1, 100)
			if err != nil {
				fmt.Printf("[LOOP-CHECK]-%s\n", fmt.Sprintf("Recheck address utxo list err:%s", err.Error()))
				return
			}

			if utxoResp.UtxoList != nil && len(utxoResp.UtxoList) != 0 {
				utxoList = append(utxoList, utxoResp.UtxoList...)
				has := false
				for _, u := range utxoResp.UtxoList {
					if u.TxId == inscriptionTxId {
						has = true
						break
					}
				}
				if has {
					break
				}
			} else {
				break
			}
			time.Sleep(1 * time.Second)
		}

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
			major.Println(fmt.Sprintf("[LOOP-CHECK][%s]update for ask in orderId:%s, finishERR, success", v.Tick, v.OrderId))
			order_brc20_service.UpdateMarketPrice(v.Net, v.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(v.Tick)))
			order_brc20_service.UpdateMarketPriceV2(v.Net, v.Tick, fmt.Sprintf("%s-BTC", strings.ToUpper(v.Tick)))
		}
	}
	major.Println(fmt.Sprintf("[LOOP-CHECK]check success, time:%d", tool.MakeTimestamp()-t))
}
