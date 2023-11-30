package task

import (
	"fmt"
	"testing"
)

func Test_getStartBlockAndEndBlockByBigBlock(t *testing.T) {
	start, end := getStartBlockAndEndBlockByBigBlock(19)
	fmt.Println(start, end)
}
