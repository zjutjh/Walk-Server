package model

import (
	"time"
)

const TableNameWrongRouteRecord = "wrong_route_records"

// WrongRouteRecord mapped from table <wrong_route_records>
type WrongRouteRecord struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	TeamID          int64     `gorm:"column:team_id;not null;comment:队伍ID" json:"team_id"`
	OriginRouteCode string    `gorm:"column:origin_route_code;not null;comment:原正确路线code" json:"origin_route_code"`
	WrongRouteCode  string    `gorm:"column:wrong_route_code;not null;comment:错走的路线code" json:"wrong_route_code"`
	AdminID         *int64    `gorm:"column:admin_id;comment:记录该情况的管理员ID" json:"admin_id"`
	Remark          *string   `gorm:"column:remark;comment:备注说明" json:"remark"`
	CreatedTime     time.Time `gorm:"column:created_time;not null" json:"created_time"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName WrongRouteRecord's table name
func (*WrongRouteRecord) TableName() string {
	return TableNameWrongRouteRecord
}
