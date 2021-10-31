/*
 * Copyright (c) 2021 IInfo.
 */

package controller

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"walk-server/utility"
	"walk-server/utility/initial"
)

func Oauth(ctx *gin.Context) {
	redirectUrl := "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" +
		initial.Config.GetString("server.wechatAPPID") +
		"&redirect_uri=" + initial.Config.GetString("server.oauth") + initial.Config.GetString("server.wechatRedirect") +
		"&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect"
	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

func Login(ctx *gin.Context) {
	var jwtData utility.JwtData
	code := ctx.Query("code") // 微信回调的 code 参数

	// TODO 对微信返回的 code 做校验

	// 获取用户的 open id
	openID, err := utility.GetOpenID(code)
	if err != nil {
		utility.ResponseError(ctx, "open ID 错误，请重新打开网页重试")
		return
	}
	jwtData.OpenID = fmt.Sprintf("%x", md5.Sum([]byte(openID)))

	// 生成 JWT
	jwtToken, err := utility.GenerateStandardJwt(&jwtData)
	if err != nil {
		utility.ResponseError(ctx, "登陆错误，请重新打开网页重试")
		return
	}
	utility.ResponseSuccess(ctx, gin.H{
		"jwt": jwtToken,
	})
}
