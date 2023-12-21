package config

import (
	"fmt"
	"github.com/spf13/viper"
	"ordbook-aggregation/conf"
)

var (
	Port         = ""
	HiroDomain   = ""
	OklinkDomain = ""
	OklinkKey    = ""
	MempoolSpace = ""
	UnisatDomain = ""
	OwnDomain    = ""

	WsPort            = ""
	RedisEndpoint     = ""
	RedisPassword     = ""
	RedisDbUtxo   int = 1

	RpcUrlMainnet      = ""
	RpcUsernameMainnet = ""
	RpcPasswordMainnet = ""
	RpcUrlTestnet      = ""
	RpcUsernameTestnet = ""
	RpcPasswordTestnet = ""

	PlatformMainnetPrivateKeySendBrc20                                   = ""
	PlatformMainnetAddressSendBrc20                                      = "" //address for send brc20
	PlatformMainnetPrivateKeySendBrc20ForAsk                             = ""
	PlatformMainnetAddressSendBrc20ForAsk                                = "" //address for send brc20 and ask
	PlatformMainnetPrivateKeyReceiveValueForAsk                          = ""
	PlatformMainnetAddressReceiveValueForAsk                             = "" //address for receive value ask
	PlatformMainnetPrivateKeyReceiveBrc20                                = ""
	PlatformMainnetAddressReceiveBrc20                                   = "" //address for receive brc20
	PlatformMainnetPrivateKeyReceiveBidValue                             = ""
	PlatformMainnetAddressReceiveBidValue                                = "" // address for receive bid value
	PlatformMainnetPrivateKeyReceiveBidValueToX                          = ""
	PlatformMainnetAddressReceiveBidValueToX                             = "" // address for receive bid value to x
	PlatformMainnetPrivateKeyReceiveBidValueToReturn                     = ""
	PlatformMainnetAddressReceiveBidValueToReturn                        = "" // address for receive bid value to return
	PlatformMainnetPrivateKeyReceiveDummyValue                           = ""
	PlatformMainnetAddressReceiveDummyValue                              = "" // address for receive dummy 1200 value
	PlatformMainnetPrivateKeyReceiveFee                                  = ""
	PlatformMainnetAddressReceiveFee                                     = "" // address for receive fee
	PlatformMainnetPrivateKeyReceiveValueForPoolBtc                      = ""
	PlatformMainnetAddressReceiveValueForPoolBtc                         = "" // address for receive pool btc
	PlatformMainnetFeeRate                                      int64    = 0
	PlatformMainnetPrivateKeyMultiSig                           string   = ""
	PlatformMainnetPublicKeyMultiSig                            string   = "" // publicKey for multi sig
	PlatformMainnetPrivateKeyMultiSigBtc                        string   = ""
	PlatformMainnetPublicKeyMultiSigBtc                         string   = "" // publicKey for multi sig in btc
	PlatformMainnetPrivateKeyInscriptionMultiSig                string   = ""
	PlatformMainnetAddressInscriptionMultiSig                   string   = "" // address for inscription brc20 transfer
	PlatformMainnetPrivateKeyInscriptionMultiSigForReceiveValue string   = ""
	PlatformMainnetAddressInscriptionMultiSigForReceiveValue    string   = "" // address for receive value to inscription
	PlatformMainnetPrivateKeyRewardBrc20                        string   = ""
	PlatformMainnetAddressRewardBrc20                           string   = "" // address for reward brc20 transfer
	PlatformMainnetPrivateKeyRewardBrc20FeeUtxos                string   = ""
	PlatformMainnetAddressRewardBrc20FeeUtxos                   string   = "" // address for reward brc20 transfer fee utxos
	PlatformMainnetPrivateKeyRepurchaseReceiveBrc20             string   = ""
	PlatformMainnetAddressRepurchaseReceiveBrc20                string   = "" // address for repurchase receive brc20
	PlatformMainnetPrivateKeyDummy                              string   = ""
	PlatformMainnetAddressDummy                                 string   = "" // address for dummy
	PlatformMainnetPrivateKeyDummyAsk                           string   = ""
	PlatformMainnetAddressDummyAsk                              string   = "" // address for dummy ask
	PlatformMainnetPrivateKeyLp                                 string   = ""
	PlatformMainnetAddressLp                                    string   = "" // address for lp
	PlatformMainnetCirculationExceptAddresses                   []string = []string{}

	PlatformTestnetPrivateKeySendBrc20                                   = ""
	PlatformTestnetAddressSendBrc20                                      = "" //address for send brc20
	PlatformTestnetPrivateKeySendBrc20ForAsk                             = ""
	PlatformTestnetAddressSendBrc20ForAsk                                = "" //address for send brc20 and ask
	PlatformTestnetPrivateKeyReceiveValueForAsk                          = ""
	PlatformTestnetAddressReceiveValueForAsk                             = "" //address for receive value ask
	PlatformTestnetPrivateKeyReceiveBrc20                                = ""
	PlatformTestnetAddressReceiveBrc20                                   = "" //address for receive brc20
	PlatformTestnetPrivateKeyReceiveBidValue                             = ""
	PlatformTestnetAddressReceiveBidValue                                = "" // address for receive bid value
	PlatformTestnetPrivateKeyReceiveBidValueToX                          = ""
	PlatformTestnetAddressReceiveBidValueToX                             = "" // address for receive bid value to x
	PlatformTestnetPrivateKeyReceiveBidValueToReturn                     = ""
	PlatformTestnetAddressReceiveBidValueToReturn                        = "" // address for receive bid value to return
	PlatformTestnetPrivateKeyReceiveDummyValue                           = ""
	PlatformTestnetAddressReceiveDummyValue                              = "" // address for receive dummy 1200 value
	PlatformTestnetPrivateKeyReceiveFee                                  = ""
	PlatformTestnetAddressReceiveFee                                     = "" // address for receive fee
	PlatformTestnetFeeRate                                      int64    = 0
	PlatformTestnetPrivateKeyReceiveValueForPoolBtc                      = ""
	PlatformTestnetAddressReceiveValueForPoolBtc                         = "" // address for receive pool btc
	PlatformTestnetPrivateKeyMultiSig                           string   = ""
	PlatformTestnetPublicKeyMultiSig                            string   = "" // publicKey for multi sig
	PlatformTestnetPrivateKeyMultiSigBtc                        string   = ""
	PlatformTestnetPublicKeyMultiSigBtc                         string   = "" // publicKey for multi sig in btc
	PlatformTestnetPrivateKeyInscriptionMultiSig                string   = ""
	PlatformTestnetAddressInscriptionMultiSig                   string   = "" // address for inscription brc20 transfer
	PlatformTestnetPrivateKeyInscriptionMultiSigForReceiveValue string   = ""
	PlatformTestnetAddressInscriptionMultiSigForReceiveValue    string   = "" // address for receive value to inscription
	PlatformTestnetPrivateKeyRewardBrc20                        string   = ""
	PlatformTestnetAddressRewardBrc20                           string   = "" // address for reward brc20 transfer
	PlatformTestnetPrivateKeyRewardBrc20FeeUtxos                string   = ""
	PlatformTestnetAddressRewardBrc20FeeUtxos                   string   = "" // address for reward brc20 transfer fee utxos
	PlatformTestnetPrivateKeyRepurchaseReceiveBrc20             string   = ""
	PlatformTestnetAddressRepurchaseReceiveBrc20                string   = "" // address for repurchase receive brc20
	PlatformTestnetPrivateKeyDummy                              string   = ""
	PlatformTestnetAddressDummy                                 string   = "" // address for dummy
	PlatformTestnetPrivateKeyDummyAsk                           string   = ""
	PlatformTestnetAddressDummyAsk                              string   = "" // address for dummy ask
	PlatformTestnetPrivateKeyLp                                 string   = ""
	PlatformTestnetAddressLp                                    string   = "" // address for lp
	PlatformTestnetCirculationExceptAddresses                   []string = []string{}

	TestnetFakePriKey         string = ""
	TestnetFakeTaprootAddress string = ""

	PlatformRewardLinearReleaseMonths     int64  = 0
	PlatformRewardDayBase                 int64  = 0
	PlatformRewardExtraRewardDuration     int64  = 0
	PlatformRewardDiminishingDays         int64  = 0
	PlatformRewardDiminishingPeriod       int64  = 0
	PlatformRewardDiminishing1            int64  = 0
	PlatformRewardDiminishing2            int64  = 0
	PlatformRewardDiminishing3            int64  = 0
	PlatformRewardExtraDurationRewardRate int64  = 0
	PlatformRewardCalStartTime            int64  = 0
	PlatformRewardCalStartBlock           int64  = 0
	PlatformRewardCalEndBlock             int64  = 0
	PlatformRewardCalCycleBlock           int64  = 0
	PlatformRewardTick                    string = ""
	PlatformRewardDecreasingCycleTime     int64  = 0
	PlatformRewardDecreasingCycleBlock    int64  = 0

	NotificationTitlePoolUsed    string = ""
	NotificationTitleBidInvalid  string = ""
	NotificationTitleOrderFinish string = ""

	NotificationDescPoolUsed    string = ""
	NotificationDescBidInvalid  string = ""
	NotificationDescOrderFinish string = ""
)

