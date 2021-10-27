/*
 * Copyright (c) 2021 IInfo.
 */

package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"walk-server/utility"
)

func Oauth(ctx *gin.Context) {
	redirectUrl := "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" +
		utility.Config.GetString("server.wechatAPPID") +
		"&redirect_uri=" + utility.Config.GetString("server.oauth") + utility.Config.GetString("server.wechatRedirect") +
		"&response_type=code&scope=snsapi_base&state=STATE#wechat_redirect"
	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

func Login(ctx *gin.Context) {
	code := ctx.Query("code") // 微信回调的 code 参数

	// TODO: 对微信返回的 code 做校验

	// 获取用户的 open id
	openID, err := utility.GetOpenID(code)
	fmt.Println(openID) // debug
	if err != nil {
		fmt.Println("open ID 获取失败")
		fmt.Println(err)
	}

	// 生成 JWT
	jwtToken, err := utility.GenerateJWT(openID)
	if err != nil {
		fmt.Println("JWT 生成失败")
		fmt.Println(err)
	}
	ctx.JSON(http.StatusOK, gin.H{"code": "200", "msg": "login", "data": gin.H{
		"jwt": jwtToken,
	}})
}

func Register(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": "200", "msg": "register"})
}
