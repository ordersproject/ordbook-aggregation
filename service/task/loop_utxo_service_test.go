package task

import (
	"fmt"
	"testing"
)

func Test_compareCountForPerAmount(t *testing.T) {
	var (
		count1w  int64 = 11
		count5w  int64 = 3
		count10w int64 = 1
	)
	min, perAmount := compareCountForPerAmount(count1w, count5w, count10w)
	fmt.Println(min)
	fmt.Println(perAmount)
}
