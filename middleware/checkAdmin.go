package middleware

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"walk-server/global"
	"walk-server/service/adminService"
	"walk-server/utility"
)

func CheckAdmin(context *gin.Context) {
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

	userID := utility.AesDecrypt(jwtData.OpenID, global.Config.GetString("server.AESSecret"))
	user_id, err := strconv.Atoi(userID)
	if err != nil {
		utility.ResponseError(context, "jwt error")
		context.Abort()
		return
	}

	user, err := adminService.GetAdminByID(uint(user_id))
	if err != nil {
		utility.ResponseError(context, "jwt error")
		context.Abort()
		return
	}

	if user == nil {
		utility.ResponseError(context, "未登陆")
		return
	}
	context.Next()
}
