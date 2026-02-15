package model

import (
	"time"
)

const TableNameCheckin = "checkins"

// Checkin mapped from table <checkins>
type Checkin struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	AdminID   *int64    `gorm:"column:admin_id;comment:签到管理员ID" json:"admin_id"`
	TeamID    int64     `gorm:"column:team_id;not null;comment:队伍ID" json:"team_id"`
	PointID   *int8     `gorm:"column:point_id;comment:签到点位ID" json:"point_id"`
	RouteCode string    `gorm:"column:route_code;not null;comment:路线代码" json:"route_code"`
	Time      time.Time `gorm:"column:time;comment:签到时间" json:"time"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName Checkin's table name
func (*Checkin) TableName() string {
	return TableNameCheckin
}
