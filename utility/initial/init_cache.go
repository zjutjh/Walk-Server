package initial

import (
	"time"
	"walk-server/global"

	"github.com/patrickmn/go-cache"
)


func MemCacheInit() {
	// 缓存 20 分钟过期 10 分种检测, 每 10 分种删除过期缓存
	global.Cache = cache.New(20*time.Minute, 10*time.Minute)
}
