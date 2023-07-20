package order_brc20_service

import (
	"fmt"
	"ordbook-aggregation/major"
	"ordbook-aggregation/model"
	"ordbook-aggregation/service/mongo_service"
	"ordbook-aggregation/tool"
	"strings"
)

func InitCommon() {
	initPoolPair()
}

func initPoolPair() {
	var (
		netList []string = []string{
			"livenet",
			"testnet",
		}
		tickList []string = []string{
			"rdex",
			"ordi",
		}
	)

	for _, net := range netList {
		for _, tick := range tickList {
			pair := fmt.Sprintf("%s_BTC", strings.ToUpper(tick))
			entity, _ := mongo_service.FindPoolInfoModelByPair(net, pair)
			if entity != nil && entity.Id != 0 {
				continue
			}
			entity = &model.PoolInfoModel{
				Net:            net,
				Tick:           tick,
				Pair:           pair,
				CoinAmount:     0,
				CoinDecimalNum: 18,
				Amount:         0,
				DecimalNum:     8,
				Timestamp:      tool.MakeTimestamp(),
			}
			_, err := mongo_service.SetPoolInfoModel(entity)
			if err != nil {
				major.Println(fmt.Sprintf("SetPoolInfoModel err:%s", err.Error()))
				continue
			}
		}
	}
	fmt.Printf("[INIT]Init pool pair info completed")
}
