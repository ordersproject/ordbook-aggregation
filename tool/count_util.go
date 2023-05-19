package tool

import (
	"fmt"
	"strconv"
)

func ChangeByte(total int64) string {
	totalReal := total
	totalStr := "0.00"
	d := 0
	unit := "B"
	for {
		if total >= 1024 {
			totalReal = total/1024
			d++
		}else {
			if totalReal == total {
				totalStr = strconv.FormatFloat(float64(total), 'f', 2, 64)
				break
			}
		}
		if totalReal >= 1024 {
			total = totalReal
			continue
		}else {
			totalStr = strconv.FormatFloat(float64(total)/1024, 'f', 2, 64)
			break
		}
	}
	switch d {
	case 0:
		unit = "B"
		break
	case 1:
		unit = "KB"
		break
	case 2:
		unit = "MB"
		break
	case 3:
		unit = "GB"
		break
	case 4:
		unit = "TB"
		break
	case 5:
		unit = "PB"
		break
	case 6:
		unit = "EB"
		break
	case 7:
		unit = "ZB"
		break
	default:
		unit = "Too big"
	}
	return fmt.Sprintf("%s %s", totalStr, unit)
}
