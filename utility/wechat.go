package utility

import (
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"walk-server/utility/initial"
)

// GetOpenID 获取用户的 open ID
func GetOpenID(code string) (string, error) {
	client := resty.New()
	getOpenIDURL := "https://api.weixin.qq.com/sns/oauth2/access_token"
	resp, err := client.R().SetQueryParams(map[string]string{
		"appid":      initial.Config.GetString("server.wechatAPPID"),
		"secret":     initial.Config.GetString("server.wechatSecret"),
		"code":       code,
		"grant_type": "authorization_code",
	}).Get(getOpenIDURL)
	if err != nil {
		return "", err
	}

	jsonData := string(resp.Body())
	return gjson.Get(jsonData, "openid").String(), nil
}
