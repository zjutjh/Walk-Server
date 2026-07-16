package basic

import (
	"net/http"
	"net/url"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/mygo/foundation/reply"
	"github.com/zjutjh/mygo/kit"
	"github.com/zjutjh/mygo/nlog"
	"github.com/zjutjh/mygo/swagger"

	"app/comm"
)

func WechatOAuthRedirectHandler() gin.HandlerFunc {
	api := WechatOAuthRedirectApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfWechatOAuthRedirect).Pointer()).Name()] = api
	return hfWechatOAuthRedirect
}

type WechatOAuthRedirectApi struct {
	Info     struct{} `name:"微信OAuth入口" desc:"跳转到微信公众号授权页"`
	Request  struct{}
	Response struct{}
}

func (h *WechatOAuthRedirectApi) Init(ctx *gin.Context) error {
	return nil
}

func (h *WechatOAuthRedirectApi) Run(ctx *gin.Context) kit.Code {
	if comm.BizConf.WechatAppID == "" || comm.BizConf.WechatRedirect == "" {
		return comm.CodeServerError
	}

	redirectURL := "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" +
		url.QueryEscape(comm.BizConf.WechatAppID) +
		"&redirect_uri=" + url.QueryEscape(comm.BizConf.WechatRedirect) +
		"&response_type=code&scope=snsapi_userinfo&state=STATE#wechat_redirect"
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
	ctx.Abort()
	return comm.CodeOK
}

func hfWechatOAuthRedirect(ctx *gin.Context) {
	api := &WechatOAuthRedirectApi{}
	err := api.Init(ctx)
	if err != nil {
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
