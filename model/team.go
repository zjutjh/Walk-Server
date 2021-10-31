package model

type Team struct {
	ID        uint
	Name      string // 队伍的名字
	Num       uint8  // 团队里的人数
	Password  string
	Captain   string // 队长的 Open ID
	Route     uint8  // 1 是朝晖路线，2 屏峰半程，3 屏峰全程，4 莫干山半程，5 莫干山全程
	Submitted bool   // 是否已经提交
}
