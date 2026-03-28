package repo

import "time"

type Person struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	OpenID     string    `gorm:"column:open_id"`
	Name       string    `gorm:"column:name"`
	Gender     int8      `gorm:"column:gender"`
	StuID      string    `gorm:"column:stu_id"`
	Campus     string    `gorm:"column:campus"`
	Identity   string    `gorm:"column:identity"`
	Role       string    `gorm:"column:role"`
	QQ         string    `gorm:"column:qq"`
	Wechat     string    `gorm:"column:wechat"`
	College    string    `gorm:"column:college"`
	Tel        string    `gorm:"column:tel"`
	CreatedOp  uint8     `gorm:"column:created_op"`
	JoinOp     uint8     `gorm:"column:join_op"`
	TeamID     int64     `gorm:"column:team_id"`
	Type       string    `gorm:"column:type"`
	WalkStatus string    `gorm:"column:walk_status"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (Person) TableName() string {
	return "peoples"
}

type Team struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	Name          string     `gorm:"column:name"`
	Num           uint8      `gorm:"column:num"`
	Password      string     `gorm:"column:password"`
	Slogan        string     `gorm:"column:slogan"`
	AllowMatch    bool       `gorm:"column:allow_match"`
	Captain       string     `gorm:"column:captain"`
	RouteName     string     `gorm:"column:route_name"`
	PrevPointName string     `gorm:"column:prev_point_name"`
	Submit        bool       `gorm:"column:submit"`
	Status        string     `gorm:"column:status"`
	IsWrongRoute  bool       `gorm:"column:is_wrong_route"`
	IsReunite     bool       `gorm:"column:is_reunite"`
	Code          string     `gorm:"column:code"`
	Time          *time.Time `gorm:"column:time"`
	IsLost        bool       `gorm:"column:is_lost"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (Team) TableName() string {
	return "teams"
}
