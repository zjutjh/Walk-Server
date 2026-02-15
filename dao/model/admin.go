package model

import (
	"encoding/json"
	"time"
)

const TableNameAdmin = "admins"

// Admin mapped from table <admins>
type Admin struct {
	ID           int64           `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	WxOpenID     *string         `gorm:"column:wx_openid" json:"wx_openid"`
	Name         *string         `gorm:"column:name" json:"name"`
	Account      *string         `gorm:"column:account" json:"account"`
	Password     *string         `gorm:"column:password" json:"password"`
	RouteCode    string          `gorm:"column:route_code;not null;comment:路线代码" json:"route_code"`
	AdminType    int8            `gorm:"column:admin_type;not null;comment:权限级别(1最高权限,2负责人权限,3内部权限,4外部权限)" json:"admin_type"`
	PointID      *int8           `gorm:"column:point_id;default:0" json:"point_id"`
	Campus       *uint8          `gorm:"column:campus;comment:负责校区(1朝晖,2屏峰,3莫干山)" json:"campus"`
	Capabilities json.RawMessage `gorm:"column:capabilities;type:json;comment:权限能力JSON" json:"capabilities"`
	UpdatedAt    time.Time       `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt    time.Time       `gorm:"column:created_at" json:"created_at"`
}

// TableName Admin's table name
func (*Admin) TableName() string {
	return TableNameAdmin
}
