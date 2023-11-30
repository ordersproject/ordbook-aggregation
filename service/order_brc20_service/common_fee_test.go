package order_brc20_service

import "testing"

func TestGenerateBidTakerFee(t *testing.T) {
	networkFeeRate := int64(40)
	feeAmountForReleaseInscription, feeAmountForRewardInscription, feeAmountForRewardSend := GenerateBidTakerFee(networkFeeRate)
	t.Logf("\nfeeAmountForReleaseInscription: %d\nfeeAmountForRewardInscription: %d\nfeeAmountForRewardSend: %d", feeAmountForReleaseInscription, feeAmountForRewardInscription, feeAmountForRewardSend)

}
