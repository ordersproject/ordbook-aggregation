package tool

import (
	"fmt"
	"strconv"
	"time"
)

const hourTime  = 3600 * 1000

func GetLocalBetween24Time() (int64, int64) {
	currentTime := MakeTimestamp()
	//if currentTime - GetLocalToday0Time() >= 8 * hourTime {
	//	return GetLocalToday0Time(), currentTime
	//}else {
	//	return GetLocalYesterday0Time(), currentTime
	//}

	return currentTime - (24 * hourTime), currentTime
}

func GetLocalBetween48Time() (int64, int64) {
	currentTime := MakeTimestamp()
	return currentTime - (2 * 24 * hourTime), currentTime
}

func GetLocalBetweenWeek() (int64, int64) {
	currentTime := MakeTimestamp()
	return currentTime - (7 * 24 * hourTime), currentTime
}

func GetLocalBetweenMonth() (int64, int64) {
	currentTime := MakeTimestamp()
	return currentTime - (30 * 24 * hourTime), currentTime
}

//获取今天0点0时0分的时间戳
func GetLocalToday0Time() int64 {
	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	return startTime.UnixNano() / 1e6
}

//获取今天23:59:59秒的时间戳
func GetLocalToday24Time() int64 {
	currentTime := time.Now()
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location())
	return endTime.UnixNano() / 1e6
}

//获取昨天0点0时0分的时间戳
func GetLocalYesterday0Time() int64 {
	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
	return (startTime.UnixNano() / 1e6) - (24 * hourTime)
}

//获取1分钟之前的时间
//获取1小时之前的时间
func GetAfterTime() int64 {
	currentTime := time.Now()
	m, _ := time.ParseDuration("1m")
	//m, _ := time.ParseDuration("1h")
	result := currentTime.Add(m)
	//fmt.Println(result.Format("2006/01/02 15:04:05"))
	return result.UnixNano() / 1e6
}

func ConsumeTime(time int64) string {
	sub := MakeTimestamp() - time
	//return strconv.FormatFloat(float64(sub)/1000, 'f', 2, 64)
	return strconv.FormatUint(uint64(sub), 10)
}

func GetAfterDay(day int64) int64 {
	currentTime := time.Now()
	dayStr := fmt.Sprintf("%dh", day * 24)
	m, _ := time.ParseDuration(dayStr)
	//m, _ := time.ParseDuration("1h")
	result := currentTime.Add(m)
	//fmt.Println(result.Format("2006/01/02 15:04:05"))
	return result.UnixNano() / 1e6
}

func MakeTimestamp() int64 {
	//time.Now().UnixMilli()
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func MakeTimeDate(timestamp int64) string {
	var (
		cst_sh, _  = time.LoadLocation("Asia/Shanghai") //上海
		timaDate string = ""
		format string = "2006-01-02 15:04:05"
	)
	timaDate =time.Unix(timestamp/1000, 0).In(cst_sh).Format(format)
	return timaDate
}