package wechat

import (
	"log"
	"strings"
	"walk-server/global"
)

type wechatConfig struct {
	Driver    string
	AppId     string
	AppSecret string
}

func getConfigs() wechatConfig {

	wc := wechatConfig{}
	if !global.Config.IsSet("wechat.appid") {
		log.Fatal("ConfigError")
	}
	if !global.Config.IsSet("wechat.appsecret") {
		log.Fatal("ConfigError")
	}
	wc.AppId = global.Config.GetString("wechat.appid")
	wc.AppSecret = global.Config.GetString("wechat.appsecret")

	wc.Driver = string(Memory)
	if global.Config.IsSet("wechat.driver") {
		wc.Driver = strings.ToLower(global.Config.GetString("wechat.driver"))
	}
	return wc
}
