package comm

// BizConf 业务配置
var BizConf BizConfig

type BizConfig struct {
	IdentitySecret  string `mapstructure:"identity_secret"`
	StartDate       string `mapstructure:"start_date"`
	ExpiredDate     string `mapstructure:"expired_date"`
	MaxTeamSize     int    `mapstructure:"max_team_size"`
	TeamTotalLimit  int    `mapstructure:"team_total_limit"`
	DailyTeamLimits []int  `mapstructure:"daily_team_limits"`
}
