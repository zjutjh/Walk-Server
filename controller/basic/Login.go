package basic

import (
	"fmt"
	"net/http"
	"walk-server/global"
	"walk-server/utility"

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
	openID, err := utility.GetOpenID(code)
	if err != nil {
		utility.ResponseError(ctx, "open ID 错误，请重新打开网页重试")
		return
	} else if openID == "" {
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