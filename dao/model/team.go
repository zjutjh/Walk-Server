package model

import (
	"time"
)

const TableNameTeam = "teams"

// Team mapped from table <teams>
type Team struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement:true;comment:队伍ID" json:"id"`
	Name        string    `gorm:"column:name;not null;comment:队伍名称" json:"name"`
	Num         uint8     `gorm:"column:num;not null;default:1;comment:团队人数" json:"num"`
	Password    string    `gorm:"column:password;not null;comment:团队加入密码" json:"password"`
	Slogan      *string   `gorm:"column:slogan;comment:团队标语" json:"slogan"`
	AllowMatch  bool      `gorm:"column:allow_match;not null;default:0;comment:是否允许随机匹配" json:"allow_match"`
	Captain     string    `gorm:"column:captain;not null;comment:队长OpenID" json:"captain"`
	Route       string    `gorm:"column:route;not null;comment:团队所属路线代码" json:"route"`
	PointID     *int8     `gorm:"column:point_id;default:0;comment:当前所在点位ID" json:"point_id"`
	IsDeparted  bool      `gorm:"column:is_departed;not null;default:0;comment:是否出发" json:"is_departed"`
	IsCompleted bool      `gorm:"column:is_completed;not null;default:0;comment:是否已完成" json:"is_completed"`
	Submit      bool      `gorm:"column:submit;not null;default:0;comment:是否已提交报名" json:"submit"`
	Code        *string   `gorm:"column:code;comment:签到二维码绑定码" json:"code"`
	Time        time.Time `gorm:"column:time;comment:队伍状态更新时间" json:"time"`
	IsLost      bool      `gorm:"column:is_lost;not null;default:0;comment:是否失联" json:"is_lost"`
	IsWithdrawn bool      `gorm:"column:is_withdrawn;not null;default:0;comment:是否下撤" json:"is_withdrawn"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName Team's table name
func (*Team) TableName() string {
	return TableNameTeam
}
