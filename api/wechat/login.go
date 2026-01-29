package wechat

import (
	"app/comm"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/jwt"
	oa "github.com/zjutjh/mygo/wechat/officialAccount"
)

// LoginHandler 处理OAuth回调和登录
func LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		officialAccount := oa.Pick()
		oauth := officialAccount.OAuth
		user, err := oauth.UserFromCode(code)
		if err != nil {
			reply.Fail(c, comm.CodeThirdServiceError)
			return
		}

		token, err := jwt.Pick().GenerateToken(user.GetOpenID())
		if err != nil {
			reply.Fail(c, comm.CodeUnknownError)
			return
		}

		reply.Success(c, gin.H{
			"token":   token,
			"open_id": user.GetOpenID(),
			"user":    user,
		})
	}
}
