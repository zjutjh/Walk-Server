package middleware

import (
	"github.com/gin-gonic/gin"
	"walk-server/utility"
)

// TimeValidity Require implement ... Check if in open time
func TimeValidity(ctx *gin.Context) {
	if !utility.CanOpenApi() {
		utility.ResponseError(ctx, "time error")
		ctx.Abort()
		return
	}

	ctx.Next()
}

// CanSubmit 是否开发提交队伍
func CanSubmit(context *gin.Context) {
	if !utility.CanSubmit() {
		utility.ResponseError(context, "尚且不能提交")
		context.Abort()
	} else {
		context.Next()
	}
}

// RegisterJWTValidity 注册的时候验证 JWT 是否合法
func RegisterJWTValidity(context *gin.Context) {
	jwtToken := context.GetHeader("Authorization")
	if jwtToken == "" {
		utility.ResponseError(context, "缺少登录凭证")
		context.Abort()
		return
	} else {
		jwtToken = jwtToken[7:]
	}
	_, err := utility.ParseToken(jwtToken)

	if err != nil {
		utility.ResponseError(context, "请先登录")
		context.Abort()
	} else {
		context.Next() // 转到 controller 继续执行
	}
}
