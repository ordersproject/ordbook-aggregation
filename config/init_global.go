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
	PlatformPrivateKey = ""
	PlatformTaprootAddress = ""
	PlatformTaprootAddress2 = ""
)

func InitConfig() {
	viper.SetConfigFile(conf.GetYaml())
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	Port = viper.GetString("port")
	HiroDomain = viper.GetString("hiro_domain")
	OklinkDomain, OklinkKey = viper.GetString("oklink.domain"), viper.GetString("oklink.key")
	PlatformPrivateKey, PlatformTaprootAddress, PlatformTaprootAddress2 =
		viper.GetString("platform.private_key"), viper.GetString("platform.taproot_address"), viper.GetString("platform.taproot_address_2")
	MempoolSpace = viper.GetString("mempool_space.domain")
}
