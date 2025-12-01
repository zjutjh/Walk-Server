package basic

import (
	"net/http"
	"walk-server/global"
	"walk-server/utility"

	"github.com/gin-gonic/gin"
)

func Oauth(ctx *gin.Context) {
	oauth := global.OfficialAccount.OAuth
	redirectURI := global.Config.GetString("server.wechatRedirect")

	redirectUrl, err := oauth.Scopes([]string{"snsapi_userinfo"}).
		WithState("STATE").
		Redirect(redirectURI)

	if err != nil {
		utility.ResponseError(ctx, "生成授权链接失败")
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}
