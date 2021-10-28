package middleware

import (
	"github.com/gin-gonic/gin"
	"walk-server/model"
	"walk-server/utility"
	"walk-server/utility/initial"
)

func Auth(context *gin.Context) {
	jwtToken := context.GetHeader("Authorization")[7:]
	jwtData, err := utility.ParseToken(jwtToken)
	// jwt token 解析失败
	if err != nil {
		utility.ResponseError(context, "登陆错误，重新进入网页试试")
		return
	}

	// 检查 open ID 是否有对应的用户
	openID := jwtData.OpenID
	person := model.Person{}
	result := initial.DB.Where("open_id = ?", openID).First(&person)
	if result.RowsAffected == 0 {
		utility.ResponseError(context, "请先报名")
		return
	}

	context.Next()
}
