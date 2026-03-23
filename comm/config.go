package comm

import "time"

const defaultRegisterLockTTL = 5 * time.Second

// BizConf 业务配置
var BizConf *BizConfig

type BizConfig struct {
	JWTSecret       string `mapstructure:"jwt_secret"`
	AESSecret       string `mapstructure:"aes_secret"`
	WechatAppID     string `mapstructure:"wechat_app_id"`
	WechatSecret    string `mapstructure:"wechat_secret"`
	FrontEndURL     string `mapstructure:"front_end_url"`
	ExpiredDate     string `mapstructure:"expired_date"`
	OpenDate        string `mapstructure:"open_date"`
	SubmitDate      string `mapstructure:"submit_date"`
	RegisterLockTTL string `mapstructure:"register_lock_ttl"`
	MaxTeamSize     int    `mapstructure:"max_team_size"`
	MinSubmitSize   int    `mapstructure:"min_submit_size"`
}

func (c *BizConfig) GetRegisterLockTTL() time.Duration {
	if c == nil || c.RegisterLockTTL == "" {
		return defaultRegisterLockTTL
	}

	ttl, err := time.ParseDuration(c.RegisterLockTTL)
	if err != nil || ttl <= 0 {
		return defaultRegisterLockTTL
	}

	return ttl
}
