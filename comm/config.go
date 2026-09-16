package comm

// BizConf 业务配置
var BizConf BizConfig

type BizConfig struct {
	IdentitySecret  string           `mapstructure:"identity_secret"`
	Phases          PhaseConfig      `mapstructure:"phases"`
	MaxTeamSize     int              `mapstructure:"max_team_size"`
	TeamTotalLimit  int              `mapstructure:"team_total_limit"`
	DailyTeamLimits map[string][]int `mapstructure:"daily_team_limits"`
}

type TimeRangeConfig struct {
	Start string `mapstructure:"start"`
	End   string `mapstructure:"end"`
}

type SubmissionPhaseConfig struct {
	TimeRangeConfig `mapstructure:",squash"`
	DailyStartTime  string `mapstructure:"daily_start_time"`
	DailyEndTime    string `mapstructure:"daily_end_time"`
}

type PhaseConfig struct {
	Registration TimeRangeConfig       `mapstructure:"registration"`
	Submission   SubmissionPhaseConfig `mapstructure:"submission"`
	Adjustment   TimeRangeConfig       `mapstructure:"adjustment"`
	Preparation  TimeRangeConfig       `mapstructure:"preparation"`
	Activity     TimeRangeConfig       `mapstructure:"activity"`
}
