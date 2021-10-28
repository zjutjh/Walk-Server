package middleware

import (
	"github.com/gin-gonic/gin"
	"walk-server/utility"
)

// TimeValidity Require implement ... Check if in open time
func TimeValidity(ctx *gin.Context) {
	// TODO: 开放时间中间件尚未完工
	ctx.Next()
}

// RegisterJWTValidity 注册的时候验证 JWT 是否合法
func RegisterJWTValidity(context *gin.Context) {
	jwtToken := context.GetHeader("Authorization")[7:]
	_, err := utility.ParseToken(jwtToken)

	if err != nil {
		utility.ResponseError(context, "请先登陆")
	} else {
		context.Next() // 转到 controller 继续执行
	}
}
