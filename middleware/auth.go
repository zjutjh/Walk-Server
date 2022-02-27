package middleware

import (
	"walk-server/global"
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
		utility.ResponseError(context, "登录错误，重新进入网页试试")
		context.Abort()
		return
	}

	result := global.DB.Where("open_id = ?", jwtData.OpenID).Take(&model.Person{})
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "请先报名注册")
		context.Abort()
		return
	}

	context.Next()
}
