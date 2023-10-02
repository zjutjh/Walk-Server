package wechat

import (
	wechat "github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	miniConfig "github.com/silenceper/wechat/v2/miniprogram/config"
	"log"
	"walk-server/global"
)

type driver string

const (
	Memory driver = "memory"
	Redis  driver = "redis"
)

func WeChatInit() {
	config := getConfigs()

	wc := wechat.NewWechat()
	var wcCache cache.Cache
	switch config.Driver {
	case string(Redis):
		wcCache = setRedis(wcCache)
		break
	case string(Memory):
		wcCache = cache.NewMemory()
		break
	default:
		log.Fatal("ConfigError")
	}

	cfg := &miniConfig.Config{
		AppID:     config.AppId,
		AppSecret: config.AppSecret,
		Cache:     wcCache,
	}

	global.MiniProgram = wc.GetMiniProgram(cfg)
}
