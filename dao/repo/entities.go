package repo

import "time"

type Person struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	OpenID     string    `gorm:"column:open_id"`
	Name       string    `gorm:"column:name"`
	Gender     int8      `gorm:"column:gender"`
	StuID      string    `gorm:"column:stu_id"`
	Campus     uint8     `gorm:"column:campus"`
	Identity   string    `gorm:"column:identity"`
	Role       uint8     `gorm:"column:role"`
	QQ         string    `gorm:"column:qq"`
	Wechat     string    `gorm:"column:wechat"`
	College    string    `gorm:"column:college"`
	Tel        string    `gorm:"column:tel"`
	CreatedOp  uint8     `gorm:"column:created_op"`
	JoinOp     uint8     `gorm:"column:join_op"`
	TeamID     int64     `gorm:"column:team_id"`
	Type       uint8     `gorm:"column:type"`
	WalkStatus uint8     `gorm:"column:walk_status"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (Person) TableName() string {
	return "people"
}

type Team struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	Name       string     `gorm:"column:name"`
	Num        uint8      `gorm:"column:num"`
	Password   string     `gorm:"column:password"`
	Slogan     string     `gorm:"column:slogan"`
	AllowMatch bool       `gorm:"column:allow_match"`
	Captain    string     `gorm:"column:captain"`
	RouteID    int64      `gorm:"column:route_id"`
	PointID    int8       `gorm:"column:point_id"`
	StartNum   uint32     `gorm:"column:start_num"`
	Submit     bool       `gorm:"column:submit"`
	Status     uint8      `gorm:"column:status"`
	Code       string     `gorm:"column:code"`
	Time       *time.Time `gorm:"column:time"`
	IsLost     bool       `gorm:"column:is_lost"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
}

func (Team) TableName() string {
	return "teams"
}
