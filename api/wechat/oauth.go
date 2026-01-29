package wechat

import (
	"app/comm"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	oa "github.com/zjutjh/mygo/wechat/officialAccount"
)

// OauthHandler handles the OAuth redirect
func OauthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		callbackURL := c.Query("callback")
		if callbackURL == "" {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		officialAccount := oa.Pick()
		oauth := officialAccount.OAuth
		redirectURL, err := oauth.Scopes([]string{"snsapi_userinfo"}).Redirect(callbackURL)
		if err != nil {
			reply.Fail(c, comm.CodeThirdServiceError)
			return
		}
		c.Redirect(302, redirectURL)
	}
}
