package wechat

import (
	"app/comm"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/zjutjh/mygo/foundation/reply"
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

		// TODO: Implement your login logic here (e.g., create user, generate token)

		reply.Success(c, user)
	}
}
