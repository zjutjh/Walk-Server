package initial

import (
	"fmt"
	"github.com/spf13/viper"
)

var Config = viper.New()

func ConfigInit() {
	Config.SetConfigType("yaml")
	Config.AddConfigPath("./config")
	Config.SetConfigName("config")
	Config.WatchConfig() // Watch the change of the config file

	err := Config.ReadInConfig()
	if err != nil {
		fmt.Println("配置读取错误! ")
		fmt.Println(err)
	}
}
