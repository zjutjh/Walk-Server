package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

const TableNameTeam = "team"

// Team mapped from table <team>
type Team struct {
	ID         int64                 `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Name       string                `gorm:"column:name;not null" json:"name"`
	Route      uint8                 `gorm:"column:route" json:"route"`
	Password   string                `gorm:"column:password" json:"password"`
	Slogan     string                `gorm:"column:slogan" json:"slogan"`
	AllowMatch bool                  `gorm:"column:allow_match" json:"allow_match"`
	Status     uint8                 `gorm:"column:status" json:"status"`
	StartNum   int                   `gorm:"column:start_num" json:"start_num"`
	Code       string                `gorm:"column:code" json:"code"`
	Submit     bool                  `gorm:"column:submit" json:"submit"`
	Point      int8                  `gorm:"column:point" json:"point"`
	Num        int                   `gorm:"column:num" json:"num"`
	CreatedAt  time.Time             `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP(3)" json:"created_at"`
	UpdatedAt  time.Time             `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP(3)" json:"updated_at"`
	DeletedAt  soft_delete.DeletedAt `gorm:"column:deleted_at;not null;softDelete:milli" json:"-"`
}

// TableName Team's table name
func (*Team) TableName() string {
	return TableNameTeam
}
