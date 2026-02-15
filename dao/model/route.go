package model

import (
	"encoding/json"
	"time"
)

const TableNameRoute = "routes"

// Route mapped from table <routes>
type Route struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Code        string          `gorm:"column:code;not null;comment:路线代码" json:"code"`
	Name        string          `gorm:"column:name;not null;comment:路线名称" json:"name"`
	Campus      uint8           `gorm:"column:campus;not null;comment:所属校区(1朝晖,2屏峰,3莫干山)" json:"campus"`
	TotalPoints uint8           `gorm:"column:total_points;not null;comment:总点位数" json:"total_points"`
	PointList   json.RawMessage `gorm:"column:point_list;not null;type:json;comment:路线包含的点位列表" json:"point_list"`
	IsActive    bool            `gorm:"column:is_active;not null;default:1;comment:是否启用" json:"is_active"`
	Description *string         `gorm:"column:description;comment:路线描述" json:"description"`
	UpdatedAt   time.Time       `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt   time.Time       `gorm:"column:created_at" json:"created_at"`
}

// TableName Route's table name
func (*Route) TableName() string {
	return TableNameRoute
}
