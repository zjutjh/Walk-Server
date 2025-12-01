package basic

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"walk-server/global"
	"walk-server/service/userService"
	"walk-server/utility"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	var jwtData utility.JwtData
	code := ctx.Query("code") // 微信回调的 code 参数

	if code == "" {
		utility.ResponseError(ctx, "请在微信客户端中打开")
		return
	}
	// 获取用户的 open id
	oauth := global.OfficialAccount.OAuth
	user, err := oauth.UserFromCode(code)
	if err != nil {
		utility.ResponseError(ctx, "open ID 错误，请重新打开网页重试")
		return
	}
	openID := user.GetOpenID()
	if openID == "" {
		utility.ResponseError(ctx, "请在微信中打开")
		return
	}
	jwtData.OpenID = utility.AesEncrypt(openID, global.Config.GetString("server.AESSecret"))

	// 生成 JWT
	urlToken, err := utility.UrlToken(&jwtData)
	if err != nil {
		utility.ResponseError(ctx, "登录错误，请重新打开网页重试")
		return
	}

	// 如果在调试模式下就输出用户的 jwt token
	if utility.IsDebugMode() {
		fmt.Printf("[Debug Info] %v\n", urlToken)
	}

	frontEndUrl := global.Config.GetString("frontEnd.url")
	redirectUrl := frontEndUrl + "?jwt=" + urlToken
	ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

func LoginByOpenID(ctx *gin.Context) {
	var jwtData utility.JwtData
	openID := ctx.DefaultQuery("open_id", "")
	if openID == "" {
		utility.ResponseError(ctx, "openID 为空")
		return
	}
	decodedOpenID, err := url.QueryUnescape(openID)
	if err != nil {
		utility.ResponseError(ctx, "无法解码 open_id")
		return
	}

	// 如果解码后的字符串中有空格，可以恢复为 "+"
	decodedOpenID = strings.Replace(decodedOpenID, " ", "+", -1)

	user, err := userService.GetUserByOpenID(decodedOpenID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utility.ResponseError(ctx, "获取用户信息失败")
		return
	}

	jwtData.OpenID = decodedOpenID

	// 生成 JWT
	urlToken, err := utility.GenerateStandardJwt(&jwtData)
	if err != nil {
		utility.ResponseError(ctx, "登录错误，请重新打开网页重试")
		return
	}

	// 如果在调试模式下就输出用户的 jwt token
	if utility.IsDebugMode() {
		fmt.Printf("[Debug Info] %v\n", urlToken)
	}

	utility.ResponseSuccess(ctx, gin.H{
		"jwt":  urlToken,
		"user": user,
	})
}
