package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	"app/dao/repo"
)

func WechatLoginHandler() gin.HandlerFunc {
	api := WechatLoginApi{}
	swagger.CM[runtime.FuncForPC(reflect.ValueOf(hfWechatLogin).Pointer()).Name()] = api
	return hfWechatLogin
}

type WechatLoginApi struct {
	Info     struct{} `name:"微信登录" desc:"微信登录并换取系统Token"`
	Request  WechatLoginApiRequest
	Response WechatLoginApiResponse
}

type WechatLoginApiRequest struct {
	Code string `form:"code" binding:"required" desc:"微信临时登录code"`
}

type WechatLoginApiResponse struct {
	Token       string `json:"token" desc:"登录Token"`
	HasRegister bool   `json:"has_register" desc:"是否已完成报名"`
}

func (h *WechatLoginApi) Init(ctx *gin.Context) error {
	return ctx.ShouldBindQuery(&h.Request)
}

func (h *WechatLoginApi) Run(ctx *gin.Context) kit.Code {
	if h.Request.Code == "" {
		return comm.CodeWechatCodeMissing
	}

	openID, err := fetchWechatOpenID(ctx.Request.Context(), h.Request.Code)
	if err != nil || openID == "" {
		nlog.Pick().WithContext(ctx).WithError(err).Error("微信换取OpenID失败")
		return comm.CodeOAuthFailed
	}

	token, err := comm.GenerateToken(openID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("生成Token失败")
		return comm.CodeUnknownError
	}

	peopleRepo := repo.NewPeopleRepo()
	person, err := peopleRepo.FindPeopleByOpenID(ctx, openID)
	if err != nil {
		nlog.Pick().WithContext(ctx).WithError(err).Error("查询用户信息失败")
		return comm.CodeDatabaseError
	}

	h.Response.Token = token
	h.Response.HasRegister = person != nil
	return comm.CodeOK
}

func hfWechatLogin(ctx *gin.Context) {
	api := &WechatLoginApi{}
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

func fetchWechatOpenID(ctx context.Context, code string) (string, error) {
	if comm.BizConf.WechatAppID == "" || comm.BizConf.WechatSecret == "" {
		return "", fmt.Errorf("wechat config missing")
	}
	endpoint := "https://api.weixin.qq.com/sns/jscode2session"
	query := url.Values{}
	query.Set("appid", comm.BizConf.WechatAppID)
	query.Set("secret", comm.BizConf.WechatSecret)
	query.Set("js_code", code)
	query.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 512))
		if readErr != nil {
			return "", fmt.Errorf("wechat status %d and read body failed: %w", resp.StatusCode, readErr)
		}
		return "", fmt.Errorf("wechat status %d: %s", resp.StatusCode, string(body))
	}

	result := struct {
		OpenID  string `json:"openid"`
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.OpenID == "" {
		return "", fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return result.OpenID, nil
}
