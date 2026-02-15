package model

import (
	"time"
)

const TableNamePerson = "people"

// Person mapped from table <people>
type Person struct {
	OpenID     string    `gorm:"column:open_id;primaryKey;comment:微信OpenID" json:"open_id"`
	Name       string    `gorm:"column:name;not null;comment:姓名" json:"name"`
	Gender     int8      `gorm:"column:gender;not null;comment:性别(1男,2女)" json:"gender"`
	StuID      *string   `gorm:"column:stu_id;comment:学号" json:"stu_id"`
	Campus     uint8     `gorm:"column:campus;not null;comment:校区(1朝晖,2屏峰,3莫干山)" json:"campus"`
	Identity   string    `gorm:"column:identity;not null;comment:身份证号" json:"identity"`
	Status     uint8     `gorm:"column:status;not null;default:0;comment:队伍状态(0未加入,1队员,2队长)" json:"status"`
	QQ         *string   `gorm:"column:qq;comment:QQ号" json:"qq"`
	Wechat     *string   `gorm:"column:wechat;comment:微信号" json:"wechat"`
	College    string    `gorm:"column:college;not null;comment:学院" json:"college"`
	Tel        string    `gorm:"column:tel;not null;comment:联系电话" json:"tel"`
	CreatedOp  uint8     `gorm:"column:created_op;not null;default:3;comment:创建团队次数" json:"created_op"`
	JoinOp     uint8     `gorm:"column:join_op;not null;default:5;comment:加入团队次数" json:"join_op"`
	TeamID     *int64    `gorm:"column:team_id;default:-1;comment:所属团队ID" json:"team_id"`
	Type       uint8     `gorm:"column:type;not null;comment:人员类型(1学生,2教职工,3校友)" json:"type"`
	WalkStatus uint8     `gorm:"column:walk_status;not null;default:0;comment:活动状态(0未开始,1待出发,2已放弃,3进行中,4已下撤,5已违规,6已完成)" json:"walk_status"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName Person's table name
func (*Person) TableName() string {
	return TableNamePerson
}
