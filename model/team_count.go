package model

type TeamCount struct {
	DayCampus uint8 `gorm:"primaryKey"` // 天数和校区编号
	Count     int   // 对应的报名团队数量
}
