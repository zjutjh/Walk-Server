package middleware

import (
	"github.com/gin-gonic/gin"
	"time"
	"walk-server/global"
	"walk-server/utility"
)

// 令牌桶实现限流
func TokenRateLimiter(context *gin.Context) {
	// 每次请求拿出一个令牌
	if global.Bucket.TakeAvailable(1) == 0 {
		utility.ResponseData(context, 503, "系统繁忙", nil)
		context.Abort()
		return
	}

	context.Next()
}

// 限制单个用户每秒请求次数
func PerRateLimiter(context *gin.Context) {
	// 获取 jwt 数据
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, _ := utility.ParseToken(jwtToken)

	// 每秒刷新
	global.Rdb.SetNX(global.Rctx, jwtData.OpenID+"Limit", 0, time.Second)

	// 每次访问，对应的值加一
	global.Rdb.Incr(global.Rctx, jwtData.OpenID+"Limit")

	// 获取访问的次数
	val, _ := global.Rdb.Get(global.Rctx, jwtData.OpenID+"Limit").Int()

	// 如果大于5次，返回403
	if val > 5 {
		utility.ResponseData(context, 403, "访问频繁", nil)
		context.Abort()
		return
	}

	context.Next()
}
