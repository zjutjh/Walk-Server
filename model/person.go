package model

type Person struct {
	OpenId    uint `gorm:"primaryKey"`
	Name      string
	Gender    uint8
	StuId     string
	Campus    uint8
	Qq        string
	Tel       string
	CreatedOp uint8
	JoinOp    uint8
	TeamId    string
}
