package model

type Person struct {
	OpenId    string `gorm:"primaryKey"` // openID
	Name      string
	Gender    int8
	StuId     string
	Campus    uint8
	Identity  string // 身份证号
	Status    uint8  // 0 没加入团队，1 加入了并且是队员，2 是队长
	Qq        string
	Wechat    string
	College   string // 学院
	Tel       string
	CreatedOp uint8
	JoinOp    uint8
	TeamId    int `gorm:"index;default:-1"`
}

// encOpenID 是加密后的 openID
// GetPerson 使用加密后的 open ID 获取 person 数据
func GetPerson(encOpenID string) (*Person, error) {
	

	// if x, found := main.Cache.Get(encOpenID); found {
	// 	// 如果缓存中找到了这个数据
	// 	return x.(*Person), nil
	// } else {
	// 	// 如果缓存中没有就进数据库查询

	// }
	return nil, nil
}

// UpdatePerson 更新 person 数据
