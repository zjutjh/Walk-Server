package middleware

import (
	"walk-server/model"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func IsRegistered(context *gin.Context) {
	jwtToken := context.GetHeader("Authorization")
	if jwtToken == "" {
		utility.ResponseError(context, "缺少登录凭证")
		context.Abort()
		return
	} else {
		jwtToken = jwtToken[7:]
	}
	jwtData, err := utility.ParseToken(jwtToken)
	// jwt token 解析失败
	if err != nil {
		utility.ResponseError(context, "jwt error")
		context.Abort()
		return
	}

	if _, err = model.GetPerson(jwtData.OpenID); err != nil {
		utility.ResponseError(context, "请先报名注册")
		context.Abort()
		return
	}

	context.Next()
}
