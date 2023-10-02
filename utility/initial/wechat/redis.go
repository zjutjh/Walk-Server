package wechat

import (
	"github.com/silenceper/wechat/v2/cache"
	"walk-server/global"
)

func setRedis(wcCache cache.Cache) cache.Cache {
	db := global.Config.GetInt("redis.DB")
	host := global.Config.GetString("redis.host")
	port := global.Config.GetString("redis.port")
	password := global.Config.GetString("redis.password")
	redisOpts := &cache.RedisOpts{
		Host:        host + ":" + port,
		Database:    db,
		MaxActive:   10,
		MaxIdle:     10,
		IdleTimeout: 60,
		Password:    password,
	}
	wcCache = cache.NewRedis(global.Wctx, redisOpts)
	return wcCache
}
