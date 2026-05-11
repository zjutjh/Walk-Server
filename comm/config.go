package comm

// BizConf 业务配置
var BizConf BizConfig

type BizConfig struct {
	AESSecret     string `mapstructure:"aes_secret"`
	WechatAppID   string `mapstructure:"wechat_app_id"`
	WechatSecret  string `mapstructure:"wechat_secret"`
	ExpiredDate   string `mapstructure:"expired_date"`
	MaxTeamSize   int    `mapstructure:"max_team_size"`
}
