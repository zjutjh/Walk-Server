package comm

// BizConf 业务配置
var BizConf BizConfig

type BizConfig struct {
	AESSecret     string `mapstructure:"aes_secret"`
	WechatAppID   string `mapstructure:"wechat_app_id"`
	WechatSecret  string `mapstructure:"wechat_secret"`
	FrontEndURL   string `mapstructure:"front_end_url"`
	ExpiredDate   string `mapstructure:"expired_date"`
	OpenDate      string `mapstructure:"open_date"`
	SubmitDate    string `mapstructure:"submit_date"`
	MaxTeamSize   int    `mapstructure:"max_team_size"`
	MinSubmitSize int    `mapstructure:"min_submit_size"`
}
