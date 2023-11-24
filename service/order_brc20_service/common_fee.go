package order_brc20_service

var (
	sendModulus                 int64 = 340
	inscriptionModulus          int64 = 378
	constant                    int64 = 30
	multiSigInscriptionConstant int64 = 1030
)

// generate bid taker fee
// networkFeeRate: sat/byte
// releaseInscriptionFee: inscribe fee for release
// rewardInscriptionFee: inscribe fee for reward
// rewardSendFee: send fee for reward
func GenerateBidTakerFee(networkFeeRate int64) (int64, int64, int64) {
	var (
		releaseInscriptionFee = int64(0)
		rewardInscriptionFee  = int64(0)
		rewardSendFee         = int64(0)
	)
	releaseInscriptionFee = inscriptionModulus*networkFeeRate + multiSigInscriptionConstant
	rewardInscriptionFee = inscriptionModulus*networkFeeRate + constant
	rewardSendFee = sendModulus*networkFeeRate + constant
	return releaseInscriptionFee, rewardInscriptionFee, rewardSendFee
}
