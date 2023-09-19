package initial

import (
	"fmt"
	"os"
	"strconv"
	"walk-server/global"

	"github.com/redis/go-redis/v9"
)

func RedisInit() {
	// 从配置文件中读取数据库信息
	host := global.Config.GetString("redis.host")
	port := global.Config.GetString("redis.port")
	password := global.Config.GetString("redis.password")
	db := global.Config.GetInt("redis.DB")

	global.Rdb = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	})

	// 测试是否正常连接
	_, err := global.Rdb.Ping(global.Rctx).Result()
	if err != nil {
		fmt.Println("Redis连接失败")
		fmt.Println(err)
		os.Exit(-1)
	}

	// 初始化每天各路线报名上限
	for i := 0; i <= 3; i++ { // 枚举天数
		for j := 1; j <= 5; j++ { // 枚举路线编号
			key := strconv.Itoa(i*10 + j)
			value := global.Config.GetInt("teamUpperLimit" + "." + strconv.Itoa(i) + "." + strconv.Itoa(j))
			if _, err := global.Rdb.Get(global.Rctx, key).Result(); err == redis.Nil {
				global.Rdb.Set(global.Rctx, key, value, 0)
			}
		}
	}
}
