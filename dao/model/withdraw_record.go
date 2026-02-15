package model

import (
	"time"
)

const TableNameWithdrawRecord = "withdraw_records"

// WithdrawRecord mapped from table <withdraw_records>
type WithdrawRecord struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	PersonOpenID string    `gorm:"column:person_openid;not null;comment:下撤人员OpenID" json:"person_openid"`
	TeamID       int64     `gorm:"column:team_id;not null;comment:队伍ID" json:"team_id"`
	RouteCode    string    `gorm:"column:route_code;not null;comment:所在路线code" json:"route_code"`
	PointID      int64     `gorm:"column:point_id;not null;comment:下撤点位 CPn" json:"point_id"`
	Reason       *string   `gorm:"column:reason;comment:下撤原因" json:"reason"`
	AdminID      *int64    `gorm:"column:admin_id;comment:处理的管理员ID" json:"admin_id"`
	WithdrawTime time.Time `gorm:"column:withdraw_time;not null" json:"withdraw_time"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName WithdrawRecord's table name
func (*WithdrawRecord) TableName() string {
	return TableNameWithdrawRecord
}
