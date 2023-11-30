package tool

func GetEndBlockByStartBlockAndCycleBlock(startBlock, cycleBlock int64) int64 {
	var (
		endBlock int64 = 0
	)
	endBlock = startBlock + cycleBlock - 1
	return endBlock
}
