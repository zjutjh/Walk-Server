package basic

import (
	"net/http"
	"walk-server/utility/initial"

	"github.com/gin-gonic/gin"
)

func Oauth(ctx *gin.Context) {
	redirectUrl := "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" +
		initial.Config.GetString("server.wechatAPPID") +
		"&redirect_uri=" + initial.Config.GetString("server.oauth") + initial.Config.GetString("server.wechatRedirect") +
		"&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect"
	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}