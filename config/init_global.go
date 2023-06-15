package config

import (
	"fmt"
	"github.com/spf13/viper"
	"ordbook-aggregation/conf"
)

var (
	Port = ""
	HiroDomain = ""
	OklinkDomain = ""
	OklinkKey = ""
	MempoolSpace = ""
	UnisatDomain = ""

	WsPort = ""
	RedisEndpoint = ""
	RedisPassword = ""
	RedisDbUtxo int = 1

	RpcUrlMainnet = ""
	RpcUsernameMainnet = ""
	RpcPasswordMainnet = ""
	RpcUrlTestnet = ""
	RpcUsernameTestnet = ""
	RpcPasswordTestnet = ""

	PlatformMainnetPrivateKeySendBrc20 = ""
	PlatformMainnetAddressSendBrc20 = ""//address for send brc20
	PlatformMainnetPrivateKeyReceiveBrc20 = ""
	PlatformMainnetAddressReceiveBrc20 = ""//address for receive brc20
	PlatformMainnetPrivateKeyReceiveBidValue = ""
	PlatformMainnetAddressReceiveBidValue = ""// address for receive bid value
	PlatformMainnetPrivateKeyReceiveDummyValue = ""
	PlatformMainnetAddressReceiveDummyValue = ""// address for receive dummy 1200 value
	PlatformMainnetPrivateKeyReceiveFee = ""
	PlatformMainnetAddressReceiveFee = ""// address for receive fee
	PlatformMainnetFeeRate int64 = 0

	PlatformTestnetPrivateKeySendBrc20 = ""
	PlatformTestnetAddressSendBrc20 = ""//address for send brc20
	PlatformTestnetPrivateKeyReceiveBrc20 = ""
	PlatformTestnetAddressReceiveBrc20 = ""//address for receive brc20
	PlatformTestnetPrivateKeyReceiveBidValue = ""
	PlatformTestnetAddressReceiveBidValue = ""// address for receive bid value
	PlatformTestnetPrivateKeyReceiveDummyValue = ""
	PlatformTestnetAddressReceiveDummyValue = ""// address for receive dummy 1200 value
	PlatformTestnetPrivateKeyReceiveFee = ""
	PlatformTestnetAddressReceiveFee = ""// address for receive fee
	PlatformTestnetFeeRate int64 = 0
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
	PlatformTestnetPrivateKeySendBrc20, PlatformTestnetAddressSendBrc20,
	PlatformTestnetPrivateKeyReceiveBrc20, PlatformTestnetAddressReceiveBrc20,
		PlatformTestnetPrivateKeyReceiveBidValue, PlatformTestnetAddressReceiveBidValue,
		PlatformTestnetPrivateKeyReceiveDummyValue, PlatformTestnetAddressReceiveDummyValue,
		PlatformTestnetPrivateKeyReceiveFee, PlatformTestnetAddressReceiveFee =
		viper.GetString("platform.testnet.private_key_send_brc20"), viper.GetString("platform.testnet.address_send_brc20"),
		viper.GetString("platform.testnet.private_key_receive_brc20"), viper.GetString("platform.testnet.address_receive_brc20"),
		viper.GetString("platform.testnet.private_key_receive_bid_value"), viper.GetString("platform.testnet.address_receive_bid_value"),
		viper.GetString("platform.testnet.private_key_receive_dummy_value"), viper.GetString("platform.testnet.address_receive_dummy_value"),
		viper.GetString("platform.testnet.private_key_receive_fee"), viper.GetString("platform.testnet.address_receive_fee")
	PlatformMainnetPrivateKeySendBrc20, PlatformMainnetAddressSendBrc20,
	PlatformMainnetPrivateKeyReceiveBrc20, PlatformMainnetAddressReceiveBrc20,
		PlatformMainnetPrivateKeyReceiveBidValue, PlatformMainnetAddressReceiveBidValue,
		PlatformMainnetPrivateKeyReceiveDummyValue, PlatformMainnetAddressReceiveDummyValue,
		PlatformMainnetPrivateKeyReceiveFee, PlatformMainnetAddressReceiveFee =
		viper.GetString("platform.mainnet.private_key_Send_brc20"), viper.GetString("platform.mainnet.address_Send_brc20"),
		viper.GetString("platform.mainnet.private_key_receive_brc20"), viper.GetString("platform.mainnet.address_receive_brc20"),
		viper.GetString("platform.mainnet.private_key_receive_bid_value"), viper.GetString("platform.mainnet.address_receive_bid_value"),
		viper.GetString("platform.mainnet.private_key_receive_dummy_value"), viper.GetString("platform.mainnet.address_receive_dummy_value"),
		viper.GetString("platform.mainnet.private_key_receive_fee"), viper.GetString("platform.mainnet.address_receive_fee")
	PlatformMainnetFeeRate, PlatformTestnetFeeRate = viper.GetInt64("platform.mainnet.fee_rate"), viper.GetInt64("platform.testnet.fee_rate")
		MempoolSpace = viper.GetString("mempool_space.domain")
	WsPort = viper.GetString("ws.port")
	RedisEndpoint, RedisPassword = viper.GetString("redis.endpoint"), viper.GetString("redis.password")
	RedisDbUtxo = viper.GetInt("redis.db_utxo")

	RpcUrlTestnet, RpcUsernameTestnet, RpcPasswordTestnet = viper.GetString("node.testnet.url"), viper.GetString("node.testnet.username"), viper.GetString("node.testnet.password")
	RpcUrlMainnet, RpcUsernameMainnet, RpcPasswordMainnet = viper.GetString("node.mainnet.url"), viper.GetString("node.mainnet.username"), viper.GetString("node.mainnet.password")
}
