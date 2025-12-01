package wechat

import (
	"log"
	"walk-server/global"

	"github.com/zjutjh/mygo/wechat/miniprogram"
	"github.com/zjutjh/mygo/wechat/officialAccount"
)

type driver string

const (
	Memory driver = "memory"
	Redis  driver = "redis"
)

func WeChatInit() {
	config := getConfigs()

	// Initialize MiniProgram
	var err error
	global.MiniProgram, err = miniprogram.New(miniprogram.Config{
		AppId:     config.AppId,
		AppSecret: config.AppSecret,
		Driver:    config.Driver,
	})
	if err != nil {
		log.Fatalf("Failed to initialize MiniProgram: %v", err)
	}

	// Initialize OfficialAccount
	// Read Official Account config
	oaAppID := global.Config.GetString("server.wechatAPPID")
	oaSecret := global.Config.GetString("server.wechatSecret")

	global.OfficialAccount, err = officialAccount.New(officialAccount.Config{
		AppID:  oaAppID,
		Secret: oaSecret,
		Token:  global.Config.GetString("server.wechatToken"),
		AESKey: global.Config.GetString("server.wechatEncodingAESKey"),
		Driver: config.Driver,
	})
	if err != nil {
		log.Fatalf("Failed to initialize OfficialAccount: %v", err)
	}
}
