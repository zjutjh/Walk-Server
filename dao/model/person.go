package model

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

const TableNamePerson = "person"

// Person mapped from table <person>
type Person struct {
	ID         int64                 `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	OpenId     string                `gorm:"column:open_id;not null" json:"open_id"`
	Name       string                `gorm:"column:name;not null" json:"name"`
	StuId      string                `gorm:"column:stu_id" json:"stu_id"`
	Gender     uint8                 `gorm:"column:gender" json:"gender"`
	Campus     uint8                 `gorm:"column:campus" json:"campus"`
	College    string                `gorm:"column:college" json:"college"`
	Status     uint8                 `gorm:"column:status" json:"status"` // 1: member, 2: captain
	CreatedOp  bool                  `gorm:"column:created_op" json:"created_op"`
	JoinOp     bool                  `gorm:"column:join_op" json:"join_op"`
	TeamId     int64                 `gorm:"column:team_id" json:"team_id"`
	Type       uint8                 `gorm:"column:type" json:"type"` // 1: student, 2: teacher, 3: alumnus
	Qq         string                `gorm:"column:qq" json:"qq"`
	Wechat     string                `gorm:"column:wechat" json:"wechat"`
	Tel        string                `gorm:"column:tel" json:"tel"`
	WalkStatus uint8                 `gorm:"column:walk_status" json:"walk_status"`
	Identity   string                `gorm:"column:identity" json:"identity"` // ID card or similar
	CreatedAt  time.Time             `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP(3)" json:"created_at"`
	UpdatedAt  time.Time             `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP(3)" json:"updated_at"`
	DeletedAt  soft_delete.DeletedAt `gorm:"column:deleted_at;not null;softDelete:milli" json:"-"`
}

// TableName Person's table name
func (*Person) TableName() string {
	return TableNamePerson
}
