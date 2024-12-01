package middleware

import (
	"time"
	"walk-server/global"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

// TimeValidity Require implement ... Check if in open time
func TimeValidity(ctx *gin.Context) {
	if !utility.CanOpenApi() {
		utility.ResponseError(ctx, "还没到开放时间，不能访问哦")
		ctx.Abort()
		return
	}

	ctx.Next()
}

// IsExpired 检查是否过了报名时间，报名时间过了就无法修改用户信息了
func IsExpired(context *gin.Context) {
	expiredTime, _ := time.ParseInLocation(
		utility.TimeLayout,
		global.Config.GetString("expiredDate"),
		time.Local,
	)
	expiredTimeUnix := expiredTime.Unix()
	deltaTimeUnix := time.Now().Unix() - expiredTimeUnix

	if deltaTimeUnix >= 0 { // 过期了
		utility.ResponseError(context, "报名截止了哦")
		context.Abort()
	} else {
		context.Next()
	}
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
