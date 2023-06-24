package tool

import (
	"fmt"
	"testing"
)

func TestGetToday0Time(t *testing.T) {
	fmt.Println(GetToday0Time())
	fmt.Println(GetToday24Time())
}