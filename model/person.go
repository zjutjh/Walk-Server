package model

import (
	"encoding/json"
	"errors"
	"time"
	"walk-server/global"
)

type Person struct {
	OpenId    string `gorm:"primaryKey"` // openID
	Name      string
	Gender    int8 // 1 男，2 女
	StuId     string
	Campus    uint8  // 1 朝晖，2 屏峰，3 莫干山
	Identity  string // 身份证号
	Status    uint8  // 0 未加入团队，1 队员，2 队长
	Qq        string
	Wechat    string
	College   string // 学院
	Tel       string
	CreatedOp uint8
	JoinOp    uint8
	TeamId    int `gorm:"index;default:-1"`
}

func (p *Person) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

func (p *Person) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}

// encOpenID 是加密后的 openID
// 如果没有找到这个用户就返回 error
// GetPerson 使用加密后的 open ID 获取 person 数据
func GetPerson(encOpenID string) (*Person, error) {
	// 如果缓存中找到了这个数据 直接返回缓存数据
	var person Person
	if err := global.Rdb.Get(global.Rctx, encOpenID).Scan(&person); err == nil {
		return &person, nil
	}

	// 如果缓存中没有就进数据库查询用户数据
	result := global.DB.Where(&Person{OpenId: encOpenID}).Take(&person)
	if result.RowsAffected == 0 {
		return nil, errors.New("no person")
	} else {
		global.Rdb.Set(global.Rctx, encOpenID, &person, 20*time.Minute)
		return &person, nil
	}
}

// encOpenID 加密后的用户 openID
// person 用户数据 (完整的)
// UpdatePerson 更新 person 数据
func UpdatePerson(encOpenID string, person *Person) {
	// 如果缓存中存在这个数据, 先更新缓存
	if _, err := global.Rdb.Get(global.Rctx, encOpenID).Result(); err == nil {
		global.Rdb.Set(global.Rctx, encOpenID, person, 20*time.Minute)
	}

	// 更新数据库中的数据
	global.DB.Where(&Person{OpenId: encOpenID}).Save(person)
}
