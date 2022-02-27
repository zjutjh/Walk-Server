package model

import (
	"errors"
	"walk-server/global"
)

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
	// 如果缓存中找到了这个数据 直接返回缓存数据
	if x, found := global.Cache.Get(encOpenID); found {
		return x.(*Person), nil
	}

	// 如果缓存中没有就进数据库查询用户数据
	person := new(Person)
	result := global.DB.Where("open_id = ?", encOpenID).Take(&person)
	if result.RowsAffected == 0 {
		return nil, errors.New("no person")
	} else {
		global.Cache.SetDefault(encOpenID, person)
		return person, nil
	}
}

// encOpenID 加密后的用户 openID
// person 用户数据
// UpdatePerson 更新 person 数据
func UpdatePerson(encOpenID string, person *Person) {
	// 如果缓存中存在这个数据, 先更新缓存
	if x, found := global.Cache.Get(encOpenID); found {
		cachePerson := x.(*Person)

		cachePerson.Name = person.Name
		cachePerson.Gender = person.Gender
		cachePerson.StuId = person.StuId
		cachePerson.Campus = person.Campus
		cachePerson.College = person.College
		cachePerson.Identity = person.Identity
		cachePerson.Qq = person.Qq
		cachePerson.Wechat = person.Wechat
		cachePerson.Tel = person.Tel

		global.Cache.SetDefault(encOpenID, cachePerson)
	}

	// 更新数据库中的数据
	global.DB.Model(&person).Where("open_id = ?", encOpenID).Updates(*person)
}
