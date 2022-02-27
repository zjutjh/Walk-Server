package initial

import (
	"fmt"
	"walk-server/global"
)

func ConfigInit() {
	global.Config.SetConfigType("yaml")
	global.Config.AddConfigPath("./config")
	global.Config.SetConfigName("config")
	global.Config.WatchConfig() // Watch the change of the config file

	err := global.Config.ReadInConfig()
	if err != nil {
		fmt.Println("配置读取错误! ")
		fmt.Println(err)
	}
}
