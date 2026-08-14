package comm

// BizConf 业务配置
var BizConf BizConfig

type BizConfig struct {
	AESSecret       string `mapstructure:"aes_secret"`
	IdentitySecret  string `mapstructure:"identity_secret"`
	WechatAppID     string `mapstructure:"wechat_app_id"`
	WechatSecret    string `mapstructure:"wechat_secret"`
	WechatRedirect  string `mapstructure:"wechat_redirect"`
	FrontEndURL     string `mapstructure:"front_end_url"`
	StartDate       string `mapstructure:"start_date"`
	ExpiredDate     string `mapstructure:"expired_date"`
	MaxTeamSize     int    `mapstructure:"max_team_size"`
	TeamTotalLimit  int    `mapstructure:"team_total_limit"`
	DailyTeamLimits []int  `mapstructure:"daily_team_limits"`
}
