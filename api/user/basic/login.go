package basic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func WechatOAuthCallbackHandler() gin.HandlerFunc {
	api := WechatOAuthCallbackApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfWechatOAuthCallback).Pointer()).Name()] = api
	return hfWechatOAuthCallback
}

type WechatOAuthCallbackApi struct {
	Info     struct{} `name:"微信OAuth回调" desc:"接收微信回调code并换取系统JWT"`
	Request  WechatOAuthCallbackApiRequest
	Response struct{}
}

type WechatOAuthCallbackApiRequest struct {
	Query struct {
		Code string `form:"code" desc:"微信OAuth回调code" binding:"required"`
	}
}

type wechatOAuthAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}

func (h *WechatOAuthCallbackApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindQuery(&h.Request.Query)
}

func (h *WechatOAuthCallbackApi) Run(ctx *gin.Context) kit.Code {
	openID, err := fetchWechatOpenID(h.Request.Query.Code)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("获取微信OpenID失败")
		return comm.CodeOAuthFailed
	}
	if openID == "" {
		return comm.CodeOpenIDEmpty
	}

	tokenOpenID := openID
	if comm.BizConf.AESSecret != "" {
		tokenOpenID, err = comm.AesEncrypt(openID, comm.BizConf.AESSecret)
		if err != nil {
			nlog.Pick().WithContext(ctx).WithError(err).Warn("加密OpenID失败")
			return comm.CodeServerError
		}
	}

	jwt, err := comm.GenerateToken(tokenOpenID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("生成用户JWT失败")
		return comm.CodeServerError
	}

	if comm.BizConf.FrontEndURL != "" {
		redirectURL := comm.BizConf.FrontEndURL
		separator := "?"
		if parsed, parseErr := url.Parse(redirectURL); parseErr == nil && parsed.RawQuery != "" {
			separator = "&"
		}
		ctx.Redirect(http.StatusTemporaryRedirect, redirectURL+separator+"jwt="+url.QueryEscape(jwt))
		ctx.Abort()
	}
	return comm.CodeOK
}

func fetchWechatOpenID(code string) (string, error) {
	endpoint := "https://api.weixin.qq.com/sns/oauth2/access_token"
	query := url.Values{}
	query.Set("appid", comm.BizConf.WechatAppID)
	query.Set("secret", comm.BizConf.WechatSecret)
	query.Set("code", code)
	query.Set("grant_type", "authorization_code")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(endpoint + "?" + query.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data wechatOAuthAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.ErrCode != 0 {
		return "", fmt.Errorf("wechat oauth failed: %d %s", data.ErrCode, data.ErrMsg)
	}
	return data.OpenID, nil
}

func hfWechatOAuthCallback(ctx *gin.Context) {
	api := &WechatOAuthCallbackApi{}
	if err := api.Init(ctx); err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Warn("参数绑定校验错误")
		reply.Fail(ctx, comm.CodeParameterInvalid)
		return
	}
	code := api.Run(ctx)
	if !ctx.IsAborted() {
		if code == comm.CodeOK {
			reply.Reply(ctx, comm.CodeOK, api.Response)
		} else {
			reply.Fail(ctx, code)
		}
	}
}
