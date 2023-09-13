package initial

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"walk-server/global"
)

func RedisInit() {
	host := global.Config.GetString("redis.host")
	port := global.Config.GetString("redis.port")
	password := global.Config.GetString("redis.password")
	db := global.Config.GetInt("redis.DB")

	global.Rdb = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	})

	_, err := global.Rdb.Ping(global.Rctx).Result()
	if err != nil {
		fmt.Println("Redis连接失败")
		os.Exit(-1)
	}
}
