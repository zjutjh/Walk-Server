package wechat

import (
	"app/comm"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/miniProgram"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/zjutjh/mygo/foundation/reply"
)

// MiniProgramLoginHandler handles the Mini Program login
func MiniProgramLoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		miniProgram := do.MustInvoke[*miniProgram.MiniProgram](nil)
		session, err := miniProgram.Auth.Session(c.Request.Context(), code)
		if err != nil {
			reply.Fail(c, comm.CodeThirdServiceError)
			return
		}

		// TODO: Implement your login logic here (e.g., create user, generate token)

		reply.Success(c, session)
	}
}
