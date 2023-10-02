package model

import (
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
	"walk-server/global"
)

type Person struct {
	OpenId     string `gorm:"primaryKey"` // openID
	Name       string
	Gender     int8 // 1 男，2 女
	StuId      string
	Campus     uint8  // 1 朝晖，2 屏峰，3 莫干山
	Identity   string // 身份证号
	Status     uint8  // 0 未加入团队，1 队员，2 队长
	Qq         string
	Wechat     string
	College    string // 学院
	Tel        string
	CreatedOp  uint8 // 创建团队次数
	JoinOp     uint8 // 加入团队次数
	TeamId     int   `gorm:"index;default:-1"`
	WalkStatus uint8 // 1 未开始，2 进行中，3 扫码成功，4 放弃，5 完成
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

// 事务中更新
func TxUpdatePerson(tx *gorm.DB, person *Person) error {
	// 如果缓存中存在这个数据, 先更新缓存
	if _, err := global.Rdb.Get(global.Rctx, person.OpenId).Result(); err == nil {
		global.Rdb.Set(global.Rctx, person.OpenId, person, 20*time.Minute)
	} else if err != redis.Nil {
		return err
	}

	// 更新数据库中的数据
	if err := tx.Where(&Person{OpenId: person.OpenId}).Save(person).Error; err != nil {
		return err
	}

	return nil
}
