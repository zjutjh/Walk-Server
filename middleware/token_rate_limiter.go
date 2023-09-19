package middleware

import (
	"github.com/gin-gonic/gin"
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
