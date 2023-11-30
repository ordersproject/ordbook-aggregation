package order_brc20_service

import (
	"fmt"
	"ordbook-aggregation/config"
	"ordbook-aggregation/tool"
	"testing"
)

func TestGetEventNowCalStartTimeAndEndTime(t *testing.T) {
	config.InitConfig()
	calStart, calEnd, cal := GetEventNowCalStartTimeAndEndTime()
	fmt.Printf("calStart:%d, calEnd:%d, cal:%d\n", config.EventOneStartTime, config.EventOneEndTime, cal)
	fmt.Printf("calStart:%s, calEnd:%s, cal:%s\n", tool.MakeDate(config.EventOneStartTime), tool.MakeDate(config.EventOneEndTime), tool.MakeDate(cal))

	fmt.Printf("calStart:%d, calEnd:%d, cal:%d\n", calStart, calEnd, cal)
	fmt.Printf("calStart:%s, calEnd:%s, cal:%s\n", tool.MakeDate(calStart), tool.MakeDate(calEnd), tool.MakeDate(cal))
}
