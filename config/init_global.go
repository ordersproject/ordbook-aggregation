package config

import (
	"fmt"
	"github.com/spf13/viper"
	"ordbook-aggregation/conf"
)

var (
	Port = ""
)

func InitConfig() {
	viper.SetConfigFile(conf.GetYaml())
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	Port = viper.GetString("port")
}
