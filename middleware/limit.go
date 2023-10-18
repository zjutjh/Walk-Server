package middleware

import (
	"time"
	"walk-server/global"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// 令牌桶实现限流
func TokenRateLimiter(context *gin.Context) {
	// 每次请求拿出一个令牌
	if global.Bucket.TakeAvailable(1) == 0 {
		utility.ResponseError(context, "系统繁忙")
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
	global.Rdb.SetNX(global.Rctx, jwtData.OpenID+"Limit", 0, time.Minute)

	// 每次访问，对应的值加一
	global.Rdb.Incr(global.Rctx, jwtData.OpenID+"Limit")

	// 获取访问的次数
	val, _ := global.Rdb.Get(global.Rctx, jwtData.OpenID+"Limit").Int()

	if val > 200 {
		utility.ResponseError(context, "访问频繁")
		context.Abort()
		return
	}

	context.Next()
}
