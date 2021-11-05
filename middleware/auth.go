package middleware

import (
	"github.com/gin-gonic/gin"
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"
)

func Auth(context *gin.Context) {
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

	// 检查 open ID 是否有对应的用户
	openID := jwtData.OpenID
	person := model.Person{}
	result := initial.DB.Where("open_id = ?", openID).First(&person)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "请先报名")
		context.Abort()
		return
	}

	context.Next()
}

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

	result := initial.DB.Where("open_id = ?", jwtData.OpenID).First(&model.Person{})
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "请先报名注册")
		context.Abort()
		return
	}

	context.Next()
}
