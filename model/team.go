package model

type Team struct {
	TeamID    string `gorm:"primaryKey"`
	Password  string
	Captain   string // 队长的 Open ID
	Route     uint8
	Submitted bool // 是否已经提交
}
