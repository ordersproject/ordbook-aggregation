package task

import (
	"ordbook-aggregation/logger"
	"time"
)

func Run() {
	loopUtxoService()
}

func loopUtxoService() {
	go func() {
		logger.Logger.Infof(" \n")
		timeTickerChan := time.Tick(time.Minute * 10)
		for {
			logger.Logger.Infof("Check utxo receive \n")
			LoopCheckPlatformAddressForBidValue("livenet")
			<-timeTickerChan
		}
	}()

	go func() {
		logger.Logger.Infof(" \n")
		timeTickerChan := time.Tick(time.Minute * 30)
		for {
			logger.Logger.Infof("Check utxo receive \n")
			LoopCheckPlatformAddressForDummyValue("livenet")
			<-timeTickerChan
		}
	}()

	go func() {
		logger.Logger.Infof(" \n")
		timeTickerChan := time.Tick(time.Minute * 60)
		for {
			logger.Logger.Infof("Check ask receive value \n")
			LoopCheckAskReceiveValueChangeAsk()
			<-timeTickerChan
		}
	}()

	go func() {
		logger.Logger.Infof(" \n")
		timeTickerChan := time.Tick(time.Minute * 30)
		for {
			logger.Logger.Infof("Check ask state \n")
			LoopForCheckAsk()
			<-timeTickerChan
		}
	}()

}
