package tool

import (
	"time"
)

const hourTime  = 3600 * 1000


func MakeTimestamp() int64 {
	//time.Now().UnixMilli()
	return time.Now().UnixNano() / int64(time.Millisecond)
}
