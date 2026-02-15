package model

import (
	"time"
)

const TableNamePoint = "points"

// Point mapped from table <points>
type Point struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true;comment:全局唯一点位ID" json:"id"`
	Cid       int64     `gorm:"column:cid;not null;comment:校区内点位编号(可跨校区重复)" json:"cid"`
	Code      string    `gorm:"column:code;not null;comment:点位二维码唯一标识" json:"code"`
	Name      *string   `gorm:"column:name;comment:点位名称" json:"name"`
	IsActive  *bool     `gorm:"column:is_active;default:1;comment:是否启用" json:"is_active"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName Point's table name
func (*Point) TableName() string {
	return TableNamePoint
}
