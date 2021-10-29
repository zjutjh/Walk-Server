package model

type Team struct {
	ID        uint
	Name      string // 队伍的名字
	Password  string
	Captain   string // 队长的 Open ID
	Route     uint8
	Submitted bool // 是否已经提交
}
