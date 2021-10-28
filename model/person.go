package model

type Person struct {
	OpenId    string `gorm:"primaryKey"`
	Name      string
	Gender    uint8
	StuId     string
	Campus    uint8
	Qq        string
	Wechat    string
	Tel       string
	CreatedOp uint8
	JoinOp    uint8
	TeamId    string
}