// Event
var (
	EventOneStartTime                          int64    = 0
	EventOneStartBlock                         int64    = 0
	EventOneEndTime                            int64    = 0
	EventOneEndBlock                           int64    = 0
	EventOneRewardTick                         string   = ""
	EventOneExtraRewardAmount                  int64    = 0
	EventOneExtraRewardLpUnusedRate            int64    = 0
	EventOneExtraRewardLpUnusedDuration        int64    = 0
	EventOneExtraRewardLpUsedRate              int64    = 0
	EventOneExtraRewardBidRate                 int64    = 0
	EventOneExtraRewardExcludedOfficialAddress []string = []string{}
	EventOneRewardCalCycleBlock                int64    = 0
	EventPlatformPrivateKeyRewardBrc20         string   = ""
	EventPlatformAddressRewardBrc20            string   = "" // address for reward brc20 transfer
)

func InitConfig() {
	viper.SetConfigFile(conf.GetYaml())
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	Port = viper.GetString("port")
	HiroDomain = viper.GetString("hiro_domain")
	OklinkDomain, OklinkKey = viper.GetString("oklink.domain"), viper.GetString("oklink.key")
	UnisatDomain = viper.GetString("unisat.domain")
	OwnDomain = viper.GetString("own.domain")
	PlatformTestnetPrivateKeySendBrc20, PlatformTestnetAddressSendBrc20,
		PlatformTestnetPrivateKeySendBrc20ForAsk, PlatformTestnetAddressSendBrc20ForAsk,
		PlatformTestnetPrivateKeyReceiveValueForAsk, PlatformTestnetAddressReceiveValueForAsk,
		PlatformTestnetPrivateKeyReceiveBrc20, PlatformTestnetAddressReceiveBrc20,
		PlatformTestnetPrivateKeyReceiveBidValue, PlatformTestnetAddressReceiveBidValue,
		PlatformTestnetPrivateKeyReceiveBidValueToX, PlatformTestnetAddressReceiveBidValueToX,
		PlatformTestnetPrivateKeyReceiveBidValueToReturn, PlatformTestnetAddressReceiveBidValueToReturn,
		PlatformTestnetPrivateKeyReceiveDummyValue, PlatformTestnetAddressReceiveDummyValue,
		PlatformTestnetPrivateKeyReceiveFee, PlatformTestnetAddressReceiveFee,
		PlatformTestnetPrivateKeyReceiveValueForPoolBtc, PlatformTestnetAddressReceiveValueForPoolBtc,
		PlatformTestnetPrivateKeyMultiSig, PlatformTestnetPublicKeyMultiSig,
		PlatformTestnetPrivateKeyMultiSigBtc, PlatformTestnetPublicKeyMultiSigBtc,
		PlatformTestnetPrivateKeyInscriptionMultiSig, PlatformTestnetAddressInscriptionMultiSig,
		PlatformTestnetPrivateKeyInscriptionMultiSigForReceiveValue, PlatformTestnetAddressInscriptionMultiSigForReceiveValue,
		PlatformTestnetPrivateKeyRewardBrc20, PlatformTestnetAddressRewardBrc20,
		PlatformTestnetPrivateKeyRewardBrc20FeeUtxos, PlatformTestnetAddressRewardBrc20FeeUtxos,
		PlatformTestnetPrivateKeyRepurchaseReceiveBrc20, PlatformTestnetAddressRepurchaseReceiveBrc20,
		PlatformTestnetPrivateKeyDummy, PlatformTestnetAddressDummy,
		PlatformTestnetPrivateKeyDummyAsk, PlatformTestnetAddressDummyAsk,
		PlatformTestnetPrivateKeyLp, PlatformTestnetAddressLp =
		viper.GetString("platform.testnet.private_key_send_brc20"), viper.GetString("platform.testnet.address_send_brc20"),
		viper.GetString("platform.testnet.private_key_send_brc20_for_ask"), viper.GetString("platform.testnet.address_send_brc20_for_ask"),
		viper.GetString("platform.testnet.private_key_receive_value_for_ask"), viper.GetString("platform.testnet.address_receive_value_for_ask"),
		viper.GetString("platform.testnet.private_key_receive_brc20"), viper.GetString("platform.testnet.address_receive_brc20"),
		viper.GetString("platform.testnet.private_key_receive_bid_value"), viper.GetString("platform.testnet.address_receive_bid_value"),
		viper.GetString("platform.testnet.private_key_receive_bid_value_to_x"), viper.GetString("platform.testnet.address_receive_bid_value_to_x"),
		viper.GetString("platform.testnet.private_key_receive_bid_value_to_return"), viper.GetString("platform.testnet.address_receive_bid_value_to_return"),
		viper.GetString("platform.testnet.private_key_receive_dummy_value"), viper.GetString("platform.testnet.address_receive_dummy_value"),
		viper.GetString("platform.testnet.private_key_receive_fee"), viper.GetString("platform.testnet.address_receive_fee"),
		viper.GetString("platform.testnet.private_key_receive_value_for_pool_btc"), viper.GetString("platform.testnet.address_receive_value_for_pool_btc"),
		viper.GetString("platform.testnet.private_key_platform_multi_sig"), viper.GetString("platform.testnet.public_key_platform_multi_sig"),
		viper.GetString("platform.testnet.private_key_platform_multi_sig_btc"), viper.GetString("platform.testnet.public_key_platform_multi_sig_btc"),
		viper.GetString("platform.testnet.private_key_inscription_multi_sig"), viper.GetString("platform.testnet.address_inscription_multi_sig"),
		viper.GetString("platform.testnet.private_key_inscription_multi_sig_for_receive_value"), viper.GetString("platform.testnet.address_inscription_multi_sig_for_receive_value"),
		viper.GetString("platform.testnet.private_key_reward_brc20"), viper.GetString("platform.testnet.address_reward_brc20"),
		viper.GetString("platform.testnet.private_key_reward_brc20_fee_utxos"), viper.GetString("platform.testnet.address_reward_brc20_fee_utxos"),
		viper.GetString("platform.testnet.private_key_repurchase_receive_brc20"), viper.GetString("platform.testnet.address_repurchase_receive_brc20"),
		viper.GetString("platform.testnet.private_key_dummy"), viper.GetString("platform.testnet.address_dummy"),
		viper.GetString("platform.testnet.private_key_dummy_ask"), viper.GetString("platform.testnet.address_dummy_ask"),
		viper.GetString("platform.testnet.private_key_lp"), viper.GetString("platform.testnet.address_lp")
	PlatformMainnetPrivateKeySendBrc20, PlatformMainnetAddressSendBrc20,
		PlatformMainnetPrivateKeySendBrc20ForAsk, PlatformMainnetAddressSendBrc20ForAsk,
		PlatformMainnetPrivateKeyReceiveValueForAsk, PlatformMainnetAddressReceiveValueForAsk,
		PlatformMainnetPrivateKeyReceiveBrc20, PlatformMainnetAddressReceiveBrc20,
		PlatformMainnetPrivateKeyReceiveBidValue, PlatformMainnetAddressReceiveBidValue,
		PlatformMainnetPrivateKeyReceiveBidValueToX, PlatformMainnetAddressReceiveBidValueToX,
		PlatformMainnetPrivateKeyReceiveBidValueToReturn, PlatformMainnetAddressReceiveBidValueToReturn,
		PlatformMainnetPrivateKeyReceiveDummyValue, PlatformMainnetAddressReceiveDummyValue,
		PlatformMainnetPrivateKeyReceiveFee, PlatformMainnetAddressReceiveFee,
		PlatformMainnetPrivateKeyReceiveValueForPoolBtc, PlatformMainnetAddressReceiveValueForPoolBtc,
		PlatformMainnetPrivateKeyMultiSig, PlatformMainnetPublicKeyMultiSig,
		PlatformMainnetPrivateKeyMultiSigBtc, PlatformMainnetPublicKeyMultiSigBtc,
		PlatformMainnetPrivateKeyInscriptionMultiSig, PlatformMainnetAddressInscriptionMultiSig,
		PlatformMainnetPrivateKeyInscriptionMultiSigForReceiveValue, PlatformMainnetAddressInscriptionMultiSigForReceiveValue,
		PlatformMainnetPrivateKeyRewardBrc20, PlatformMainnetAddressRewardBrc20,
		PlatformMainnetPrivateKeyRewardBrc20FeeUtxos, PlatformMainnetAddressRewardBrc20FeeUtxos,
		PlatformMainnetPrivateKeyRepurchaseReceiveBrc20, PlatformMainnetAddressRepurchaseReceiveBrc20,
		PlatformMainnetPrivateKeyDummy, PlatformMainnetAddressDummy,
		PlatformMainnetPrivateKeyDummyAsk, PlatformMainnetAddressDummyAsk,
		PlatformMainnetPrivateKeyLp, PlatformMainnetAddressLp =
		viper.GetString("platform.mainnet.private_key_Send_brc20"), viper.GetString("platform.mainnet.address_Send_brc20"),
		viper.GetString("platform.mainnet.private_key_send_brc20_for_ask"), viper.GetString("platform.mainnet.address_Send_brc20_for_ask"),
		viper.GetString("platform.mainnet.private_key_receive_value_for_ask"), viper.GetString("platform.mainnet.address_receive_value_for_ask"),
		viper.GetString("platform.mainnet.private_key_receive_brc20"), viper.GetString("platform.mainnet.address_receive_brc20"),
		viper.GetString("platform.mainnet.private_key_receive_bid_value"), viper.GetString("platform.mainnet.address_receive_bid_value"),
		viper.GetString("platform.mainnet.private_key_receive_bid_value_to_x"), viper.GetString("platform.mainnet.address_receive_bid_value_to_x"),
		viper.GetString("platform.mainnet.private_key_receive_bid_value_to_return"), viper.GetString("platform.mainnet.address_receive_bid_value_to_return"),
		viper.GetString("platform.mainnet.private_key_receive_dummy_value"), viper.GetString("platform.mainnet.address_receive_dummy_value"),
		viper.GetString("platform.mainnet.private_key_receive_fee"), viper.GetString("platform.mainnet.address_receive_fee"),
		viper.GetString("platform.mainnet.private_key_receive_value_for_pool_btc"), viper.GetString("platform.mainnet.address_receive_value_for_pool_btc"),
		viper.GetString("platform.mainnet.private_key_platform_multi_sig"), viper.GetString("platform.mainnet.public_key_platform_multi_sig"),
		viper.GetString("platform.mainnet.private_key_platform_multi_sig_btc"), viper.GetString("platform.mainnet.public_key_platform_multi_sig_btc"),
		viper.GetString("platform.mainnet.private_key_inscription_multi_sig"), viper.GetString("platform.mainnet.address_inscription_multi_sig"),
		viper.GetString("platform.mainnet.private_key_inscription_multi_sig_for_receive_value"), viper.GetString("platform.mainnet.address_inscription_multi_sig_for_receive_value"),
		viper.GetString("platform.mainnet.private_key_reward_brc20"), viper.GetString("platform.mainnet.address_reward_brc20"),
		viper.GetString("platform.mainnet.private_key_reward_brc20_fee_utxos"), viper.GetString("platform.mainnet.address_reward_brc20_fee_utxos"),
		viper.GetString("platform.mainnet.private_key_repurchase_receive_brc20"), viper.GetString("platform.mainnet.address_repurchase_receive_brc20"),
		viper.GetString("platform.mainnet.private_key_dummy"), viper.GetString("platform.mainnet.address_dummy"),
		viper.GetString("platform.mainnet.private_key_dummy_ask"), viper.GetString("platform.mainnet.address_dummy_ask"),
		viper.GetString("platform.mainnet.private_key_lp"), viper.GetString("platform.mainnet.address_lp")
	PlatformMainnetFeeRate, PlatformTestnetFeeRate = viper.GetInt64("platform.mainnet.fee_rate"), viper.GetInt64("platform.testnet.fee_rate")
	PlatformMainnetCirculationExceptAddresses, PlatformTestnetCirculationExceptAddresses =
		viper.GetStringSlice("platform.mainnet.circulation_except_addresses"), viper.GetStringSlice("platform.testnet.circulation_except_addresses")

	MempoolSpace = viper.GetString("mempool_space.domain")
	WsPort = viper.GetString("ws.port")
	RedisEndpoint, RedisPassword = viper.GetString("redis.endpoint"), viper.GetString("redis.password")
	RedisDbUtxo = viper.GetInt("redis.db_utxo")

	RpcUrlTestnet, RpcUsernameTestnet, RpcPasswordTestnet = viper.GetString("node.testnet.url"), viper.GetString("node.testnet.username"), viper.GetString("node.testnet.password")
	RpcUrlMainnet, RpcUsernameMainnet, RpcPasswordMainnet = viper.GetString("node.mainnet.url"), viper.GetString("node.mainnet.username"), viper.GetString("node.mainnet.password")

	TestnetFakePriKey = viper.GetString("testnet.pri_key")
	TestnetFakeTaprootAddress = viper.GetString("testnet.taproot_address")

	PlatformRewardLinearReleaseMonths = viper.GetInt64("platform_service_reward.linear_release_months")
	PlatformRewardDayBase = viper.GetInt64("platform_service_reward.day_base")
	PlatformRewardExtraRewardDuration = viper.GetInt64("platform_service_reward.extra_reward_duration")
	PlatformRewardDiminishingDays = viper.GetInt64("platform_service_reward.diminishing_reward_day")
	PlatformRewardDiminishingPeriod = viper.GetInt64("platform_service_reward.diminishing_period")
	PlatformRewardDiminishing1 = viper.GetInt64("platform_service_reward.diminishing_1")
	PlatformRewardDiminishing2 = viper.GetInt64("platform_service_reward.diminishing_2")
	PlatformRewardDiminishing3 = viper.GetInt64("platform_service_reward.diminishing_3")
	PlatformRewardExtraDurationRewardRate = viper.GetInt64("platform_service_reward.extra_reward_duration_rate")
	PlatformRewardCalStartTime = viper.GetInt64("platform_service_reward.cal_start_time")
	PlatformRewardCalStartBlock = viper.GetInt64("platform_service_reward.cal_start_block")
	PlatformRewardCalEndBlock = viper.GetInt64("platform_service_reward.cal_end_block")
	PlatformRewardCalCycleBlock = viper.GetInt64("platform_service_reward.cal_cycle_block")
	PlatformRewardTick = viper.GetString("platform_service_reward.reward_tick")
	PlatformRewardDecreasingCycleTime = viper.GetInt64("platform_service_reward.decreasing_cycle_time")
	PlatformRewardDecreasingCycleBlock = viper.GetInt64("platform_service_reward.decreasing_cycle_block")
	fmt.Printf("PlatformRewardCalStartTime-[%d]\n", PlatformRewardCalStartTime)
	fmt.Printf("PlatformRewardCalStartBlock-[%d]\n", PlatformRewardCalStartBlock)
	fmt.Printf("decreasing_cycle_time-[%d]\n", PlatformRewardDecreasingCycleTime)
	fmt.Printf("decreasing_cycle_block-[%d]\n", PlatformRewardDecreasingCycleBlock)

	NotificationTitlePoolUsed = viper.GetString("notification.title.pool_used")
	NotificationTitleBidInvalid = viper.GetString("notification.title.bid_invalid")
	NotificationTitleOrderFinish = viper.GetString("notification.title.order_finish")
	NotificationDescPoolUsed = viper.GetString("notification.description.pool_used")
	NotificationDescBidInvalid = viper.GetString("notification.description.bid_invalid")
	NotificationDescOrderFinish = viper.GetString("notification.description.order_finish")

	//Event
	EventOneStartTime = viper.GetInt64("event.one.start_time")
	EventOneStartBlock = viper.GetInt64("event.one.start_block")
	EventOneEndTime = viper.GetInt64("event.one.end_time")
	EventOneEndBlock = viper.GetInt64("event.one.end_block")
	EventOneRewardTick = viper.GetString("event.one.reward_tick")
	EventOneExtraRewardAmount = viper.GetInt64("event.one.extra_reward_amount")
	EventOneExtraRewardLpUnusedRate = viper.GetInt64("event.one.extra_reward_lp_unused_rate")
	EventOneExtraRewardLpUnusedDuration = viper.GetInt64("event.one.extra_reward_lp_unused_duration")
	EventOneExtraRewardLpUsedRate = viper.GetInt64("event.one.extra_reward_lp_used_rate")
	EventOneExtraRewardBidRate = viper.GetInt64("event.one.extra_reward_bid_rate")
	EventOneExtraRewardExcludedOfficialAddress = viper.GetStringSlice("event.one.official_excluded_address")
	EventOneRewardCalCycleBlock = viper.GetInt64("event.one.cal_cycle_block")
	EventPlatformPrivateKeyRewardBrc20, EventPlatformAddressRewardBrc20 = viper.GetString("event.one.private_key_reward_brc20"), viper.GetString("event.one.address_reward_brc20")
}
