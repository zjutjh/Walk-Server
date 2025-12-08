package wechat

import (
	"app/comm"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/zjutjh/mygo/foundation/reply"
)

// OauthHandler handles the OAuth redirect
func OauthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		callbackURL := c.Query("callback")
		if callbackURL == "" {
			reply.Fail(c, comm.CodeParameterInvalid)
			return
		}

		oa := do.MustInvoke[*officialAccount.OfficialAccount](nil)
		oauth := oa.OAuth
		redirectURL, err := oauth.Scopes([]string{"snsapi_userinfo"}).Redirect(callbackURL)
		if err != nil {
			reply.Fail(c, comm.CodeThirdServiceError)
			return
		}
		c.Redirect(302, redirectURL)
	}
}
