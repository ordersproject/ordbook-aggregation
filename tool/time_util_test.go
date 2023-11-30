package tool

import (
	"fmt"
	"testing"
)

func TestGetToday0Time(t *testing.T) {
	fmt.Println(GetToday0Time())
	fmt.Println(GetToday24Time())
}

func TestGetToday0And24Time(t *testing.T) {
	start, end := GetToday0And24Time()
	fmt.Printf("start:%d, end:%d\n", start, end)
	fmt.Printf("start:%s, end:%s\n", MakeDate(start), MakeDate(end))
}

func TestGetYesterday0And24Time(t *testing.T) {
	start, end := GetYesterday0And24Time()
	fmt.Printf("start:%d, end:%d\n", start, end)
	fmt.Printf("start:%s, end:%s\n", MakeDate(start), MakeDate(end))
}
