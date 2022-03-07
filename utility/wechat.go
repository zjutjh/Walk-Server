package utility

import (
	"time"
	"walk-server/global"

	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

// GetOpenID 获取用户的 open ID
func GetOpenID(code string) (string, error) {
	client := resty.New()
	getOpenIDURL := "https://api.weixin.qq.com/sns/oauth2/access_token"
	resp, err := client.R().SetQueryParams(map[string]string{
		"appid":      global.Config.GetString("server.wechatAPPID"),
		"secret":     global.Config.GetString("server.wechatSecret"),
		"code":       code,
		"grant_type": "authorization_code",
	}).Get(getOpenIDURL)
	if err != nil {
		return "", err
	}

	jsonData := string(resp.Body())
	return gjson.Get(jsonData, "openid").String(), nil
}

// GetAccessToken 获取用户 access token
func GetAccessToken(wechatAPPID string, wechatSecret string) (string, error) {
	// 先从缓存中获取 access_token
	if x, found := global.Cache.Get("access_token"); found {
		return x.(string), nil
	}
	
	// 没有在缓存中找到 重新更新 access token
	client := resty.New()
	resp, err := client.R().
      SetQueryParams(map[string]string{
          "grant_type": "client_credential",
		  "appid": wechatAPPID,
		  "secret": wechatSecret,
      }).
      SetHeader("Accept", "application/json").
      Get("https://api.weixin.qq.com/cgi-bin/token")
	
	if err != nil {
		return "", err
	}

	accessToken := gjson.Get(string(resp.Body()), "access_token").String()
	
	// 缓存 access token
	expireTime := gjson.Get(string(resp.Body()), "expires_in").Int() - 60 * 30 // 单位 s (缓存比 access token 过期时间早 30 分种)
	global.Cache.Set("access_token", accessToken, time.Duration(expireTime) * time.Second)
	
	return accessToken, nil
}