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

	WsPort = ""
	RedisEndpoint = ""
	RedisPassword = ""
	RedisDbUtxo int = 1

	PlatformMainnetPrivateKeySendBrc20 = ""
	PlatformMainnetAddressSendBrc20 = ""//address for send brc20
	PlatformMainnetPrivateKeyReceiveBrc20 = ""
	PlatformMainnetAddressReceiveBrc20 = ""//address for receive brc20
	PlatformMainnetPrivateKeyReceiveBidValue = ""
	PlatformMainnetAddressReceiveBidValue = ""// address for receive bid value
	PlatformMainnetPrivateKeyReceiveDummyValue = ""
	PlatformMainnetAddressReceiveDummyValue = ""// address for receive dummy 1200 value
	PlatformMainnetFeeRate int64 = 0

	PlatformTestnetPrivateKeySendBrc20 = ""
	PlatformTestnetAddressSendBrc20 = ""//address for send brc20
	PlatformTestnetPrivateKeyReceiveBrc20 = ""
	PlatformTestnetAddressReceiveBrc20 = ""//address for receive brc20
	PlatformTestnetPrivateKeyReceiveBidValue = ""
	PlatformTestnetAddressReceiveBidValue = ""// address for receive bid value
	PlatformTestnetPrivateKeyReceiveDummyValue = ""
	PlatformTestnetAddressReceiveDummyValue = ""// address for receive dummy 1200 value
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
	PlatformTestnetPrivateKeySendBrc20, PlatformTestnetAddressSendBrc20,
	PlatformTestnetPrivateKeyReceiveBrc20, PlatformTestnetAddressReceiveBrc20,
		PlatformTestnetPrivateKeyReceiveBidValue, PlatformTestnetAddressReceiveBidValue,
		PlatformTestnetPrivateKeyReceiveDummyValue, PlatformTestnetAddressReceiveDummyValue =
		viper.GetString("platform.testnet.private_key_send_brc20"), viper.GetString("platform.testnet.address_send_brc20"),
		viper.GetString("platform.testnet.private_key_receive_brc20"), viper.GetString("platform.testnet.address_receive_brc20"),
		viper.GetString("platform.testnet.private_key_receive_bid_value"), viper.GetString("platform.testnet.address_receive_bid_value"),
		viper.GetString("platform.testnet.private_key_receive_dummy_value"), viper.GetString("platform.testnet.address_receive_dummy_value")
	PlatformMainnetPrivateKeySendBrc20, PlatformMainnetAddressSendBrc20,
	PlatformMainnetPrivateKeyReceiveBrc20, PlatformMainnetAddressReceiveBrc20,
		PlatformMainnetPrivateKeyReceiveBidValue, PlatformMainnetAddressReceiveBidValue,
		PlatformMainnetPrivateKeyReceiveDummyValue, PlatformMainnetAddressReceiveDummyValue =
		viper.GetString("platform.mainnet.private_key_Send_brc20"), viper.GetString("platform.mainnet.address_Send_brc20"),
		viper.GetString("platform.mainnet.private_key_receive_brc20"), viper.GetString("platform.mainnet.address_receive_brc20"),
		viper.GetString("platform.mainnet.private_key_receive_bid_value"), viper.GetString("platform.mainnet.address_receive_bid_value"),
		viper.GetString("platform.mainnet.private_key_receive_dummy_value"), viper.GetString("platform.mainnet.address_receive_dummy_value")
	PlatformMainnetFeeRate, PlatformTestnetFeeRate = viper.GetInt64("platform.mainnet.fee_rate"), viper.GetInt64("platform.testnet.fee_rate")
		MempoolSpace = viper.GetString("mempool_space.domain")
	WsPort = viper.GetString("ws.port")
	RedisEndpoint, RedisPassword = viper.GetString("redis.endpoint"), viper.GetString("redis.password")
	RedisDbUtxo = viper.GetInt("redis.db_utxo")
}
