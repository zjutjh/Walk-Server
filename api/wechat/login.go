package wechat

import (
	"app/comm"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/jwt"
)

// LoginHandler handles the OAuth callback and login
func LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		oa := do.MustInvoke[*officialAccount.OfficialAccount](nil)
		oauth := oa.OAuth
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
