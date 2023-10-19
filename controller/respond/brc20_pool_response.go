package respond

import "ordbook-aggregation/model"

type PoolResponse struct {
	Total   int64            `json:"total,omitempty"`
	Results []*PoolBrc20Item `json:"results,omitempty"`
	Flag    int64            `json:"flag,omitempty"`
}

type PoolBrc20Item struct {
	Net                                     string          `json:"net,omitempty"`            //Net env
	OrderId                                 string          `json:"orderId,omitempty"`        //Order ID
	Tick                                    string          `json:"tick,omitempty"`           //Brc20 symbol
	Pair                                    string          `json:"pair,omitempty"`           //Brc20 pair
	CoinAmount                              uint64          `json:"coinAmount,omitempty"`     //tick
	CoinDecimalNum                          int             `json:"coinDecimalNum,omitempty"` //tick
	CoinAddress                             string          `json:"coinAddress,omitempty"`    //tick
	Amount                                  uint64          `json:"amount,omitempty"`         //
	DecimalNum                              int             `json:"decimalNum,omitempty"`     //
	PoolType                                model.PoolType  `json:"poolType,omitempty"`       //pool type：1-tick,2-btc
	PoolState                               model.PoolState `json:"poolState,omitempty"`      //pool state：1-add,2-remove,3-used,4-claim
	Address                                 string          `json:"address,omitempty"`        //address
	InscriptionId                           string          `json:"inscriptionId,omitempty"`  //InscriptionId
	CoinPsbtRaw                             string          `json:"coinPsbtRaw,omitempty"`    //coin PSBT Raw
	PsbtRaw                                 string          `json:"psbtRaw,omitempty"`        //PSBT Raw
	UtxoId                                  string          `json:"utxoId"`                   //UtxoId
	MultiSigScriptAddress                   string          `json:"multiSigScriptAddress"`
	DealInscriptionId                       string          `json:"dealInscriptionId"` //InscriptionId
	DealInscriptionTx                       string          `json:"dealInscriptionTx"`
	DealInscriptionTxIndex                  int64           `json:"dealInscriptionTxIndex"`
	DealInscriptionTxOutValue               int64           `json:"dealInscriptionTxOutValue"`
	DealInscriptionTime                     int64           `json:"dealInscriptionTime"`
	MultiSigScriptAddressTickAvailableState int64           `json:"multiSigScriptAddressTickAvailableState"` //0-no, 1-available
	Timestamp                               int64           `json:"timestamp"`                               //Create time
	RewardCoinAmount                        int64           `json:"rewardCoinAmount,omitempty"`
	ReleaseTx                               string          `json:"releaseTx"`
	ReleaseTime                             int64           `json:"releaseTime"`
	ReleaseTxBlock                          int64           `json:"releaseTxBlock"`
	DealTime                                int64           `json:"dealTime"`
	Decreasing                              int64           `json:"decreasing"`
}

type PoolInfoResponse struct {
	Total   int64           `json:"total,omitempty"`
	Results []*PoolInfoItem `json:"results,omitempty"`
	Flag    int64           `json:"flag,omitempty"`
}
type PoolInfoItem struct {
	Net            string `json:"net,omitempty"`  //Net env
	Tick           string `json:"tick,omitempty"` //Brc20 symbol
	Pair           string `json:"pair,omitempty"` //Brc20 pair
	CoinAmount     uint64 `json:"coinAmount"`     //
	CoinDecimalNum int    `json:"coinDecimalNum"` //omitempty
	Amount         uint64 `json:"amount"`         //Btc: sat
	DecimalNum     int    `json:"decimalNum"`     //Btc decimal
	OwnCoinAmount  uint64 `json:"ownCoinAmount,omitempty"`
	OwnAmount      uint64 `json:"ownAmount,omitempty"`
	OwnCount       uint64 `json:"ownCount,omitempty"`
}

type PoolKeyInfoResp struct {
	Net               string `json:"net,omitempty"`               //Net env
	PublicKey         string `json:"publicKey,omitempty"`         // key
	BtcPublicKey      string `json:"btcPublicKey,omitempty"`      // key for btc
	BtcReceiveAddress string `json:"btcReceiveAddress,omitempty"` // key
}

type PoolInscriptionResp struct {
	Net   string                 `json:"net,omitempty"`
	Tick  string                 `json:"tick,omitempty"` //
	List  []*PoolInscriptionItem `json:"availableList,omitempty"`
	Total int64                  `json:"total,omitempty"`
}

type PoolInscriptionItem struct {
	InscriptionId     string `json:"inscriptionId,omitempty"`
	InscriptionNumber string `json:"inscriptionNumber,omitempty"`
	CoinAmount        string `json:"coinAmount,omitempty"`
}

type PoolBrc20ClaimResp struct {
	Net                 string `json:"net,omitempty"`           //Net env
	OrderId             string `json:"orderId,omitempty"`       //Order ID
	Tick                string `json:"tick,omitempty"`          //Brc20 symbol
	Fee                 uint64 `json:"fee,omitempty"`           //claim fee
	CoinAmount          uint64 `json:"coinAmount,omitempty"`    //Brc20 amount
	InscriptionId       string `json:"inscriptionId,omitempty"` //InscriptionId
	PsbtRaw             string `json:"psbtRaw"`                 //PSBT Raw
	CoinTransferPsbtRaw string `json:"coinTransferPsbtRaw"`     //coin transfer PSBT Raw
	CoinPsbtRaw         string `json:"coinPsbtRaw"`             //coin PSBT Raw
	RewardPsbtRaw       string `json:"rewardPsbtRaw"`           //reward PSBT Raw
	RewardCoinAmount    int64  `json:"rewardCoinAmount,omitempty"`
}

type PoolBrc20RewardResp struct {
	Net               string `json:"net,omitempty"`  //Net env
	Tick              string `json:"tick,omitempty"` //Brc20 symbol
	TotalRewardAmount uint64 `json:"totalRewardAmount"`
	//ClaimedOwnCoinAmount     uint64 `json:"claimedOwnCoinAmount"`
	//ClaimedOwnAmount         uint64 `json:"claimedOwnAmount"`
	//ClaimedOwnCount          uint64 `json:"claimedOwnCount"`
	HadClaimRewardAmount uint64 `json:"hadClaimRewardAmount"`
	//HadClaimRewardOrderCount uint64 `json:"HadClaimRewardOrderCount"`
	HasReleasePoolOrderCount int64 `json:"hasReleasePoolOrderCount"`
}

type PoolRewardOrderResponse struct {
	Total   int64                  `json:"total,omitempty"`
	Results []*PoolRewardOrderItem `json:"results,omitempty"`
	Flag    int64                  `json:"flag,omitempty"`
}

type PoolRewardOrderItem struct {
	Net              string            `json:"net,omitempty"`
	Tick             string            `json:"tick,omitempty"`
	OrderId          string            `json:"orderId,omitempty"`
	Pair             string            `json:"pair"`
	RewardCoinAmount int64             `json:"rewardCoinAmount,omitempty"`
	Address          string            `json:"address,omitempty"`
	RewardState      model.RewardState `json:"rewardState,omitempty"`
	Timestamp        int64             `json:"timestamp,omitempty"`
}
