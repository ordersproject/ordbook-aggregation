package common_service

import "strings"

func ChangeRealTick(tick string) string {
	switch tick {
	case "rdex", "oxbt", "grum", "vmpx", "lger", "sayc":
		tick = strings.ToUpper(tick)
	}
	return tick
}
